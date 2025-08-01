package handlers_test

import (
	"backend/repositories"
	"backend/server/handlers"
	"backend/server/models"
	"backend/tests/mocks"
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func createTestHandler() *handlers.LyricsHandler {
	mockGenius := &mocks.MockGeniusService{}
	mockOllama := &mocks.MockOllamaService{}
	musicRepo := repositories.NewMusicRepository(mockGenius)
	return handlers.NewLyricsHandler(musicRepo, mockOllama)
}

func TestLyricsHandler_UpdateNowPlaying(t *testing.T) {
	handler := createTestHandler()
	
	track := models.SpotifyTrack{
		ID:     "test123",
		Name:   "Test Song",
		Artist: "Test Artist",
		Album:  "Test Album",
	}
	
	body, _ := json.Marshal(track)
	req := httptest.NewRequest("POST", "/api/now-playing", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	handler.UpdateNowPlaying(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
	
	expected := "Now playing updated"
	if w.Body.String() != expected {
		t.Errorf("Expected body %s, got %s", expected, w.Body.String())
	}
}

func TestLyricsHandler_UpdateNowPlayingUnified(t *testing.T) {
	handler := createTestHandler()
	
	track := models.UnifiedTrack{
		ID:     "yt123",
		Name:   "Test YouTube Song",
		Artist: "Test YouTube Artist",
		Album:  "Test YouTube Album",
		Source: "youtube",
	}
	
	body, _ := json.Marshal(track)
	req := httptest.NewRequest("POST", "/api/now-playing", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	handler.UpdateNowPlaying(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
	
	expected := "Now playing updated"
	if w.Body.String() != expected {
		t.Errorf("Expected body %s, got %s", expected, w.Body.String())
	}
}

func TestLyricsHandler_UpdateNowPlaying_InvalidJSON(t *testing.T) {
	handler := createTestHandler()
	
	req := httptest.NewRequest("POST", "/api/now-playing", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	handler.UpdateNowPlaying(w, req)
	
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestLyricsHandler_UpdateNowPlaying_MissingFields(t *testing.T) {
	handler := createTestHandler()
	
	testCases := []struct {
		name string
		track interface{}
	}{
		{
			name: "SpotifyTrack missing ID",
			track: models.SpotifyTrack{
				Name:   "Test Song",
				Artist: "Test Artist",
				// Missing ID
			},
		},
		{
			name: "UnifiedTrack missing ID",
			track: models.UnifiedTrack{
				Name:   "Test Song",
				Artist: "Test Artist",
				Source: "youtube",
				// Missing ID
			},
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			body, _ := json.Marshal(tc.track)
			req := httptest.NewRequest("POST", "/api/now-playing", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			
			handler.UpdateNowPlaying(w, req)
			
			if w.Code != http.StatusBadRequest {
				t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
			}
		})
	}
}

func TestLyricsHandler_GetNowPlaying_NoSong(t *testing.T) {
	handler := createTestHandler()
	
	req := httptest.NewRequest("GET", "/api/now-playing", nil)
	w := httptest.NewRecorder()
	
	handler.GetNowPlaying(w, req)
	
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestLyricsHandler_GetNowPlaying_WithSong(t *testing.T) {
	mockGenius := &mocks.MockGeniusService{}
	mockOllama := &mocks.MockOllamaService{}
	musicRepo := repositories.NewMusicRepository(mockGenius)
	handler := handlers.NewLyricsHandler(musicRepo, mockOllama)
	
	// First update a song
	track := models.SpotifyTrack{
		ID:     "test123",
		Name:   "Test Song",
		Artist: "Test Artist",
		Album:  "Test Album",
	}
	musicRepo.UpdateNowPlaying(track)
	
	req := httptest.NewRequest("GET", "/api/now-playing", nil)
	w := httptest.NewRecorder()
	
	handler.GetNowPlaying(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
	
	var result models.NowPlaying
	err := json.Unmarshal(w.Body.Bytes(), &result)
	if err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}
	
	if result.TrackID != track.ID {
		t.Errorf("Expected TrackID %s, got %s", track.ID, result.TrackID)
	}
}

func TestLyricsHandler_GetPlayHistory(t *testing.T) {
	mockGenius := &mocks.MockGeniusService{}
	mockOllama := &mocks.MockOllamaService{}
	musicRepo := repositories.NewMusicRepository(mockGenius)
	handler := handlers.NewLyricsHandler(musicRepo, mockOllama)
	
	// Add a song to history
	track := models.SpotifyTrack{
		ID:     "test123",
		Name:   "Test Song",
		Artist: "Test Artist",
		Album:  "Test Album",
	}
	musicRepo.UpdateNowPlaying(track)
	
	req := httptest.NewRequest("GET", "/api/history", nil)
	w := httptest.NewRecorder()
	
	handler.GetPlayHistory(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
	
	var history []models.PlayHistoryItem
	err := json.Unmarshal(w.Body.Bytes(), &history)
	if err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}
	
	if len(history) != 1 {
		t.Errorf("Expected 1 item in history, got %d", len(history))
	}
	
	if history[0].TrackID != track.ID {
		t.Errorf("Expected TrackID %s, got %s", track.ID, history[0].TrackID)
	}
}

func TestLyricsHandler_HandleChat_InvalidJSON(t *testing.T) {
	handler := createTestHandler()
	
	req := httptest.NewRequest("POST", "/api/chat", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	handler.HandleChat(w, req)
	
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestLyricsHandler_HandleChat_EmptyQuery(t *testing.T) {
	handler := createTestHandler()
	
	chatReq := models.ChatRequest{Query: ""}
	body, _ := json.Marshal(chatReq)
	req := httptest.NewRequest("POST", "/api/chat", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	handler.HandleChat(w, req)
	
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestLyricsHandler_HandleChat_LyricsQuery_NoSong(t *testing.T) {
	handler := createTestHandler()
	
	chatReq := models.ChatRequest{Query: "tell me about the current song"}
	body, _ := json.Marshal(chatReq)
	req := httptest.NewRequest("POST", "/api/chat", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	handler.HandleChat(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
	
	var response models.ChatResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}
	
	if response.Answer == "" {
		t.Error("Expected non-empty answer")
	}
	
	if response.Error != "" {
		t.Errorf("Expected no error, got %s", response.Error)
	}
}

func TestLyricsHandler_HandleChat_LyricsQuery_WithSong(t *testing.T) {
	mockGenius := &mocks.MockGeniusService{
		GetLyricsFunc: func(trackName, artistName string) (string, error) {
			return "Mock lyrics", nil
		},
	}
	mockOllama := &mocks.MockOllamaService{
		AnalyzeLyricsFunc: func(query, lyrics, songInfo string) (string, error) {
			return "Mock analysis of the song", nil
		},
	}
	musicRepo := repositories.NewMusicRepository(mockGenius)
	handler := handlers.NewLyricsHandler(musicRepo, mockOllama)
	
	// Add a song
	track := models.SpotifyTrack{
		ID:     "test123",
		Name:   "Test Song",
		Artist: "Test Artist",
		Album:  "Test Album",
	}
	musicRepo.UpdateNowPlaying(track)
	
	chatReq := models.ChatRequest{Query: "what does this song mean?"}
	body, _ := json.Marshal(chatReq)
	req := httptest.NewRequest("POST", "/api/chat", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	handler.HandleChat(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
	
	var response models.ChatResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}
	
	if response.Answer != "Mock analysis of the song" {
		t.Errorf("Expected mock analysis, got %s", response.Answer)
	}
	
	if response.Error != "" {
		t.Errorf("Expected no error, got %s", response.Error)
	}
}

func TestLyricsHandler_HandleChat_GeneralQuery_NonMusic(t *testing.T) {
	handler := createTestHandler()
	
	chatReq := models.ChatRequest{Query: "What is artificial intelligence?"}
	body, _ := json.Marshal(chatReq)
	req := httptest.NewRequest("POST", "/api/chat", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	handler.HandleChat(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
	
	var response models.ChatResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}
	
	expectedMessage := "I can only help with questions about music, songs, lyrics, and artists. Please ask me something related to music!"
	if response.Answer != expectedMessage {
		t.Errorf("Expected restriction message, got %s", response.Answer)
	}
}

func TestLyricsHandler_HandleChat_GeneralQuery_Music(t *testing.T) {
	mockOllama := &mocks.MockOllamaService{
		GenerateResponseFunc: func(prompt string) (string, error) {
			return "Mock music response", nil
		},
	}
	mockGenius := &mocks.MockGeniusService{}
	musicRepo := repositories.NewMusicRepository(mockGenius)
	handler := handlers.NewLyricsHandler(musicRepo, mockOllama)
	
	chatReq := models.ChatRequest{Query: "What is jazz music?"}
	body, _ := json.Marshal(chatReq)
	req := httptest.NewRequest("POST", "/api/chat", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	handler.HandleChat(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
	
	var response models.ChatResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}
	
	if response.Answer != "Mock music response" {
		t.Errorf("Expected mock music response, got %s", response.Answer)
	}
}

func TestLyricsHandler_HandleChat_OllamaError_MusicQuery(t *testing.T) {
	mockOllama := &mocks.MockOllamaService{
		GenerateResponseFunc: func(prompt string) (string, error) {
			return "", errors.New("ollama service error")
		},
	}
	mockGenius := &mocks.MockGeniusService{}
	musicRepo := repositories.NewMusicRepository(mockGenius)
	handler := handlers.NewLyricsHandler(musicRepo, mockOllama)
	
	chatReq := models.ChatRequest{Query: "What is jazz music?"}
	body, _ := json.Marshal(chatReq)
	req := httptest.NewRequest("POST", "/api/chat", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	handler.HandleChat(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
	
	var response models.ChatResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}
	
	if response.Error == "" {
		t.Error("Expected error in response")
	}
	
	if response.Answer != "" {
		t.Errorf("Expected empty answer, got %s", response.Answer)
	}
}

func TestLyricsHandler_HandleChat_NonMusicQueries(t *testing.T) {
	handler := createTestHandler()
	
	nonMusicQueries := []string{
		"What is artificial intelligence?",
		"How do I cook pasta?",
		"What is the weather today?",
		"How to code in Python?",
		"What is mathematics?",
		"How do I fix my car?",
	}
	
	for _, query := range nonMusicQueries {
		t.Run(query, func(t *testing.T) {
			chatReq := models.ChatRequest{Query: query}
			body, _ := json.Marshal(chatReq)
			req := httptest.NewRequest("POST", "/api/chat", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			
			handler.HandleChat(w, req)
			
			if w.Code != http.StatusOK {
				t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
			}
			
			var response models.ChatResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			if err != nil {
				t.Errorf("Failed to unmarshal response: %v", err)
			}
			
			expectedMessage := "I can only help with questions about music, songs, lyrics, and artists. Please ask me something related to music!"
			if response.Answer != expectedMessage {
				t.Errorf("Expected restriction message for query '%s', got %s", query, response.Answer)
			}
		})
	}
}

func TestLyricsHandler_HandleChat_MusicQueries(t *testing.T) {
	mockOllama := &mocks.MockOllamaService{
		GenerateResponseFunc: func(prompt string) (string, error) {
			return "Mock music response", nil
		},
	}
	mockGenius := &mocks.MockGeniusService{}
	musicRepo := repositories.NewMusicRepository(mockGenius)
	handler := handlers.NewLyricsHandler(musicRepo, mockOllama)
	
	musicQueries := []string{
		"What is jazz music?",
		"How does a guitar work?",
		"What are musical scales?",
		"Who is the best singer?",
		"What is hip hop genre?",
		"How do drums work?",
	}
	
	for _, query := range musicQueries {
		t.Run(query, func(t *testing.T) {
			chatReq := models.ChatRequest{Query: query}
			body, _ := json.Marshal(chatReq)
			req := httptest.NewRequest("POST", "/api/chat", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			
			handler.HandleChat(w, req)
			
			if w.Code != http.StatusOK {
				t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
			}
			
			var response models.ChatResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			if err != nil {
				t.Errorf("Failed to unmarshal response: %v", err)
			}
			
			if response.Answer != "Mock music response" {
				t.Errorf("Expected music response for query '%s', got %s", query, response.Answer)
			}
		})
	}
}