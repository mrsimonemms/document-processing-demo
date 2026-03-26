package providers

import (
	"context"

	"github.com/mrsimonemms/document-processing-demo/golang/internal/models"
)

// SummariseRequest is the input to a provider's Summarise call.
type SummariseRequest struct {
	Chunks []string
}

// SummariseResponse is the output from a provider's Summarise call.
type SummariseResponse struct {
	Summary string
	Model   string
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
	Model  string
}

// QuestionAnswerer is the interface each provider must implement for Q&A.
// It is separate from Summariser so both capabilities can evolve independently.
type QuestionAnswerer interface {
	Name() models.ProviderName
	Answer(ctx context.Context, req AnswerRequest) (AnswerResponse, error)
}
