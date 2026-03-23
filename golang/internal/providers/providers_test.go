package providers_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrsimonemms/document-processing-demo/golang/internal/models"
	"github.com/mrsimonemms/document-processing-demo/golang/internal/providers"
)

var testReq = providers.SummariseRequest{Chunks: []string{"the quick brown fox"}}

func TestSummariseWithFailover_HappyPath(t *testing.T) {
	chain := providers.DefaultChain()

	result, err := providers.SummariseWithFailover(context.Background(), chain, testReq)

	require.NoError(t, err)
	assert.Equal(t, models.ProviderOpenAI, result.Provider)
	assert.False(t, result.FallbackOccurred)
	assert.Contains(t, result.Summary, "[openai]")
}

func TestSummariseWithFailover_FallbackOnPrimaryFailure(t *testing.T) {
	chain := providers.DefaultChain()
	chain[0] = providers.NewFaultyProvider(chain[0], "primary exploded")

	result, err := providers.SummariseWithFailover(context.Background(), chain, testReq)

	require.NoError(t, err)
	assert.Equal(t, models.ProviderAnthropic, result.Provider)
	assert.True(t, result.FallbackOccurred)
	assert.Contains(t, result.Summary, "[anthropic]")
}

func TestSummariseWithFailover_AllProvidersFail(t *testing.T) {
	chain := providers.DefaultChain()
	chain[0] = providers.NewFaultyProvider(chain[0], "primary failed")
	chain[1] = providers.NewFaultyProvider(chain[1], "fallback failed")

	_, err := providers.SummariseWithFailover(context.Background(), chain, testReq)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "all providers failed")
}

func TestSummariseWithFailover_EmptyChain(t *testing.T) {
	_, err := providers.SummariseWithFailover(context.Background(), nil, testReq)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "all providers failed")
}

func TestFaultyProvider_AlwaysFails(t *testing.T) {
	base := providers.DefaultChain()[0]
	faulty := providers.NewFaultyProvider(base, "injected failure")

	assert.Equal(t, base.Name(), faulty.Name())

	_, err := faulty.Summarise(context.Background(), testReq)
	require.Error(t, err)
	assert.Equal(t, "injected failure", err.Error())
}

func TestOpenAI_Summarise(t *testing.T) {
	chain := providers.DefaultChain()
	openai := chain[0]

	resp, err := openai.Summarise(context.Background(), testReq)

	require.NoError(t, err)
	assert.Contains(t, resp.Summary, "[openai]")
	assert.Contains(t, resp.Summary, "1 chunk(s)")
}

func TestAnthropic_Summarise(t *testing.T) {
	chain := providers.DefaultChain()
	anthropicProvider := chain[1]

	resp, err := anthropicProvider.Summarise(context.Background(), testReq)

	require.NoError(t, err)
	assert.Contains(t, resp.Summary, "[anthropic]")
	assert.Contains(t, resp.Summary, "1 chunk(s)")
}

var testAnswerReq = providers.AnswerRequest{
	Content:  "the quick brown fox",
	Question: "what did the fox do?",
}

func TestAnswerWithFailover_HappyPath(t *testing.T) {
	chain := providers.DefaultQuestionChain()

	result, err := providers.AnswerWithFailover(context.Background(), chain, testAnswerReq)

	require.NoError(t, err)
	assert.Equal(t, models.ProviderOpenAI, result.Provider)
	assert.False(t, result.FallbackOccurred)
	assert.Contains(t, result.Answer, "[openai]")
}

func TestAnswerWithFailover_FallbackOnPrimaryFailure(t *testing.T) {
	chain := providers.DefaultQuestionChain()
	chain[0] = providers.NewFaultyQuestionProvider(chain[0], "primary exploded")

	result, err := providers.AnswerWithFailover(context.Background(), chain, testAnswerReq)

	require.NoError(t, err)
	assert.Equal(t, models.ProviderAnthropic, result.Provider)
	assert.True(t, result.FallbackOccurred)
	assert.Contains(t, result.Answer, "[anthropic]")
}

func TestAnswerWithFailover_AllProvidersFail(t *testing.T) {
	chain := providers.DefaultQuestionChain()
	chain[0] = providers.NewFaultyQuestionProvider(chain[0], "primary failed")
	chain[1] = providers.NewFaultyQuestionProvider(chain[1], "fallback failed")

	_, err := providers.AnswerWithFailover(context.Background(), chain, testAnswerReq)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "all providers failed")
}

func TestFaultyQuestionProvider_AlwaysFails(t *testing.T) {
	base := providers.DefaultQuestionChain()[0]
	faulty := providers.NewFaultyQuestionProvider(base, "injected failure")

	assert.Equal(t, base.Name(), faulty.Name())

	_, err := faulty.Answer(context.Background(), testAnswerReq)
	require.Error(t, err)
	assert.Equal(t, "injected failure", err.Error())
}
