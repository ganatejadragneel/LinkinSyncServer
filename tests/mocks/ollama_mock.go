package mocks

import "backend/services/ollama"

// MockOllamaService implements ollama.Service for testing
type MockOllamaService struct {
	AnalyzeLyricsFunc    func(query, lyrics, songInfo string) (string, error)
	GenerateResponseFunc func(prompt string) (string, error)
	IsAvailableFunc      func() error
}

// Ensure MockOllamaService implements ollama.Service
var _ ollama.Service = (*MockOllamaService)(nil)

// AnalyzeLyrics calls the mock function if set, otherwise returns default values
func (m *MockOllamaService) AnalyzeLyrics(query, lyrics, songInfo string) (string, error) {
	if m.AnalyzeLyricsFunc != nil {
		return m.AnalyzeLyricsFunc(query, lyrics, songInfo)
	}
	return "Mock analysis for: " + query, nil
}

// GenerateResponse calls the mock function if set, otherwise returns default values
func (m *MockOllamaService) GenerateResponse(prompt string) (string, error) {
	if m.GenerateResponseFunc != nil {
		return m.GenerateResponseFunc(prompt)
	}
	return "Mock response for: " + prompt, nil
}

// IsAvailable calls the mock function if set, otherwise returns nil
func (m *MockOllamaService) IsAvailable() error {
	if m.IsAvailableFunc != nil {
		return m.IsAvailableFunc()
	}
	return nil
}