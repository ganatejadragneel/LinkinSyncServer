package spotify

import "backend/server/models"

// Service defines the interface for Spotify operations
type Service interface {
	GetAccessToken() (string, error)
	GetTrackByID(trackID string) (*models.SpotifyTrack, error)
}