package workflows_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/testsuite"

	"github.com/mrsimonemms/document-processing-demo/golang/internal/activities"
	"github.com/mrsimonemms/document-processing-demo/golang/internal/models"
	"github.com/mrsimonemms/document-processing-demo/golang/internal/providers"
	"github.com/mrsimonemms/document-processing-demo/golang/internal/workflows"
)

func newWorkflowEnv() *testsuite.TestWorkflowEnvironment {
	suite := &testsuite.WorkflowTestSuite{}
	env := suite.NewTestWorkflowEnvironment()
	env.RegisterWorkflow(workflows.DocumentWorkflow)
	env.RegisterActivity(activities.NewActivities(nil, nil))
	return env
}

func TestDocumentWorkflow_SummaryProviderFailureSetsSummaryFailedState(t *testing.T) {
	env := newWorkflowEnv()
	input := models.DocumentInput{
		DocumentID:       "doc-1",
		Content:          "hello world",
		Scenario:         models.ScenarioProviderDown,
		ProviderOverride: models.ProviderOverrideDefault,
	}

	env.OnActivity(activities.ExtractDocumentTextActivityName, mock.Anything, input).Return("hello world", nil)
	env.OnActivity(activities.ChunkDocumentActivityName, mock.Anything, "hello world").Return([]string{"hello world"}, nil)
	env.OnActivity(activities.SummariseDocumentActivityName, mock.Anything, models.SummariseInput{
		Chunks:           []string{"hello world"},
		Scenario:         models.ScenarioProviderDown,
		ProviderOverride: models.ProviderOverrideDefault,
	}).Return(models.SummariseResult{}, temporal.NewNonRetryableApplicationError(
		"all providers failed: openai attempt 3/3: simulated openai provider down",
		"ProviderError",
		providers.NewProviderError(providers.FailureFatal, "all providers failed: openai attempt 3/3: simulated openai provider down"),
	))

	env.RegisterDelayedCallback(func() {
		val, err := env.QueryWorkflow("getState")
		require.NoError(t, err)

		var state models.DocumentState
		require.NoError(t, val.Get(&state))
		assert.Equal(t, "summary_failed", state.Phase)
		assert.Contains(t, state.SummaryError, "all providers failed: openai attempt 3/3: simulated openai provider down")
		assert.Empty(t, state.Summary)
	}, time.Second)

	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow("end", nil)
	}, 2*time.Second)

	env.ExecuteWorkflow(workflows.DocumentWorkflow, input)

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
}

func TestDocumentWorkflow_QuestionFailureSetsLastQuestionErrorWithoutAppendingHistory(t *testing.T) {
	env := newWorkflowEnv()
	input := models.DocumentInput{
		DocumentID:       "doc-2",
		Content:          "hello world",
		Scenario:         models.ScenarioHappyPath,
		ProviderOverride: models.ProviderOverrideDefault,
	}

	env.OnActivity(activities.ExtractDocumentTextActivityName, mock.Anything, input).Return("hello world", nil)
	env.OnActivity(activities.ChunkDocumentActivityName, mock.Anything, "hello world").Return([]string{"hello world"}, nil)
	env.OnActivity(activities.SummariseDocumentActivityName, mock.Anything, models.SummariseInput{
		Chunks:           []string{"hello world"},
		Scenario:         models.ScenarioHappyPath,
		ProviderOverride: models.ProviderOverrideDefault,
	}).Return(models.SummariseResult{
		Summary:          "summary",
		Provider:         models.ProviderOpenAI,
		Model:            "gpt-4o",
		FallbackOccurred: false,
	}, nil)
	env.OnActivity(activities.AnswerQuestionActivityName, mock.Anything, mock.MatchedBy(func(input models.AnswerInput) bool {
		return input.Content == "hello world" &&
			input.Question == "What is this?" &&
			input.Scenario == models.ScenarioProviderRateLimit &&
			input.ProviderOverride == models.ProviderOverrideDefault &&
			len(input.History) == 0
	})).Return(models.AnswerResult{}, temporal.NewNonRetryableApplicationError(
		"all providers failed: openai attempt 1/3: simulated openai provider rate limit",
		"ProviderError",
		providers.NewProviderError(providers.FailureFatal, "all providers failed: openai attempt 1/3: simulated openai provider rate limit"),
	))

	env.RegisterDelayedCallback(func() {
		env.UpdateWorkflow("askQuestion", "", &testsuite.TestUpdateCallback{
			OnReject: func(err error) {
				require.Fail(t, "update should be accepted", err)
			},
			OnAccept: func() {},
			OnComplete: func(_ interface{}, err error) {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "all providers failed")
			},
		}, models.QuestionUpdate{
			Question:         "What is this?",
			Scenario:         models.ScenarioProviderRateLimit,
			ProviderOverride: models.ProviderOverrideDefault,
		})
	}, time.Second)

	env.RegisterDelayedCallback(func() {
		val, err := env.QueryWorkflow("getState")
		require.NoError(t, err)

		var state models.DocumentState
		require.NoError(t, val.Get(&state))
		assert.Equal(t, "summarised", state.Phase)
		assert.Contains(t, state.LastQuestionError, "all providers failed: openai attempt 1/3: simulated openai provider rate limit")
		assert.Len(t, state.QA, 0)

		env.SignalWorkflow("end", nil)
	}, 2*time.Second)

	env.ExecuteWorkflow(workflows.DocumentWorkflow, input)

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
}
