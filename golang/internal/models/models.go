package models

// TaskQueue is the Temporal task queue name used by the worker and workflow starter.
const TaskQueue = "document-processing"

// Scenario selects which failure-injection behaviour is active for a workflow run.
type Scenario string

const (
	// ScenarioHappyPath runs all activities without injected failures.
	ScenarioHappyPath Scenario = "happy_path"

	// ScenarioFailOnceSummarise causes the summarise activity to fail on its
	// first attempt. Temporal retries automatically and the second attempt succeeds,
	// demonstrating durable retry behaviour.
	ScenarioFailOnceSummarise Scenario = "fail_once_summarise"

	// ScenarioProviderFailover causes the primary AI provider to fail
	// deterministically, forcing the failover mechanism to use the fallback
	// provider. Demonstrates provider-level resilience without real API calls.
	ScenarioProviderFailover Scenario = "primary_provider_failure"
)

// ProviderName identifies an AI provider.
type ProviderName string

const (
	// ProviderOpenAI is the default primary provider.
	ProviderOpenAI ProviderName = "openai"

	// ProviderAnthropic is the default fallback provider.
	ProviderAnthropic ProviderName = "anthropic"
)

// ProviderOverride selects which provider(s) to use for LLM calls in a session.
type ProviderOverride string

const (
	// ProviderOverrideDefault uses the full provider chain: OpenAI primary with
	// Anthropic as fallback. This is the standard demo configuration.
	ProviderOverrideDefault ProviderOverride = "default"

	// ProviderOverrideOpenAI uses OpenAI only, with no fallback.
	// A provider failure scenario will cause the session to fail.
	ProviderOverrideOpenAI ProviderOverride = "openai"

	// ProviderOverrideAnthropic uses Anthropic only, with no fallback.
	ProviderOverrideAnthropic ProviderOverride = "anthropic"
)

// DocumentInput is the workflow input. It carries everything the workflow
// needs, including the chosen failure scenario and provider override.
type DocumentInput struct {
	DocumentID       string           `json:"documentId"`
	Content          string           `json:"content"`
	Scenario         Scenario         `json:"scenario"`
	ProviderOverride ProviderOverride `json:"providerOverride"`
}

// QA holds a single question and its answer, stored in workflow state.
type QA struct {
	Question string       `json:"question"`
	Answer   string       `json:"answer"`
	Provider ProviderName `json:"provider,omitempty"`
	Model    string       `json:"model,omitempty"`
}

// DocumentState is the result of the getState query on the document workflow.
// Phase transitions: "processing" -> "summarised" -> "ended".
// The UI polls this to drive page state.
type DocumentState struct {
	Phase            string           `json:"phase"` // "processing" | "summarised" | "ended"
	Summary          string           `json:"summary,omitempty"`
	Provider         ProviderName     `json:"provider,omitempty"`
	Model            string           `json:"model,omitempty"`
	FallbackOccurred bool             `json:"fallbackOccurred"`
	QA               []QA             `json:"qa"`
	ProviderOverride ProviderOverride `json:"providerOverride,omitempty"`
}

// SummariseInput is passed to SummariseDocumentActivity.
// Scenario and ProviderOverride are threaded through so the activity can apply
// failure injection and provider selection without relying on global state.
type SummariseInput struct {
	Chunks           []string         `json:"chunks"`
	Scenario         Scenario         `json:"scenario"`
	ProviderOverride ProviderOverride `json:"providerOverride"`
}

// SummariseResult is the output of SummariseDocumentActivity.
// It carries the summary and metadata about which provider produced it.
type SummariseResult struct {
	Summary          string       `json:"summary"`
	Provider         ProviderName `json:"provider"`
	Model            string       `json:"model"`
	FallbackOccurred bool         `json:"fallbackOccurred"`
}

// QuestionUpdate is the input to the askQuestion update handler on the
// document workflow. Scenario and ProviderOverride are applied per question.
type QuestionUpdate struct {
	Question         string           `json:"question"`
	Scenario         Scenario         `json:"scenario"`
	ProviderOverride ProviderOverride `json:"providerOverride"`
}

// QuestionUpdateResult is the output of the askQuestion update handler.
type QuestionUpdateResult struct {
	Answer string `json:"answer"`
}

// MaxAnswerHistory is the maximum number of recent Q&A pairs included in the
// context passed to AnswerQuestionActivity. Keeping this small bounds prompt
// size deterministically without token counting.
const MaxAnswerHistory = 5

// AnswerInput is passed to AnswerQuestionActivity.
type AnswerInput struct {
	Content          string           `json:"content"`
	Question         string           `json:"question"`
	Scenario         Scenario         `json:"scenario"`
	ProviderOverride ProviderOverride `json:"providerOverride"`
	// History holds the most recent Q&A pairs from the session, capped at
	// MaxAnswerHistory entries. Providers use this for conversational context.
	History []QA `json:"history,omitempty"`
}

// AnswerResult is the output of AnswerQuestionActivity.
// It mirrors SummariseResult in shape: the answer plus provider metadata.
type AnswerResult struct {
	Answer           string       `json:"answer"`
	Provider         ProviderName `json:"provider"`
	Model            string       `json:"model"`
	FallbackOccurred bool         `json:"fallbackOccurred"`
}
