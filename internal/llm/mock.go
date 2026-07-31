package llm

import (
	"context"
	"fmt"
)

// MockLLM is a mock LLM client for testing.
// It returns a pre-configured response or a simple echo.
type MockLLM struct {
	responses []string
	callCount int
}

// NewMockLLM creates a new MockLLM.
func NewMockLLM() *MockLLM {
	return &MockLLM{}
}

// WithResponses configures the mock to return specific responses in sequence.
func (m *MockLLM) WithResponses(responses ...string) *MockLLM {
	m.responses = responses
	return m
}

// Complete returns the next pre-configured response or an echo of the prompt.
func (m *MockLLM) Complete(ctx context.Context, req CompletionRequest) (CompletionResponse, error) {
	if m.callCount < len(m.responses) {
		resp := m.responses[m.callCount]
		m.callCount++
		return CompletionResponse{Text: resp}, nil
	}
	// Fallback: echo the prompt
	m.callCount++
	return CompletionResponse{
		Text: fmt.Sprintf("[MOCK RESPONSE %d]: received prompt with %d chars", m.callCount, len(req.UserPrompt)),
	}, nil
}

// ProviderName returns "mock".
func (m *MockLLM) ProviderName() string {
	return "mock"
}

// CallCount returns how many times Complete was called.
func (m *MockLLM) CallCount() int {
	return m.callCount
}

// Reset resets the call count.
func (m *MockLLM) Reset() {
	m.callCount = 0
}
