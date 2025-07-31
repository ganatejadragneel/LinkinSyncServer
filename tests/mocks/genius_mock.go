package mocks

import "backend/services/genius"

// MockGeniusService implements genius.Service for testing
type MockGeniusService struct {
	GetLyricsFunc func(trackName, artistName string) (string, error)
}

// Ensure MockGeniusService implements genius.Service
var _ genius.Service = (*MockGeniusService)(nil)

// GetLyrics calls the mock function if set, otherwise returns default values
func (m *MockGeniusService) GetLyrics(trackName, artistName string) (string, error) {
	if m.GetLyricsFunc != nil {
		return m.GetLyricsFunc(trackName, artistName)
	}
	return "Mock lyrics for " + trackName + " by " + artistName, nil
}