package main

import (
	"backend/server/database"
	"backend/server/handlers"
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
)

// NowPlaying represents the currently playing song
type NowPlaying struct {
	TrackID   string `json:"track_id"`
	TrackName string `json:"track_name"`
	Artist    string `json:"artist"`
	Album     string `json:"album"`
	Lyrics    string `json:"lyrics,omitempty"`
	Mutex     sync.RWMutex
}

// Global variable to store currently playing song
var currentlyPlaying NowPlaying

// PlayHistory stores recently played tracks
type PlayHistory struct {
	Items []PlayHistoryItem `json:"items"`
	Mutex sync.RWMutex
}

// PlayHistoryItem represents a single song in play history
type PlayHistoryItem struct {
	TrackID   string    `json:"track_id"`
	TrackName string    `json:"track_name"`
	Artist    string    `json:"artist"`
	Album     string    `json:"album"`
	PlayedAt  time.Time `json:"played_at"`
}

// Global variable to store play history
var playHistory PlayHistory

// SpotifyTokenResponse represents the response from Spotify token API
type SpotifyTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// SpotifyTrack represents a track from Spotify
type SpotifyTrack struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Artist     string `json:"artist"`
	Album      string `json:"album"`
	PreviewURL string `json:"preview_url,omitempty"`
}

// ChatRequest represents a chat request from the user
type ChatRequest struct {
	Query string `json:"query"`
}

// ChatResponse represents a response to a chat request
type ChatResponse struct {
	Answer string `json:"answer"`
	Error  string `json:"error,omitempty"`
}

// LyricsHandler struct holds the LLM client
type LyricsHandler struct {
	llm *ollama.LLM
}

// Initialize a new LyricsHandler
func NewLyricsHandler() (*LyricsHandler, error) {
	llm, err := ollama.New(ollama.WithModel("llama3.2"))
	if err != nil {
		return nil, err
	}
	return &LyricsHandler{llm: llm}, nil
}

// Get Spotify access token using client credentials flow
func getSpotifyAccessToken() (string, error) {
	clientID := os.Getenv("SPOTIFY_CLIENT_ID")
	clientSecret := os.Getenv("SPOTIFY_CLIENT_SECRET")

	// Create auth string and encode to base64
	authString := fmt.Sprintf("%s:%s", clientID, clientSecret)
	encodedAuth := base64.StdEncoding.EncodeToString([]byte(authString))

	// Create form data
	data := url.Values{}
	data.Set("grant_type", "client_credentials")

	// Create request
	req, err := http.NewRequest("POST", "https://accounts.spotify.com/api/token", strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}

	// Set headers
	req.Header.Add("Authorization", fmt.Sprintf("Basic %s", encodedAuth))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Parse response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var tokenResponse SpotifyTokenResponse
	err = json.Unmarshal(body, &tokenResponse)
	if err != nil {
		return "", err
	}

	return tokenResponse.AccessToken, nil
}

// Get track details from Spotify by ID
func getSpotifyTrackById(trackID string) (*SpotifyTrack, error) {
	token, err := getSpotifyAccessToken()
	if err != nil {
		return nil, err
	}

	// Create request
	urlStr := fmt.Sprintf("https://api.spotify.com/v1/tracks/%s", trackID)
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return nil, err
	}

	// Set headers
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Parse response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	// Extract relevant information
	name := result["name"].(string)

	var artistName string
	if artists, ok := result["artists"].([]interface{}); ok && len(artists) > 0 {
		if artist, ok := artists[0].(map[string]interface{}); ok {
			artistName = artist["name"].(string)
		}
	}

	var albumName string
	if album, ok := result["album"].(map[string]interface{}); ok {
		albumName = album["name"].(string)
	}

	return &SpotifyTrack{
		ID:     trackID,
		Name:   name,
		Artist: artistName,
		Album:  albumName,
	}, nil
}

// Search for a song on Genius
func searchGeniusSong(trackName, artistName string) (string, error) {
	accessToken := os.Getenv("GENIUS_ACCESS_TOKEN")

	// Build query
	query := url.Values{}
	query.Add("q", fmt.Sprintf("%s %s", trackName, artistName))

	// Create request
	req, err := http.NewRequest("GET", "https://api.genius.com/search?"+query.Encode(), nil)
	if err != nil {
		return "", err
	}

	// Set headers
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Parse response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return "", err
	}

	// Navigate through the JSON to get the first hit's URL
	response, ok := result["response"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid response format")
	}

	hits, ok := response["hits"].([]interface{})
	if !ok || len(hits) == 0 {
		return "", fmt.Errorf("no hits found")
	}

	hit, ok := hits[0].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid hit format")
	}

	resultObj, ok := hit["result"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid result format")
	}

	url, ok := resultObj["url"].(string)
	if !ok {
		return "", fmt.Errorf("url not found")
	}

	return url, nil
}

// Scrape lyrics from Genius webpage
func scrapeLyricsFromGenius(geniusURL string) (string, error) {
	// Send request
	resp, err := http.Get(geniusURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Parse HTML
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", err
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

// Get lyrics for the current playing song
func getLyricsForCurrentSong() (string, error) {
	// Lock for reading
	currentlyPlaying.Mutex.RLock()
	trackName := currentlyPlaying.TrackName
	artist := currentlyPlaying.Artist
	savedLyrics := currentlyPlaying.Lyrics
	currentlyPlaying.Mutex.RUnlock()

	// If we already have lyrics, return them
	if savedLyrics != "" {
		return savedLyrics, nil
	}

	// If we don't have a current song, return an error
	if trackName == "" || artist == "" {
		return "", fmt.Errorf("no song is currently playing")
	}

	// Search for the song on Genius
	geniusURL, err := searchGeniusSong(trackName, artist)
	if err != nil {
		return "", fmt.Errorf("failed to find song on Genius: %v", err)
	}

	// Scrape lyrics from Genius
	lyrics, err := scrapeLyricsFromGenius(geniusURL)
	if err != nil {
		return "", fmt.Errorf("failed to scrape lyrics: %v", err)
	}

	// Save the lyrics for future use
	currentlyPlaying.Mutex.Lock()
	currentlyPlaying.Lyrics = lyrics
	currentlyPlaying.Mutex.Unlock()

	return lyrics, nil
}

// Ask the LLM about the lyrics
func askLLMAboutLyrics(llm *ollama.LLM, query, lyrics string) (string, error) {
	// Prepare context and prompt
	ctx := context.Background()

	// Construct a prompt with the lyrics and the user's query
	prompt := fmt.Sprintf(`
Here are the lyrics to a song:

%s

Based on these lyrics, please %s
`, lyrics, query)

	// Generate a response
	completion, err := llms.GenerateFromSinglePrompt(ctx, llm, prompt)
	if err != nil {
		return "", err
	}

	return completion, nil
}

// Handler for updating the currently playing song
func (h *LyricsHandler) UpdateNowPlaying(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var track SpotifyTrack
	err := json.NewDecoder(r.Body).Decode(&track)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Update currently playing song
	currentlyPlaying.Mutex.Lock()
	currentlyPlaying.TrackID = track.ID
	currentlyPlaying.TrackName = track.Name
	currentlyPlaying.Artist = track.Artist
	currentlyPlaying.Album = track.Album
	currentlyPlaying.Lyrics = "" // Reset lyrics so they'll be fetched fresh
	currentlyPlaying.Mutex.Unlock()

	// Add to play history
	playHistory.Mutex.Lock()
	historyItem := PlayHistoryItem{
		TrackID:   track.ID,
		TrackName: track.Name,
		Artist:    track.Artist,
		Album:     track.Album,
		PlayedAt:  time.Now(),
	}
	playHistory.Items = append([]PlayHistoryItem{historyItem}, playHistory.Items...)
	// Keep only last 10 songs
	if len(playHistory.Items) > 10 {
		playHistory.Items = playHistory.Items[:10]
	}
	playHistory.Mutex.Unlock()

	log.Printf("Now playing: %s by %s", track.Name, track.Artist)

	// Return success
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Now playing updated")
}

// Handler for getting the currently playing song
func (h *LyricsHandler) GetNowPlaying(w http.ResponseWriter, r *http.Request) {
	currentlyPlaying.Mutex.RLock()
	defer currentlyPlaying.Mutex.RUnlock()

	// Check if a song is playing
	if currentlyPlaying.TrackID == "" {
		http.Error(w, "No song is currently playing", http.StatusNotFound)
		return
	}

	// Return the currently playing song
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(currentlyPlaying)
}

// Handler for getting play history
func (h *LyricsHandler) GetPlayHistory(w http.ResponseWriter, r *http.Request) {
	playHistory.Mutex.RLock()
	defer playHistory.Mutex.RUnlock()

	// Return the play history
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(playHistory.Items)
}

// Handler for chat requests to the LLM
func (h *LyricsHandler) HandleChat(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var chatReq ChatRequest
	err := json.NewDecoder(r.Body).Decode(&chatReq)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Check if the query is about the current song's lyrics
	query := strings.ToLower(chatReq.Query)
	isAboutLyrics := strings.Contains(query, "lyric") ||
		strings.Contains(query, "song") ||
		strings.Contains(query, "meaning") ||
		strings.Contains(query, "interpretation") ||
		strings.Contains(query, "what does the song mean") ||
		strings.Contains(query, "explain")

	var response ChatResponse

	if isAboutLyrics {
		// Get lyrics for the current song
		lyrics, err := getLyricsForCurrentSong()
		if err != nil {
			response.Error = fmt.Sprintf("Error getting lyrics: %v", err)
		} else {
			// Ask the LLM about the lyrics
			answer, err := askLLMAboutLyrics(h.llm, chatReq.Query, lyrics)
			if err != nil {
				response.Error = fmt.Sprintf("Error from LLM: %v", err)
			} else {
				response.Answer = answer
			}
		}
	} else {
		// For non-lyrics related queries, just use the LLM directly
		ctx := context.Background()
		completion, err := llms.GenerateFromSinglePrompt(ctx, h.llm, chatReq.Query)
		if err != nil {
			response.Error = fmt.Sprintf("Error from LLM: %v", err)
		} else {
			response.Answer = completion
		}
	}

	// Return the response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Setup database tables if they don't exist
func setupDatabase(db *sql.DB) error {
	// Create table for global messages if it doesn't exist
	_, err := db.Exec(`
        CREATE TABLE IF NOT EXISTS global_messages (
            id SERIAL PRIMARY KEY,
            user_email VARCHAR(255) NOT NULL,
            username VARCHAR(255) NOT NULL,
            message_text TEXT NOT NULL,
            created_at TIMESTAMP WITH TIME ZONE NOT NULL
        )
    `)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file:", err)
	}

	// Initialize the database
	db, err := database.InitDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Setup database tables
	if err := setupDatabase(db); err != nil {
		log.Fatal("Error setting up database:", err)
	}

	// Initialize the LyricsHandler with Ollama LLM
	lyricsHandler, err := NewLyricsHandler()
	if err != nil {
		log.Fatalf("Error initializing LyricsHandler: %v", err)
	}

	// Initialize the router
	r := mux.NewRouter()

	// Setup the existing chat handler for global chat
	chatHandler := handlers.NewChatHandler(db)

	// Global chat routes
	r.HandleFunc("/api/messages", chatHandler.GetMessages).Methods("GET")
	r.HandleFunc("/api/messages", chatHandler.PostMessage).Methods("POST")

	// Music and lyrics routes
	r.HandleFunc("/api/now-playing", lyricsHandler.UpdateNowPlaying).Methods("POST")
	r.HandleFunc("/api/now-playing", lyricsHandler.GetNowPlaying).Methods("GET")
	r.HandleFunc("/api/history", lyricsHandler.GetPlayHistory).Methods("GET")
	r.HandleFunc("/api/chat", lyricsHandler.HandleChat).Methods("POST")

	// Setup CORS
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"http://localhost:3000"},
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type", "Authorization"},
	})

	// Get port from environment
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start the server
	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, c.Handler(r)))
}
