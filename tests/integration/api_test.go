package integration_test

import (
	"backend/repositories"
	"backend/server/handlers"
	"backend/server/models"
	"backend/tests/mocks"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

// setupTestServer creates a test server with all routes and mocked dependencies
func setupTestServer() *httptest.Server {
	// Create mocked services
	mockGenius := &mocks.MockGeniusService{
		GetLyricsFunc: func(trackName, artistName string) (string, error) {
			return "Mock lyrics for " + trackName + " by " + artistName, nil
		},
	}
	
	mockOllama := &mocks.MockOllamaService{
		AnalyzeLyricsFunc: func(query, lyrics, songInfo string) (string, error) {
			return "Mock analysis: " + query, nil
		},
		GenerateResponseFunc: func(prompt string) (string, error) {
			return "Mock response: " + prompt, nil
		},
	}
	
	// Create repositories and handlers
	musicRepo := repositories.NewMusicRepository(mockGenius)
	lyricsHandler := handlers.NewLyricsHandler(musicRepo, mockOllama)
	
	// Setup router
	r := mux.NewRouter()
	api := r.PathPrefix("/api").Subrouter()
	
	// Add routes
	api.HandleFunc("/now-playing", lyricsHandler.UpdateNowPlaying).Methods("POST")
	api.HandleFunc("/now-playing", lyricsHandler.GetNowPlaying).Methods("GET")
	api.HandleFunc("/history", lyricsHandler.GetPlayHistory).Methods("GET")
	api.HandleFunc("/chat", lyricsHandler.HandleChat).Methods("POST")
	api.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")
	
	return httptest.NewServer(r)
}

func TestAPI_HealthCheck(t *testing.T) {
	server := setupTestServer()
	defer server.Close()
	
	resp, err := http.Get(server.URL + "/api/health")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

func TestAPI_NowPlayingWorkflow(t *testing.T) {
	server := setupTestServer()
	defer server.Close()
	
	// Test 1: GET now-playing when nothing is playing
	resp, err := http.Get(server.URL + "/api/now-playing")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	resp.Body.Close()
	
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, resp.StatusCode)
	}
	
	// Test 2: POST a new track
	track := models.SpotifyTrack{
		ID:     "test123",
		Name:   "Test Song",
		Artist: "Test Artist",
		Album:  "Test Album",
	}
	
	body, _ := json.Marshal(track)
	resp, err = http.Post(server.URL+"/api/now-playing", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
	
	// Test 3: GET now-playing after setting it
	resp, err = http.Get(server.URL + "/api/now-playing")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
	
	var nowPlaying models.NowPlaying
	err = json.NewDecoder(resp.Body).Decode(&nowPlaying)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	
	if nowPlaying.TrackID != track.ID {
		t.Errorf("Expected TrackID %s, got %s", track.ID, nowPlaying.TrackID)
	}
	if nowPlaying.TrackName != track.Name {
		t.Errorf("Expected TrackName %s, got %s", track.Name, nowPlaying.TrackName)
	}
}

func TestAPI_PlayHistoryWorkflow(t *testing.T) {
	server := setupTestServer()
	defer server.Close()
	
	// Add multiple tracks
	tracks := []models.SpotifyTrack{
		{ID: "track1", Name: "Song 1", Artist: "Artist 1", Album: "Album 1"},
		{ID: "track2", Name: "Song 2", Artist: "Artist 2", Album: "Album 2"},
		{ID: "track3", Name: "Song 3", Artist: "Artist 3", Album: "Album 3"},
	}
	
	for _, track := range tracks {
		body, _ := json.Marshal(track)
		resp, err := http.Post(server.URL+"/api/now-playing", "application/json", bytes.NewBuffer(body))
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		resp.Body.Close()
	}
	
	// Get play history
	resp, err := http.Get(server.URL + "/api/history")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
	
	var history []models.PlayHistoryItem
	err = json.NewDecoder(resp.Body).Decode(&history)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	
	if len(history) != 3 {
		t.Errorf("Expected 3 items in history, got %d", len(history))
	}
	
	// Most recent should be first
	if history[0].TrackID != "track3" {
		t.Errorf("Expected first item to be track3, got %s", history[0].TrackID)
	}
}

func TestAPI_ChatWorkflow_NonMusic(t *testing.T) {
	server := setupTestServer()
	defer server.Close()
	
	// Test non-music query (should be restricted)
	chatReq := models.ChatRequest{Query: "What is artificial intelligence?"}
	body, _ := json.Marshal(chatReq)
	
	resp, err := http.Post(server.URL+"/api/chat", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
	
	var chatResp models.ChatResponse
	err = json.NewDecoder(resp.Body).Decode(&chatResp)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	
	expectedMessage := "I can only help with questions about music, songs, lyrics, and artists. Please ask me something related to music!"
	if chatResp.Answer != expectedMessage {
		t.Errorf("Expected restriction message, got %s", chatResp.Answer)
	}
	
	if chatResp.Error != "" {
		t.Errorf("Expected no error, got %s", chatResp.Error)
	}
}

func TestAPI_ChatWorkflow_Music(t *testing.T) {
	server := setupTestServer()
	defer server.Close()
	
	// Test music-related query
	chatReq := models.ChatRequest{Query: "What is jazz music?"}
	body, _ := json.Marshal(chatReq)
	
	resp, err := http.Post(server.URL+"/api/chat", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
	
	var chatResp models.ChatResponse
	err = json.NewDecoder(resp.Body).Decode(&chatResp)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	
	if chatResp.Answer == "" {
		t.Error("Expected non-empty answer")
	}
	
	if chatResp.Error != "" {
		t.Errorf("Expected no error, got %s", chatResp.Error)
	}
}

func TestAPI_ChatAboutSongWorkflow(t *testing.T) {
	server := setupTestServer()
	defer server.Close()
	
	// First, set a current song
	track := models.SpotifyTrack{
		ID:     "test123",
		Name:   "Test Song",
		Artist: "Test Artist",
		Album:  "Test Album",
	}
	
	body, _ := json.Marshal(track)
	resp, err := http.Post(server.URL+"/api/now-playing", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	resp.Body.Close()
	
	// Now ask about the song
	chatReq := models.ChatRequest{Query: "tell me about the current song"}
	body, _ = json.Marshal(chatReq)
	
	resp, err = http.Post(server.URL+"/api/chat", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
	
	var chatResp models.ChatResponse
	err = json.NewDecoder(resp.Body).Decode(&chatResp)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	
	if chatResp.Answer == "" {
		t.Error("Expected non-empty answer")
	}
	
	if chatResp.Error != "" {
		t.Errorf("Expected no error, got %s", chatResp.Error)
	}
}

func TestAPI_ChatAboutSongNoCurrentSong(t *testing.T) {
	server := setupTestServer()
	defer server.Close()
	
	// Ask about song when none is playing
	chatReq := models.ChatRequest{Query: "what does the current song mean?"}
	body, _ := json.Marshal(chatReq)
	
	resp, err := http.Post(server.URL+"/api/chat", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
	
	var chatResp models.ChatResponse
	err = json.NewDecoder(resp.Body).Decode(&chatResp)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	
	if chatResp.Answer == "" {
		t.Error("Expected non-empty answer explaining no song is playing")
	}
	
	if chatResp.Error != "" {
		t.Errorf("Expected no error, got %s", chatResp.Error)
	}
}

func TestAPI_InvalidRequests(t *testing.T) {
	server := setupTestServer()
	defer server.Close()
	
	// Test invalid JSON for now-playing
	resp, err := http.Post(server.URL+"/api/now-playing", "application/json", bytes.NewBufferString("invalid json"))
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	resp.Body.Close()
	
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
	
	// Test invalid JSON for chat
	resp, err = http.Post(server.URL+"/api/chat", "application/json", bytes.NewBufferString("invalid json"))
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	resp.Body.Close()
	
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
	
	// Test empty chat query
	chatReq := models.ChatRequest{Query: ""}
	body, _ := json.Marshal(chatReq)
	resp, err = http.Post(server.URL+"/api/chat", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	resp.Body.Close()
	
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
}