package activities

import (
	"context"
	"fmt"
	"strings"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/temporal"

	"github.com/mrsimonemms/document-processing-demo/golang/internal/models"
	"github.com/mrsimonemms/document-processing-demo/golang/internal/providers"
)

const wordsPerChunk = 100

// Activity type names used when scheduling activities from a workflow.
// These match the method names on Activities as registered with Temporal.
const (
	AnswerQuestionActivityName      = "AnswerQuestionActivity"
	ChunkDocumentActivityName       = "ChunkDocumentActivity"
	ExtractDocumentTextActivityName = "ExtractDocumentTextActivity"
	SummariseDocumentActivityName   = "SummariseDocumentActivity"
)

// Activities holds the provider chains injected at worker startup.
// Construct one via NewActivities and register its methods as Temporal activities.
type Activities struct {
	summarisers []providers.Summariser
	questioners []providers.QuestionAnswerer
}

// NewActivities creates an Activities value with the given provider chains.
// The chains are used by SummariseDocumentActivity and AnswerQuestionActivity.
// Both chains are copied so that scenario injection never mutates the originals.
func NewActivities(summarisers []providers.Summariser, questioners []providers.QuestionAnswerer) *Activities {
	return &Activities{
		summarisers: summarisers,
		questioners: questioners,
	}
}

// ChunkDocumentActivity splits the extracted text into fixed-size word chunks.
//
// Chunking is deterministic: given the same input it always produces the same
// output. No randomness or external calls.
func (a *Activities) ChunkDocumentActivity(_ context.Context, text string) ([]string, error) {
	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{text}, nil
	}

	var chunks []string
	for i := 0; i < len(words); i += wordsPerChunk {
		end := min(i+wordsPerChunk, len(words))
		chunks = append(chunks, strings.Join(words[i:end], " "))
	}

	return chunks, nil
}

// ExtractDocumentTextActivity returns the document content unchanged.
//
// In a production system this would parse a binary format, call an OCR
// service, or decode a base64-encoded payload. Here it is intentionally
// trivial so the focus stays on orchestration rather than parsing logic.
func (a *Activities) ExtractDocumentTextActivity(_ context.Context, input models.DocumentInput) (string, error) {
	return input.Content, nil
}

// selectSummarisers returns the subset of chain to use based on the provider
// override. Default (or empty) returns the full chain. A named override filters
// to only the matching provider; if no match is found the full chain is used.
func selectSummarisers(chain []providers.Summariser, override models.ProviderOverride) []providers.Summariser {
	if override == models.ProviderOverrideDefault || override == "" {
		return chain
	}
	target := models.ProviderName(override)
	var selected []providers.Summariser
	for _, p := range chain {
		if p.Name() == target {
			selected = append(selected, p)
		}
	}
	if len(selected) == 0 {
		return chain
	}
	return selected
}

// selectQuestioners returns the subset of chain to use based on the provider
// override. Same rules as selectSummarisers.
func selectQuestioners(chain []providers.QuestionAnswerer, override models.ProviderOverride) []providers.QuestionAnswerer {
	if override == models.ProviderOverrideDefault || override == "" {
		return chain
	}
	target := models.ProviderName(override)
	var selected []providers.QuestionAnswerer
	for _, p := range chain {
		if p.Name() == target {
			selected = append(selected, p)
		}
	}
	if len(selected) == 0 {
		return chain
	}
	return selected
}

// SummariseDocumentActivity produces a summary from the chunks via the
// provider chain and returns metadata about which provider was used.
//
// Failure injection is driven by input.Scenario:
//
//   - fail_once_summarise: the activity fails on attempt 1 with an
//     ApplicationError. Temporal retries it automatically. No provider is
//     called on the failing attempt.
//
//   - primary_provider_failure: the primary provider in the chain is replaced
//     with a FaultyProvider shim. SummariseWithFailover tries the primary
//     (fails), then uses the fallback. FallbackOccurred is set to true.
//
//   - provider_down: the primary provider is replaced with a retryable shim.
//     SummariseWithFailover retries the primary up to 3 times before moving on
//     to the fallback provider.
//
//   - provider_rate_limit: the primary provider is replaced with a
//     non-retryable rate-limit shim. SummariseWithFailover skips retries and
//     fails over immediately to the next provider.
//
// Provider selection is controlled by input.ProviderOverride before scenario
// injection is applied. All failure injection is deterministic and reads from
// input.Scenario only. No global state or randomness is involved.
func (a *Activities) SummariseDocumentActivity(ctx context.Context, input models.SummariseInput) (models.SummariseResult, error) {
	// Activity-level retry injection: fail the whole activity on attempt 1.
	// Temporal retries the activity automatically; the second attempt succeeds.
	if input.Scenario == models.ScenarioFailOnceSummarise {
		info := activity.GetInfo(ctx)
		if info.Attempt == 1 {
			return models.SummariseResult{}, temporal.NewApplicationError(
				"simulated summarise failure on attempt 1",
				"SimulatedFailure",
			)
		}
	}

	// Select providers then copy so scenario injection does not mutate the base slice.
	base := selectSummarisers(a.summarisers, input.ProviderOverride)
	chain := make([]providers.Summariser, len(base))
	copy(chain, base)

	// Provider-level failure injection: replace the primary with a failing shim.
	// The failover loop tries primary (fails) then falls through to the fallback.
	if input.Scenario == models.ScenarioProviderFailover && len(chain) > 0 {
		chain[0] = providers.NewFaultyProvider(chain[0], "simulated primary provider failure")
	}

	if input.Scenario == models.ScenarioProviderDown && len(chain) > 0 {
		chain[0] = providers.NewDownProvider(chain[0], fmt.Sprintf("simulated %s provider down", chain[0].Name()))
	}

	if input.Scenario == models.ScenarioProviderRateLimit && len(chain) > 0 {
		chain[0] = providers.NewRateLimitProvider(chain[0], fmt.Sprintf("simulated %s provider rate limit", chain[0].Name()))
	}

	result, err := providers.SummariseWithFailover(ctx, chain, providers.SummariseRequest{Chunks: input.Chunks})
	if err != nil {
		if providers.IsProviderError(err) {
			return models.SummariseResult{}, temporal.NewNonRetryableApplicationError(err.Error(), "ProviderError", err)
		}
		return models.SummariseResult{}, err
	}

	return result, nil
}
