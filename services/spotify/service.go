package spotify

import (
	"backend/server/models"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Config holds Spotify API configuration
type Config struct {
	ClientID     string
	ClientSecret string
}

// service implements the Spotify Service interface
type service struct {
	config      Config
	httpClient  *http.Client
	accessToken string
	tokenExpiry time.Time
}

// New creates a new Spotify service
func New(config Config) Service {
	return &service{
		config: config,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetAccessToken gets or refreshes the Spotify access token
func (s *service) GetAccessToken() (string, error) {
	// Check if we have a valid token
	if s.accessToken != "" && time.Now().Before(s.tokenExpiry) {
		return s.accessToken, nil
	}

	// Create auth string and encode to base64
	authString := fmt.Sprintf("%s:%s", s.config.ClientID, s.config.ClientSecret)
	encodedAuth := base64.StdEncoding.EncodeToString([]byte(authString))

	// Create form data
	data := url.Values{}
	data.Set("grant_type", "client_credentials")

	// Create request
	req, err := http.NewRequest("POST", "https://accounts.spotify.com/api/token", strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Add("Authorization", fmt.Sprintf("Basic %s", encodedAuth))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// Send request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("spotify auth failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var tokenResponse models.SpotifyTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	// Store token and expiry
	s.accessToken = tokenResponse.AccessToken
	s.tokenExpiry = time.Now().Add(time.Duration(tokenResponse.ExpiresIn-60) * time.Second) // Subtract 60s for safety

	return s.accessToken, nil
}

// GetTrackByID gets track details from Spotify by ID
func (s *service) GetTrackByID(trackID string) (*models.SpotifyTrack, error) {
	token, err := s.GetAccessToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	// Create request
	urlStr := fmt.Sprintf("https://api.spotify.com/v1/tracks/%s", trackID)
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	// Send request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("spotify API failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Extract relevant information
	track := &models.SpotifyTrack{
		ID:   trackID,
		Name: s.getString(result, "name"),
	}

	// Extract artist name
	if artists, ok := result["artists"].([]interface{}); ok && len(artists) > 0 {
		if artist, ok := artists[0].(map[string]interface{}); ok {
			track.Artist = s.getString(artist, "name")
		}
	}

	// Extract album name
	if album, ok := result["album"].(map[string]interface{}); ok {
		track.Album = s.getString(album, "name")
	}

	// Extract preview URL if available
	track.PreviewURL = s.getString(result, "preview_url")

	return track, nil
}

// getString safely extracts a string from a map
func (s *service) getString(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}