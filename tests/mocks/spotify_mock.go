package mocks

import (
	"backend/server/models"
	"backend/services/spotify"
)

// MockSpotifyService implements spotify.Service for testing
type MockSpotifyService struct {
	GetAccessTokenFunc func() (string, error)
	GetTrackByIDFunc   func(trackID string) (*models.SpotifyTrack, error)
}

// Ensure MockSpotifyService implements spotify.Service
var _ spotify.Service = (*MockSpotifyService)(nil)

// GetAccessToken calls the mock function if set, otherwise returns default values
func (m *MockSpotifyService) GetAccessToken() (string, error) {
	if m.GetAccessTokenFunc != nil {
		return m.GetAccessTokenFunc()
	}
	return "mock_access_token", nil
}

// GetTrackByID calls the mock function if set, otherwise returns default values
func (m *MockSpotifyService) GetTrackByID(trackID string) (*models.SpotifyTrack, error) {
	if m.GetTrackByIDFunc != nil {
		return m.GetTrackByIDFunc(trackID)
	}
	return &models.SpotifyTrack{
		ID:     trackID,
		Name:   "Mock Song",
		Artist: "Mock Artist",
		Album:  "Mock Album",
	}, nil
}