package providers

// NewChain returns an ordered provider chain for summarisation.
// The first entry is the primary provider; subsequent entries are fallbacks.
// Pass the constructed OpenAI provider as primary and the Anthropic fake as fallback.
//
// To change the provider order for the demo, edit the slice in main.go.
func NewChain(primary Summariser, fallbacks ...Summariser) []Summariser {
	return append([]Summariser{primary}, fallbacks...)
}

// NewQuestionChain returns an ordered provider chain for Q&A.
// The first entry is the primary provider; subsequent entries are fallbacks.
func NewQuestionChain(primary QuestionAnswerer, fallbacks ...QuestionAnswerer) []QuestionAnswerer {
	return append([]QuestionAnswerer{primary}, fallbacks...)
}
