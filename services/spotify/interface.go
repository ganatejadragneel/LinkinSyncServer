package spotify

import "backend/server/models"

// Service defines the Spotify service interface
type Service interface {
	GetAccessToken() (string, error)
	GetTrackByID(trackID string) (*models.SpotifyTrack, error)
}