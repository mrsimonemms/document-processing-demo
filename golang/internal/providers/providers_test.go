package providers_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrsimonemms/document-processing-demo/golang/internal/models"
	"github.com/mrsimonemms/document-processing-demo/golang/internal/providers"
)

// fakeProvider is a minimal Summariser and QuestionAnswerer for tests.
// It returns deterministic output and never calls a real API.
type fakeProvider struct {
	name   models.ProviderName
	output string
}

func (f *fakeProvider) Name() models.ProviderName { return f.name }

func (f *fakeProvider) Summarise(_ context.Context, _ providers.SummariseRequest) (providers.SummariseResponse, error) {
	return providers.SummariseResponse{Summary: f.output}, nil
}

func (f *fakeProvider) Answer(_ context.Context, _ providers.AnswerRequest) (providers.AnswerResponse, error) {
	return providers.AnswerResponse{Answer: f.output}, nil
}

// testChain builds a two-provider summarise chain using fakeProvider instances.
func testChain() []providers.Summariser {
	return []providers.Summariser{
		&fakeProvider{name: models.ProviderOpenAI, output: "[openai] fake summary"},
		&fakeProvider{name: models.ProviderAnthropic, output: "[anthropic] fake summary"},
	}
}

// testQuestionChain builds a two-provider Q&A chain using fakeProvider instances.
func testQuestionChain() []providers.QuestionAnswerer {
	return []providers.QuestionAnswerer{
		&fakeProvider{name: models.ProviderOpenAI, output: "[openai] fake answer"},
		&fakeProvider{name: models.ProviderAnthropic, output: "[anthropic] fake answer"},
	}
}

var testReq = providers.SummariseRequest{Chunks: []string{"the quick brown fox"}}

func TestSummariseWithFailover_HappyPath(t *testing.T) {
	chain := testChain()

	result, err := providers.SummariseWithFailover(context.Background(), chain, testReq)

	require.NoError(t, err)
	assert.Equal(t, models.ProviderOpenAI, result.Provider)
	assert.False(t, result.FallbackOccurred)
	assert.Contains(t, result.Summary, "[openai]")
}

func TestSummariseWithFailover_FallbackOnPrimaryFailure(t *testing.T) {
	chain := testChain()
	chain[0] = providers.NewFaultyProvider(chain[0], "primary exploded")

	result, err := providers.SummariseWithFailover(context.Background(), chain, testReq)

	require.NoError(t, err)
	assert.Equal(t, models.ProviderAnthropic, result.Provider)
	assert.True(t, result.FallbackOccurred)
	assert.Contains(t, result.Summary, "[anthropic]")
}

func TestSummariseWithFailover_AllProvidersFail(t *testing.T) {
	chain := testChain()
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
	base := &fakeProvider{name: models.ProviderOpenAI, output: "ok"}
	faulty := providers.NewFaultyProvider(base, "injected failure")

	assert.Equal(t, base.Name(), faulty.Name())

	_, err := faulty.Summarise(context.Background(), testReq)
	require.Error(t, err)
	assert.Equal(t, "injected failure", err.Error())
}

func TestAnthropic_Summarise(t *testing.T) {
	p := &providers.Anthropic{}

	resp, err := p.Summarise(context.Background(), testReq)

	require.NoError(t, err)
	assert.Contains(t, resp.Summary, "[anthropic]")
	assert.Contains(t, resp.Summary, "1 chunk(s)")
}

func TestAnthropic_Answer(t *testing.T) {
	p := &providers.Anthropic{}
	req := providers.AnswerRequest{
		Content:  "the quick brown fox",
		Question: "what did the fox do?",
	}

	resp, err := p.Answer(context.Background(), req)

	require.NoError(t, err)
	assert.Contains(t, resp.Answer, "[anthropic]")
}

var testAnswerReq = providers.AnswerRequest{
	Content:  "the quick brown fox",
	Question: "what did the fox do?",
}

func TestAnswerWithFailover_HappyPath(t *testing.T) {
	chain := testQuestionChain()

	result, err := providers.AnswerWithFailover(context.Background(), chain, testAnswerReq)

	require.NoError(t, err)
	assert.Equal(t, models.ProviderOpenAI, result.Provider)
	assert.False(t, result.FallbackOccurred)
	assert.Contains(t, result.Answer, "[openai]")
}

func TestAnswerWithFailover_FallbackOnPrimaryFailure(t *testing.T) {
	chain := testQuestionChain()
	chain[0] = providers.NewFaultyQuestionProvider(chain[0], "primary exploded")

	result, err := providers.AnswerWithFailover(context.Background(), chain, testAnswerReq)

	require.NoError(t, err)
	assert.Equal(t, models.ProviderAnthropic, result.Provider)
	assert.True(t, result.FallbackOccurred)
	assert.Contains(t, result.Answer, "[anthropic]")
}

func TestAnswerWithFailover_AllProvidersFail(t *testing.T) {
	chain := testQuestionChain()
	chain[0] = providers.NewFaultyQuestionProvider(chain[0], "primary failed")
	chain[1] = providers.NewFaultyQuestionProvider(chain[1], "fallback failed")

	_, err := providers.AnswerWithFailover(context.Background(), chain, testAnswerReq)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "all providers failed")
}

func TestFaultyQuestionProvider_AlwaysFails(t *testing.T) {
	base := &fakeProvider{name: models.ProviderOpenAI, output: "ok"}
	faulty := providers.NewFaultyQuestionProvider(base, "injected failure")

	assert.Equal(t, base.Name(), faulty.Name())

	_, err := faulty.Answer(context.Background(), testAnswerReq)
	require.Error(t, err)
	assert.Equal(t, "injected failure", err.Error())
}

func TestNewChain_OrderPreserved(t *testing.T) {
	primary := &fakeProvider{name: models.ProviderOpenAI, output: "primary"}
	fallback := &fakeProvider{name: models.ProviderAnthropic, output: "fallback"}

	chain := providers.NewChain(primary, fallback)

	require.Len(t, chain, 2)
	assert.Equal(t, models.ProviderOpenAI, chain[0].Name())
	assert.Equal(t, models.ProviderAnthropic, chain[1].Name())
}

func TestNewQuestionChain_OrderPreserved(t *testing.T) {
	primary := &fakeProvider{name: models.ProviderOpenAI, output: "primary"}
	fallback := &fakeProvider{name: models.ProviderAnthropic, output: "fallback"}

	chain := providers.NewQuestionChain(primary, fallback)

	require.Len(t, chain, 2)
	assert.Equal(t, models.ProviderOpenAI, chain[0].Name())
	assert.Equal(t, models.ProviderAnthropic, chain[1].Name())
}

// TestFaultyProvider_WrapsNameCorrectly verifies that a FaultyProvider wrapping
// a provider with a known name preserves that name, which is important for demo
// logs showing which provider was injected with a failure.
func TestFaultyProvider_WrapsNameCorrectly(t *testing.T) {
	base := &fakeProvider{name: models.ProviderOpenAI, output: "ok"}
	faulty := providers.NewFaultyProvider(base, "msg")

	assert.Equal(t, models.ProviderOpenAI, faulty.Name())

	_, err := faulty.Summarise(context.Background(), testReq)
	require.Error(t, err)
	assert.True(t, errors.Is(err, errors.New("msg")) || err.Error() == "msg")
}
