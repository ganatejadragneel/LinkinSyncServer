package handlers

import (
	"backend/repositories"
	"backend/server/models"
	"backend/services/ollama"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

// LyricsHandler handles lyrics-related HTTP requests
type LyricsHandler struct {
	musicRepo     *repositories.MusicRepository
	ollamaService ollama.Service
}

// NewLyricsHandler creates a new lyrics handler
func NewLyricsHandler(musicRepo *repositories.MusicRepository, ollamaService ollama.Service) *LyricsHandler {
	return &LyricsHandler{
		musicRepo:     musicRepo,
		ollamaService: ollamaService,
	}
}

// UpdateNowPlaying handles POST /api/now-playing
func (h *LyricsHandler) UpdateNowPlaying(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var track models.SpotifyTrack
	if err := json.NewDecoder(r.Body).Decode(&track); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if track.ID == "" || track.Name == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Update the currently playing track
	h.musicRepo.UpdateNowPlaying(track)

	log.Printf("Now playing updated: %s by %s", track.Name, track.Artist)

	// Return success
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Now playing updated")
}

// GetNowPlaying handles GET /api/now-playing
func (h *LyricsHandler) GetNowPlaying(w http.ResponseWriter, r *http.Request) {
	// Check if a song is playing
	if !h.musicRepo.IsPlaying() {
		http.Error(w, "No song is currently playing", http.StatusNotFound)
		return
	}

	// Get the currently playing song
	nowPlaying := h.musicRepo.GetNowPlaying()

	// Return the currently playing song
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(nowPlaying)
}

// GetPlayHistory handles GET /api/history
func (h *LyricsHandler) GetPlayHistory(w http.ResponseWriter, r *http.Request) {
	history := h.musicRepo.GetPlayHistory()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}

// HandleChat handles POST /api/chat
func (h *LyricsHandler) HandleChat(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var chatReq models.ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&chatReq); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate query
	if chatReq.Query == "" {
		http.Error(w, "Query cannot be empty", http.StatusBadRequest)
		return
	}

	// Process the chat request
	response := h.processChatRequest(chatReq.Query)

	// Return the response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// processChatRequest processes a chat request and returns a response
func (h *LyricsHandler) processChatRequest(query string) models.ChatResponse {
	// Check if the query is about lyrics/music
	if h.isLyricsRelatedQuery(query) {
		return h.handleLyricsQuery(query)
	}

	// Handle general queries
	return h.handleGeneralQuery(query)
}

// isLyricsRelatedQuery checks if a query is specifically about current song lyrics
func (h *LyricsHandler) isLyricsRelatedQuery(query string) bool {
	lowerQuery := strings.ToLower(query)
	
	// Specific patterns for current song/lyrics queries
	lyricsPatterns := []string{
		"current song", "current track", "now playing", "playing now",
		"this song", "this track", "about the song", "about the track",
		"song mean", "track mean", "lyrics mean", "song about",
		"tell me about", "what does", "explain the", "meaning of",
	}

	// Must contain "lyric" OR match specific current song patterns
	if strings.Contains(lowerQuery, "lyric") {
		return true
	}
	
	for _, pattern := range lyricsPatterns {
		if strings.Contains(lowerQuery, pattern) {
			return true
		}
	}
	
	return false
}

// isMusicRelatedQuery checks if a general query is related to music topics
func (h *LyricsHandler) isMusicRelatedQuery(query string) bool {
	lowerQuery := strings.ToLower(query)
	musicKeywords := []string{
		"music", "song", "artist", "band", "album", "track",
		"genre", "musician", "singer", "composer", "producer",
		"concert", "performance", "instrument", "guitar", "piano",
		"drums", "vocal", "melody", "harmony", "rhythm",
		"beat", "tempo", "chord", "scale", "key",
		"recording", "studio", "label", "release", "single",
		"ep", "mixtape", "soundtrack", "cover", "remix",
		"acoustic", "electric", "classical", "jazz", "rock",
		"pop", "hip hop", "rap", "country", "folk",
		"blues", "metal", "punk", "indie", "electronic",
	}

	for _, keyword := range musicKeywords {
		if strings.Contains(lowerQuery, keyword) {
			return true
		}
	}
	return false
}

// handleLyricsQuery handles queries related to lyrics
func (h *LyricsHandler) handleLyricsQuery(query string) models.ChatResponse {
	// Check if we have a current song
	if !h.musicRepo.IsPlaying() {
		return models.ChatResponse{
			Answer: "No song is currently playing. Please play a song in Spotify first, and I'll be able to help you understand its lyrics and meaning.",
		}
	}

	// Get song info
	songInfo := h.musicRepo.GetCurrentSongInfo()

	// Try to get lyrics
	lyrics, err := h.musicRepo.GetLyricsForCurrentSong()
	if err != nil {
		// If we can't get lyrics, provide what information we can
		return models.ChatResponse{
			Answer: fmt.Sprintf("I can see that you're currently playing \"%s\", but I couldn't fetch the lyrics: %v\n\nYou can still ask me general questions about this song or artist!", songInfo, err),
		}
	}

	// Ask Ollama to analyze the lyrics
	answer, err := h.ollamaService.AnalyzeLyrics(query, lyrics, songInfo)
	if err != nil {
		return models.ChatResponse{
			Error: fmt.Sprintf("Error analyzing lyrics: %v", err),
		}
	}

	return models.ChatResponse{
		Answer: answer,
	}
}

// handleGeneralQuery handles general queries not related to lyrics
func (h *LyricsHandler) handleGeneralQuery(query string) models.ChatResponse {
	// Check if query is music-related
	if !h.isMusicRelatedQuery(query) {
		return models.ChatResponse{
			Answer: "I can only help with questions about music, songs, lyrics, and artists. Please ask me something related to music!",
		}
	}

	// For music-related general queries, provide a concise response
	musicPrompt := fmt.Sprintf("Answer this music question in EXACTLY 2 short paragraphs. Keep it brief - maximum 4-5 sentences per paragraph: %s", query)
	answer, err := h.ollamaService.GenerateResponse(musicPrompt)
	if err != nil {
		return models.ChatResponse{
			Error: fmt.Sprintf("Error generating response: %v", err),
		}
	}

	return models.ChatResponse{
		Answer: answer,
	}
}