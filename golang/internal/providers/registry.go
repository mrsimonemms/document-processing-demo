package providers

// DefaultChain returns the ordered provider chain used when no custom policy
// is specified. Primary provider is OpenAI; fallback is Anthropic.
//
// To change the provider order for the demo, edit the slice here.
// New providers can be added by appending to this slice.
func DefaultChain() []Summariser {
	return []Summariser{&OpenAI{}, &Anthropic{}}
}

// DefaultQuestionChain returns the ordered provider chain for Q&A.
// Primary provider is OpenAI; fallback is Anthropic.
func DefaultQuestionChain() []QuestionAnswerer {
	return []QuestionAnswerer{&OpenAI{}, &Anthropic{}}
}
