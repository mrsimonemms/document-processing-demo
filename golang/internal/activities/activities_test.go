package activities_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"

	"github.com/mrsimonemms/document-processing-demo/golang/internal/activities"
	"github.com/mrsimonemms/document-processing-demo/golang/internal/models"
	"github.com/mrsimonemms/document-processing-demo/golang/internal/providers"
)

// fakePrimary implements both Summariser and QuestionAnswerer.
// It returns deterministic output labelled "[openai]" so that existing
// assertions remain meaningful without calling a real API.
type fakePrimary struct{}

func (f *fakePrimary) Name() models.ProviderName { return models.ProviderOpenAI }

func (f *fakePrimary) Summarise(_ context.Context, req providers.SummariseRequest) (providers.SummariseResponse, error) {
	return providers.SummariseResponse{
		Summary: fmt.Sprintf("[openai] %d chunk(s)", len(req.Chunks)),
	}, nil
}

func (f *fakePrimary) Answer(_ context.Context, req providers.AnswerRequest) (providers.AnswerResponse, error) {
	return providers.AnswerResponse{
		Answer: fmt.Sprintf("[openai] answer to: %s", req.Question),
	}, nil
}

// fakeFallback implements both Summariser and QuestionAnswerer for the Anthropic
// fallback slot. It returns deterministic output without calling a real API.
type fakeFallback struct{}

func (f *fakeFallback) Name() models.ProviderName { return models.ProviderAnthropic }

func (f *fakeFallback) Summarise(_ context.Context, req providers.SummariseRequest) (providers.SummariseResponse, error) {
	return providers.SummariseResponse{
		Summary: fmt.Sprintf("[anthropic] %d chunk(s)", len(req.Chunks)),
	}, nil
}

func (f *fakeFallback) Answer(_ context.Context, req providers.AnswerRequest) (providers.AnswerResponse, error) {
	return providers.AnswerResponse{
		Answer: fmt.Sprintf("[anthropic] answer to: %s", req.Question),
	}, nil
}

func newActivityEnv(t *testing.T) (*testsuite.TestActivityEnvironment, *activities.Activities) {
	t.Helper()

	ts := &testsuite.WorkflowTestSuite{}
	env := ts.NewTestActivityEnvironment()

	primary := &fakePrimary{}
	fallback := &fakeFallback{}

	acts := activities.NewActivities(
		[]providers.Summariser{primary, fallback},
		[]providers.QuestionAnswerer{primary, fallback},
	)

	env.RegisterActivity(acts)

	return env, acts
}

// TestSummariseDocumentActivity_HappyPath verifies that the primary provider is
// selected by default and that no fallback occurs.
func TestSummariseDocumentActivity_HappyPath(t *testing.T) {
	env, acts := newActivityEnv(t)

	input := models.SummariseInput{
		Chunks:   []string{"the quick brown fox"},
		Scenario: models.ScenarioHappyPath,
	}

	val, err := env.ExecuteActivity(acts.SummariseDocumentActivity, input)
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
	env, acts := newActivityEnv(t)

	input := models.SummariseInput{
		Chunks:   []string{"the quick brown fox"},
		Scenario: models.ScenarioProviderFailover,
	}

	val, err := env.ExecuteActivity(acts.SummariseDocumentActivity, input)
	require.NoError(t, err)

	var result models.SummariseResult
	require.NoError(t, val.Get(&result))

	assert.Equal(t, models.ProviderAnthropic, result.Provider)
	assert.True(t, result.FallbackOccurred)
	assert.Contains(t, result.Summary, "[anthropic]")
}

func TestSummariseDocumentActivity_ProviderDown(t *testing.T) {
	env, acts := newActivityEnv(t)

	input := models.SummariseInput{
		Chunks:   []string{"the quick brown fox"},
		Scenario: models.ScenarioProviderDown,
	}

	val, err := env.ExecuteActivity(acts.SummariseDocumentActivity, input)
	require.NoError(t, err)

	var result models.SummariseResult
	require.NoError(t, val.Get(&result))

	assert.Equal(t, models.ProviderAnthropic, result.Provider)
	assert.True(t, result.FallbackOccurred)
	assert.Contains(t, result.Summary, "[anthropic]")
}

func TestSummariseDocumentActivity_ProviderRateLimit(t *testing.T) {
	env, acts := newActivityEnv(t)

	input := models.SummariseInput{
		Chunks:   []string{"the quick brown fox"},
		Scenario: models.ScenarioProviderRateLimit,
	}

	val, err := env.ExecuteActivity(acts.SummariseDocumentActivity, input)
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
	env, acts := newActivityEnv(t)

	input := models.SummariseInput{
		Chunks:   []string{"the quick brown fox"},
		Scenario: models.ScenarioFailOnceSummarise,
	}

	// The test environment runs with attempt=1 by default.
	// The activity must return an error on this attempt.
	_, err := env.ExecuteActivity(acts.SummariseDocumentActivity, input)
	require.Error(t, err)
}

// TestExtractDocumentTextActivity passes content through unchanged.
func TestExtractDocumentTextActivity(t *testing.T) {
	env, _ := newActivityEnv(t)

	input := models.DocumentInput{
		DocumentID: "doc-1",
		Content:    "hello world",
		Scenario:   models.ScenarioHappyPath,
	}

	val, err := env.ExecuteActivity(activities.ExtractDocumentTextActivityName, input)
	require.NoError(t, err)

	var text string
	require.NoError(t, val.Get(&text))
	assert.Equal(t, "hello world", text)
}

// TestChunkDocumentActivity verifies that content is split correctly.
func TestChunkDocumentActivity(t *testing.T) {
	env, _ := newActivityEnv(t)

	val, err := env.ExecuteActivity(activities.ChunkDocumentActivityName, "one two three")
	require.NoError(t, err)

	var chunks []string
	require.NoError(t, val.Get(&chunks))
	require.Len(t, chunks, 1)
	assert.Equal(t, "one two three", chunks[0])
}

// TestAnswerQuestionActivity_HappyPath verifies that the primary provider is
// selected by default and that no fallback occurs.
func TestAnswerQuestionActivity_HappyPath(t *testing.T) {
	env, acts := newActivityEnv(t)

	input := models.AnswerInput{
		Content:  "the quick brown fox",
		Question: "what did the fox do?",
		Scenario: models.ScenarioHappyPath,
	}

	val, err := env.ExecuteActivity(acts.AnswerQuestionActivity, input)
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
	env, acts := newActivityEnv(t)

	input := models.AnswerInput{
		Content:  "the quick brown fox",
		Question: "what did the fox do?",
		Scenario: models.ScenarioProviderFailover,
	}

	val, err := env.ExecuteActivity(acts.AnswerQuestionActivity, input)
	require.NoError(t, err)

	var result models.AnswerResult
	require.NoError(t, val.Get(&result))

	assert.Equal(t, models.ProviderAnthropic, result.Provider)
	assert.True(t, result.FallbackOccurred)
	assert.Contains(t, result.Answer, "[anthropic]")
}

func TestAnswerQuestionActivity_ProviderDown(t *testing.T) {
	env, acts := newActivityEnv(t)

	input := models.AnswerInput{
		Content:  "the quick brown fox",
		Question: "what did the fox do?",
		Scenario: models.ScenarioProviderDown,
	}

	val, err := env.ExecuteActivity(acts.AnswerQuestionActivity, input)
	require.NoError(t, err)

	var result models.AnswerResult
	require.NoError(t, val.Get(&result))

	assert.Equal(t, models.ProviderAnthropic, result.Provider)
	assert.True(t, result.FallbackOccurred)
	assert.Contains(t, result.Answer, "[anthropic]")
}

func TestAnswerQuestionActivity_ProviderRateLimit(t *testing.T) {
	env, acts := newActivityEnv(t)

	input := models.AnswerInput{
		Content:  "the quick brown fox",
		Question: "what did the fox do?",
		Scenario: models.ScenarioProviderRateLimit,
	}

	val, err := env.ExecuteActivity(acts.AnswerQuestionActivity, input)
	require.NoError(t, err)

	var result models.AnswerResult
	require.NoError(t, val.Get(&result))

	assert.Equal(t, models.ProviderAnthropic, result.Provider)
	assert.True(t, result.FallbackOccurred)
	assert.Contains(t, result.Answer, "[anthropic]")
}
