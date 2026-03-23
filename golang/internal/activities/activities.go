package activities

import (
	"context"
	"strings"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/temporal"

	"github.com/mrsimonemms/document-processing-demo/golang/internal/models"
	"github.com/mrsimonemms/document-processing-demo/golang/internal/providers"
)

const wordsPerChunk = 100

// ExtractDocumentTextActivity returns the document content unchanged.
//
// In a production system this would parse a binary format, call an OCR
// service, or decode a base64-encoded payload. Here it is intentionally
// trivial so the focus stays on orchestration rather than parsing logic.
func ExtractDocumentTextActivity(_ context.Context, input models.DocumentInput) (string, error) {
	return input.Content, nil
}

// ChunkDocumentActivity splits the extracted text into fixed-size word chunks.
//
// Chunking is deterministic: given the same input it always produces the same
// output. No randomness or external calls.
func ChunkDocumentActivity(_ context.Context, text string) ([]string, error) {
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
// All failure injection is deterministic and reads from input.Scenario only.
// No global state or randomness is involved.
func SummariseDocumentActivity(ctx context.Context, input models.SummariseInput) (models.SummariseResult, error) {
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

	chain := providers.DefaultChain()

	// Provider-level failure injection: replace the primary with a failing shim.
	// The failover loop tries primary (fails) then falls through to the fallback.
	if input.Scenario == models.ScenarioProviderFailover && len(chain) > 0 {
		chain[0] = providers.NewFaultyProvider(chain[0], "simulated primary provider failure")
	}

	return providers.SummariseWithFailover(ctx, chain, providers.SummariseRequest{Chunks: input.Chunks})
}
