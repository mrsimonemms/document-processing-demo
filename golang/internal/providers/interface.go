package providers

import (
	"context"
	"fmt"
	"strings"

	"github.com/mrsimonemms/document-processing-demo/golang/internal/models"
)

// SummariseRequest is the input to a provider's Summarise call.
type SummariseRequest struct {
	Chunks []string
}

// SummariseResponse is the output from a provider's Summarise call.
type SummariseResponse struct {
	Summary string
}

// Summariser is the interface each provider must implement for summarisation.
// Keeping it to two methods makes it straightforward to add real providers later.
type Summariser interface {
	Name() models.ProviderName
	Summarise(ctx context.Context, req SummariseRequest) (SummariseResponse, error)
}

// AnswerRequest is the input to a provider's Answer call.
type AnswerRequest struct {
	Content  string
	Question string
	// History holds recent Q&A pairs from the session for conversational context.
	// It is already bounded by the workflow before the activity is called.
	History []models.QA
}

// AnswerResponse is the output from a provider's Answer call.
type AnswerResponse struct {
	Answer string
}

// QuestionAnswerer is the interface each provider must implement for Q&A.
// It is separate from Summariser so both capabilities can evolve independently.
type QuestionAnswerer interface {
	Name() models.ProviderName
	Answer(ctx context.Context, req AnswerRequest) (AnswerResponse, error)
}

// buildSummary produces a deterministic summary string for fake providers.
// The label identifies which provider generated the output, which is useful
// during a demo to make provider selection visible.
func buildSummary(label string, req SummariseRequest) string {
	wordCount := 0
	for _, chunk := range req.Chunks {
		wordCount += len(strings.Fields(chunk))
	}

	preview := ""
	if len(req.Chunks) > 0 {
		preview = req.Chunks[0]
		if len(preview) > 120 {
			preview = preview[:120] + "..."
		}
	}

	return fmt.Sprintf("[%s] %d chunk(s), %d word(s). Preview: %s",
		label, len(req.Chunks), wordCount, preview)
}

// buildAnswer produces a deterministic answer string for fake providers.
// The label identifies the provider. In a real integration this would be
// replaced by an LLM call grounded on the supplied content.
func buildAnswer(label string, req AnswerRequest) string {
	words := strings.Fields(req.Content)

	context := req.Content
	if len(context) > 100 {
		context = context[:100] + "..."
	}

	return fmt.Sprintf("[%s] Answer to '%s' (%d word(s) of context): %s",
		label, req.Question, len(words), context)
}
