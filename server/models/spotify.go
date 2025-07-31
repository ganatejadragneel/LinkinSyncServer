package models

// SpotifyTrack represents a track from Spotify
type SpotifyTrack struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Artist     string `json:"artist"`
	Album      string `json:"album"`
	PreviewURL string `json:"preview_url,omitempty"`
}

// SpotifyTokenResponse represents the response from Spotify token API
type SpotifyTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}