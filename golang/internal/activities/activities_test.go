package activities_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"

	"github.com/mrsimonemms/document-processing-demo/golang/internal/activities"
	"github.com/mrsimonemms/document-processing-demo/golang/internal/models"
)

func newActivityEnv(t *testing.T) *testsuite.TestActivityEnvironment {
	t.Helper()

	ts := &testsuite.WorkflowTestSuite{}
	env := ts.NewTestActivityEnvironment()
	env.RegisterActivity(activities.ExtractDocumentTextActivity)
	env.RegisterActivity(activities.ChunkDocumentActivity)
	env.RegisterActivity(activities.SummariseDocumentActivity)
	env.RegisterActivity(activities.AnswerQuestionActivity)

	return env
}

// TestSummariseDocumentActivity_HappyPath verifies that the openai provider is
// selected by default and that no fallback occurs.
func TestSummariseDocumentActivity_HappyPath(t *testing.T) {
	env := newActivityEnv(t)

	input := models.SummariseInput{
		Chunks:   []string{"the quick brown fox"},
		Scenario: models.ScenarioHappyPath,
	}

	val, err := env.ExecuteActivity(activities.SummariseDocumentActivity, input)
	require.NoError(t, err)

	var result models.SummariseResult
	require.NoError(t, val.Get(&result))

	assert.Equal(t, models.ProviderOpenAI, result.Provider)
	assert.False(t, result.FallbackOccurred)
	assert.Contains(t, result.Summary, "[openai]")
}

// TestSummariseDocumentActivity_ProviderFailover verifies that the primary
// provider is replaced by a FaultyProvider, the failover mechanism moves to
// the fallback, and FallbackOccurred is set to true.
func TestSummariseDocumentActivity_ProviderFailover(t *testing.T) {
	env := newActivityEnv(t)

	input := models.SummariseInput{
		Chunks:   []string{"the quick brown fox"},
		Scenario: models.ScenarioProviderFailover,
	}

	val, err := env.ExecuteActivity(activities.SummariseDocumentActivity, input)
	require.NoError(t, err)

	var result models.SummariseResult
	require.NoError(t, val.Get(&result))

	assert.Equal(t, models.ProviderAnthropic, result.Provider)
	assert.True(t, result.FallbackOccurred)
	assert.Contains(t, result.Summary, "[anthropic]")
}

// TestSummariseDocumentActivity_FailOnceSummarise_FirstAttempt verifies that
// the activity returns an error on attempt 1 under the fail_once_summarise
// scenario. Temporal would retry this automatically in a real workflow run.
func TestSummariseDocumentActivity_FailOnceSummarise_FirstAttempt(t *testing.T) {
	env := newActivityEnv(t)

	input := models.SummariseInput{
		Chunks:   []string{"the quick brown fox"},
		Scenario: models.ScenarioFailOnceSummarise,
	}

	// The test environment runs with attempt=1 by default.
	// The activity must return an error on this attempt.
	_, err := env.ExecuteActivity(activities.SummariseDocumentActivity, input)
	require.Error(t, err)
}

// TestExtractDocumentTextActivity passes content through unchanged.
func TestExtractDocumentTextActivity(t *testing.T) {
	env := newActivityEnv(t)

	input := models.DocumentInput{
		DocumentID: "doc-1",
		Content:    "hello world",
		Scenario:   models.ScenarioHappyPath,
	}

	val, err := env.ExecuteActivity(activities.ExtractDocumentTextActivity, input)
	require.NoError(t, err)

	var text string
	require.NoError(t, val.Get(&text))
	assert.Equal(t, "hello world", text)
}

// TestChunkDocumentActivity verifies that content is split correctly.
func TestChunkDocumentActivity(t *testing.T) {
	env := newActivityEnv(t)

	val, err := env.ExecuteActivity(activities.ChunkDocumentActivity, "one two three")
	require.NoError(t, err)

	var chunks []string
	require.NoError(t, val.Get(&chunks))
	require.Len(t, chunks, 1)
	assert.Equal(t, "one two three", chunks[0])
}

// TestAnswerQuestionActivity_HappyPath verifies that the openai provider is
// selected by default and that no fallback occurs.
func TestAnswerQuestionActivity_HappyPath(t *testing.T) {
	env := newActivityEnv(t)

	input := models.AnswerInput{
		Content:  "the quick brown fox",
		Question: "what did the fox do?",
		Scenario: models.ScenarioHappyPath,
	}

	val, err := env.ExecuteActivity(activities.AnswerQuestionActivity, input)
	require.NoError(t, err)

	var result models.AnswerResult
	require.NoError(t, val.Get(&result))

	assert.Equal(t, models.ProviderOpenAI, result.Provider)
	assert.False(t, result.FallbackOccurred)
	assert.Contains(t, result.Answer, "[openai]")
}

// TestAnswerQuestionActivity_ProviderFailover verifies that the primary
// provider is replaced by a FaultyQuestionProvider under the failover scenario,
// causing the fallback provider to be used.
func TestAnswerQuestionActivity_ProviderFailover(t *testing.T) {
	env := newActivityEnv(t)

	input := models.AnswerInput{
		Content:  "the quick brown fox",
		Question: "what did the fox do?",
		Scenario: models.ScenarioProviderFailover,
	}

	val, err := env.ExecuteActivity(activities.AnswerQuestionActivity, input)
	require.NoError(t, err)

	var result models.AnswerResult
	require.NoError(t, val.Get(&result))

	assert.Equal(t, models.ProviderAnthropic, result.Provider)
	assert.True(t, result.FallbackOccurred)
	assert.Contains(t, result.Answer, "[anthropic]")
}
