package genius

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// Config holds Genius API configuration
type Config struct {
	AccessToken string
}

// service implements the Genius Service interface
type service struct {
	config     Config
	httpClient *http.Client
}

// New creates a new Genius service
func New(config Config) Service {
	return &service{
		config: config,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// GetLyrics fetches lyrics for a given track and artist
func (s *service) GetLyrics(trackName, artistName string) (string, error) {
	// Search for the song on Genius
	songURL, err := s.searchSong(trackName, artistName)
	if err != nil {
		return "", fmt.Errorf("failed to search song: %w", err)
	}

	// Scrape lyrics from the song page
	lyrics, err := s.scrapeLyrics(songURL)
	if err != nil {
		return "", fmt.Errorf("failed to scrape lyrics: %w", err)
	}

	return lyrics, nil
}

// searchSong searches for a song on Genius and returns the URL
func (s *service) searchSong(trackName, artistName string) (string, error) {
	// Build query
	query := url.Values{}
	query.Add("q", fmt.Sprintf("%s %s", trackName, artistName))

	// Create request
	req, err := http.NewRequest("GET", "https://api.genius.com/search?"+query.Encode(), nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", s.config.AccessToken))

	// Send request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("genius API failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	// Navigate through the JSON to get the first hit's URL
	response, ok := result["response"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid response format")
	}

	hits, ok := response["hits"].([]interface{})
	if !ok || len(hits) == 0 {
		return "", fmt.Errorf("no results found for %s by %s", trackName, artistName)
	}

	// Check first few results for best match
	for i, h := range hits {
		if i > 2 { // Check only first 3 results
			break
		}
		
		hit, ok := h.(map[string]interface{})
		if !ok {
			continue
		}

		resultObj, ok := hit["result"].(map[string]interface{})
		if !ok {
			continue
		}

		// Get the URL
		songURL, ok := resultObj["url"].(string)
		if !ok {
			continue
		}

		// Check if artist matches (case insensitive)
		primaryArtist, _ := resultObj["primary_artist"].(map[string]interface{})
		artistNameFromResult, _ := primaryArtist["name"].(string)
		
		if strings.Contains(strings.ToLower(artistNameFromResult), strings.ToLower(artistName)) ||
		   strings.Contains(strings.ToLower(artistName), strings.ToLower(artistNameFromResult)) {
			return songURL, nil
		}
	}

	// If no exact match, return the first result
	hit := hits[0].(map[string]interface{})
	resultObj := hit["result"].(map[string]interface{})
	songURL := resultObj["url"].(string)
	
	return songURL, nil
}

// scrapeLyrics scrapes lyrics from a Genius webpage
func (s *service) scrapeLyrics(geniusURL string) (string, error) {
	// Send request
	resp, err := s.httpClient.Get(geniusURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch page with status %d", resp.StatusCode)
	}

	// Parse HTML
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Try to find lyrics container in new Genius layout
	var lyrics strings.Builder
	foundLyrics := false

	// Look for lyrics containers with data-lyrics-container attribute
	doc.Find("div[data-lyrics-container='true']").Each(func(i int, s *goquery.Selection) {
		// Get text from each container
		text := s.Text()
		lyrics.WriteString(text)
		lyrics.WriteString("\n\n")
		foundLyrics = true
	})

	// If not found in new layout, try old layout
	if !foundLyrics {
		lyricsDiv := doc.Find("div.lyrics")
		if lyricsDiv.Length() > 0 {
			lyrics.WriteString(lyricsDiv.Text())
			foundLyrics = true
		}
	}

	if !foundLyrics {
		return "", fmt.Errorf("lyrics not found in the page structure")
	}

	return strings.TrimSpace(lyrics.String()), nil
}