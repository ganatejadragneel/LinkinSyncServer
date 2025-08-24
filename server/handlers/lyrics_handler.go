package handlers

import (
	"backend/repositories"
	"backend/server/models"
	"backend/services/mood"
	// "backend/services/ollama"  // Uncomment when using Ollama
	"backend/services/openai"
	"backend/services/spotify"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

// AIService defines a common interface for AI services (both Ollama and OpenAI)
type AIService interface {
	AnalyzeLyrics(query, lyrics, songInfo string) (string, error)
	GenerateResponse(prompt string) (string, error)
	IsAvailable() error
}

// LyricsHandler handles lyrics-related HTTP requests
type LyricsHandler struct {
	musicRepo      *repositories.MusicRepository
	// AI Services - comment/uncomment to switch between providers
	aiService      AIService  // Currently active AI service
	// ollamaService  ollama.Service  // Uncomment to use Ollama
	openaiService  openai.Service  // Comment to disable OpenAI
	moodService    mood.Service
	spotifyService spotify.Service
}

// NewLyricsHandler creates a new lyrics handler
func NewLyricsHandler(
	musicRepo *repositories.MusicRepository,
	// ollamaService ollama.Service,  // Uncomment to use Ollama
	openaiService openai.Service,  // Comment to disable OpenAI
	moodService mood.Service,
	spotifyService spotify.Service,
) *LyricsHandler {
	handler := &LyricsHandler{
		musicRepo:      musicRepo,
		// ollamaService:  ollamaService,  // Uncomment to use Ollama
		openaiService:  openaiService,  // Comment to disable OpenAI
		moodService:    moodService,
		spotifyService: spotifyService,
	}
	
	// Set the active AI service - comment/uncomment to switch
	// handler.aiService = ollamaService  // Use Ollama
	handler.aiService = openaiService  // Use OpenAI
	
	return handler
}

// UpdateNowPlaying handles POST /api/now-playing
func (h *LyricsHandler) UpdateNowPlaying(w http.ResponseWriter, r *http.Request) {
	// Parse request body into generic map first
	var trackData map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&trackData); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Check if it has a source field (UnifiedTrack) or default to spotify
	if source, hasSource := trackData["source"]; hasSource && source != nil {
		// Parse as UnifiedTrack
		trackBytes, _ := json.Marshal(trackData)
		var unifiedTrack models.UnifiedTrack
		if err := json.Unmarshal(trackBytes, &unifiedTrack); err != nil {
			http.Error(w, "Invalid unified track data", http.StatusBadRequest)
			return
		}
		
		// Validate required fields
		if unifiedTrack.ID == "" || unifiedTrack.Name == "" {
			http.Error(w, "Missing required fields", http.StatusBadRequest)
			return
		}
		
		// Update the currently playing track
		h.musicRepo.UpdateNowPlayingUnified(unifiedTrack)
		log.Printf("Now playing updated (%s): %s by %s", unifiedTrack.Source, unifiedTrack.Name, unifiedTrack.Artist)
	} else {
		// Parse as SpotifyTrack for backward compatibility
		trackBytes, _ := json.Marshal(trackData)
		var track models.SpotifyTrack
		if err := json.Unmarshal(trackBytes, &track); err != nil {
			http.Error(w, "Invalid spotify track data", http.StatusBadRequest)
			return
		}

		// Validate required fields
		if track.ID == "" || track.Name == "" {
			http.Error(w, "Missing required fields", http.StatusBadRequest)
			return
		}

		// Update the currently playing track
		h.musicRepo.UpdateNowPlaying(track)
		log.Printf("Now playing updated (spotify): %s by %s", track.Name, track.Artist)
	}

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
	// Check if the query is a song request first
	if h.isSongRequestQuery(query) {
		return h.handleSongRequest(query)
	}
	
	// Check if the query contains emotional content that needs mood-based recommendations
	if h.containsEmotionalContent(query) {
		return h.handleMoodBasedQuery(query)
	}
	
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

	// Ask AI service to analyze the lyrics
	answer, err := h.aiService.AnalyzeLyrics(query, lyrics, songInfo)
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
	answer, err := h.aiService.GenerateResponse(musicPrompt)
	if err != nil {
		return models.ChatResponse{
			Error: fmt.Sprintf("Error generating response: %v", err),
		}
	}

	return models.ChatResponse{
		Answer: answer,
	}
}

// isSongRequestQuery checks if a query is a song request
func (h *LyricsHandler) isSongRequestQuery(query string) bool {
	lowerQuery := strings.ToLower(query)
	
	songRequestPatterns := []string{
		"play ", "can you play", "find ", "search for",
		"put on ", "i want to hear", "i want to listen to",
		"show me ", "look for ", "get me ",
	}
	
	for _, pattern := range songRequestPatterns {
		if strings.Contains(lowerQuery, pattern) {
			return true
		}
	}
	
	return false
}

// extractSongRequest extracts song name and artist from a song request query
func (h *LyricsHandler) extractSongRequest(query string) (string, string) {
	lowerQuery := strings.ToLower(query)
	
	// Patterns with "by" separator for artist
	if strings.Contains(lowerQuery, " by ") {
		parts := strings.Split(lowerQuery, " by ")
		if len(parts) == 2 {
			// Extract song from first part
			songPart := strings.TrimSpace(parts[0])
			artistPart := strings.TrimSpace(parts[1])
			
			// Remove common prefixes
			prefixes := []string{"play ", "find ", "search for ", "put on ", "show me ", "look for ", "get me "}
			for _, prefix := range prefixes {
				if strings.HasPrefix(songPart, prefix) {
					songPart = strings.TrimSpace(songPart[len(prefix):])
					break
				}
			}
			
			// Clean up quotes
			songPart = strings.Trim(songPart, `"'`)
			artistPart = strings.Trim(artistPart, `"'`)
			
			if songPart != "" && artistPart != "" {
				return songPart, artistPart
			}
		}
	}
	
	// Patterns without artist - just extract song name
	songOnlyPatterns := []string{
		"play ", "find ", "search for ", "put on ", 
		"show me ", "look for ", "get me ",
		"can you play ", "i want to hear ", "i want to listen to ",
	}
	
	for _, prefix := range songOnlyPatterns {
		if strings.HasPrefix(lowerQuery, prefix) {
			song := strings.TrimSpace(lowerQuery[len(prefix):])
			// Clean up quotes and common suffixes
			song = strings.Trim(song, `"'`)
			song = strings.TrimSuffix(song, " please")
			song = strings.TrimSuffix(song, " song")
			
			if song != "" {
				return song, ""
			}
		}
	}
	
	// If no specific pattern matches, return the whole query as song name
	return strings.TrimSpace(query), ""
}

// handleSongRequest handles song request queries
func (h *LyricsHandler) handleSongRequest(query string) models.ChatResponse {
	songName, artist := h.extractSongRequest(query)
	
	// Create song query object
	songQuery := &models.SongQuery{
		Query:  songName,
		Artist: artist,
	}
	
	// Create response message
	var responseMsg string
	if artist != "" {
		responseMsg = fmt.Sprintf("I'm searching for \"%s\" by %s in your playlists. Let me show you what I found!", songName, artist)
	} else {
		responseMsg = fmt.Sprintf("I'm searching for \"%s\" in your playlists. Let me show you what I found!", songName)
	}
	
	return models.ChatResponse{
		Answer:    responseMsg,
		Type:      "song_request",
		SongQuery: songQuery,
	}
}

// containsEmotionalContent checks if the query contains emotional or mood-related content
func (h *LyricsHandler) containsEmotionalContent(query string) bool {
	lowerQuery := strings.ToLower(query)
	
	// Emotional keywords that suggest mood-based recommendations
	emotionalKeywords := []string{
		"feel", "feeling", "mood", "emotion", "sad", "happy", "angry", "upset",
		"depressed", "anxious", "lonely", "alone", "stressed", "overwhelmed",
		"excited", "joy", "love", "hate", "frustrated", "confused", "lost",
		"hurt", "broken", "empty", "hopeless", "worried", "scared", "afraid",
		"nervous", "calm", "peaceful", "nostalgic", "miss", "remember",
		"belong", "disconnected", "isolated", "abandoned", "rejected",
		"won", "victory", "celebrate", "celebration", "achievement", "accomplished",
		"tournament", "competition", "winning", "winner",
	}
	
	// Check for emotional content
	for _, keyword := range emotionalKeywords {
		if strings.Contains(lowerQuery, keyword) {
			// Additional context check - ensure it's about the user's feelings
			personalIndicators := []string{"i ", "i'm", "i am", "me ", "my ", "feel", "feeling"}
			for _, indicator := range personalIndicators {
				if strings.Contains(lowerQuery, indicator) {
					log.Printf("Emotional content detected in query: '%s' (keyword: %s, indicator: %s)", 
						query, keyword, indicator)
					return true
				}
			}
		}
	}
	
	log.Printf("No emotional content detected in query: '%s'", query)
	return false
}

// handleMoodBasedQuery handles queries that contain emotional content
func (h *LyricsHandler) handleMoodBasedQuery(query string) models.ChatResponse {
	// Detect mood from the query
	moodAnalysis, err := h.moodService.DetectMood(query)
	if err != nil {
		log.Printf("Error detecting mood: %v", err)
		return h.handleGeneralQuery(query) // Fallback to general query
	}
	
	// Get user's playlists and liked songs
	userTracks, err := h.getUserLibraryTracks()
	if err != nil {
		log.Printf("Error getting user tracks: %v", err)
		return models.ChatResponse{
			Answer: "I understand you're feeling " + moodAnalysis.PrimaryMood + ", but I'm having trouble accessing your music library right now. Please try again later.",
			MoodAnalysis: moodAnalysis,
		}
	}
	
	// Find mood-matched songs from user's library (5 songs)
	libraryMatches, err := h.moodService.MatchSongsToMood(moodAnalysis, userTracks, 5)
	if err != nil {
		log.Printf("Error matching songs to mood: %v", err)
	}
	
	// Get general song suggestions (10 songs)
	generalSuggestions := h.getGeneralMoodSuggestions(moodAnalysis.PrimaryMood, 10)
	
	// Create empathetic response
	response := h.createEmpatheticResponse(moodAnalysis.PrimaryMood, query)
	
	// Save mood history (using a dummy user ID for now - should get from auth context)
	userID := "default_user" // TODO: Get actual user ID from request context
	var playedSongIDs []string
	for _, match := range libraryMatches {
		playedSongIDs = append(playedSongIDs, match.Track.ID)
	}
	h.moodService.SaveUserMoodHistory(userID, moodAnalysis.PrimaryMood, playedSongIDs)
	
	log.Printf("Mood detected: %s, Library matches: %d, General suggestions: %d", 
		moodAnalysis.PrimaryMood, len(libraryMatches), len(generalSuggestions))
	
	return models.ChatResponse{
		Answer:       response,
		Type:         "mood_recommendation",
		MoodAnalysis: moodAnalysis,
		Recommendations: &models.MoodRecommendations{
			FromLibrary: libraryMatches,
			Suggested:   generalSuggestions,
		},
	}
}

// getUserLibraryTracks gets tracks from user's playlists and liked songs
func (h *LyricsHandler) getUserLibraryTracks() ([]models.UnifiedTrack, error) {
	var allTracks []models.UnifiedTrack
	
	// TODO: This should get the actual user's access token from request context
	// For now, returning empty slice
	// In production, this would:
	// 1. Get user's Spotify playlists and liked songs
	// 2. Get user's YouTube liked videos (music only)
	// 3. Combine and deduplicate
	
	return allTracks, nil
}

// getGeneralMoodSuggestions returns general song suggestions for a mood
func (h *LyricsHandler) getGeneralMoodSuggestions(mood string, limit int) []models.MoodBasedRecommendation {
	// Predefined mood-based suggestions
	moodSuggestions := map[string][]models.MoodBasedRecommendation{
		"lonely": {
			{
				Track: models.UnifiedTrack{
					ID:     "1mea3bSkSGXuIRvnydlB5b", // Real Spotify ID for "Somewhere I Belong"
					Name:   "Somewhere I Belong",
					Artist: "Linkin Park",
					Album:  "Meteora",
					Source: "spotify",
				},
				MoodScore:   0.95,
			},
			{
				Track: models.UnifiedTrack{
					ID:     "4N3y2ChKKCG3zVCfyNiMQD", // Real Spotify ID for "Mad World"
					Name:   "Mad World",
					Artist: "Gary Jules",
					Album:  "Trading Snakeoil for Wolftickets",
					Source: "spotify",
				},
				MoodScore:   0.90,
			},
			{
				Track: models.UnifiedTrack{
					ID:     "u9HBEOlMgOtK8yXGKKMhRx", // Real Spotify ID for "The Sound of Silence"
					Name:   "The Sound of Silence",
					Artist: "Disturbed",
					Album:  "Immortalized",
					Source: "spotify",
				},
				MoodScore:   0.88,
			},
		},
		"sad": {
			{
				Track: models.UnifiedTrack{
					ID:     "2DjPkzR89MSYPGaWhK8uKQ", // Real Spotify ID for "Hurt" by Johnny Cash
					Name:   "Hurt",
					Artist: "Johnny Cash",
					Album:  "American IV: The Man Comes Around",
					Source: "spotify",
				},
				MoodScore:   0.95,
			},
			{
				Track: models.UnifiedTrack{
					ID:     "0SiQrCn2h2aKOEqz5Zxwow", // Real Spotify ID for "The Night We Met"
					Name:   "The Night We Met",
					Artist: "Lord Huron",
					Album:  "Strange Trails",
					Source: "spotify",
				},
				MoodScore:   0.90,
			},
		},
		"happy": {
			{
				Track: models.UnifiedTrack{
					ID:     "3BxnGCLFNdLKgVgVz6Vn5H", // Real Spotify ID for "Good Life"
					Name:   "Good Life",
					Artist: "OneRepublic",
					Album:  "Waking Up",
					Source: "spotify",
				},
				MoodScore:   0.95,
			},
			{
				Track: models.UnifiedTrack{
					ID:     "05wIrZSwuaVWhcv5FfqeJ0", // Real Spotify ID for "Walking on Sunshine"
					Name:   "Walking on Sunshine",
					Artist: "Katrina and the Waves",
					Album:  "Walking on Sunshine",
					Source: "spotify",
				},
				MoodScore:   0.93,
			},
			{
				Track: models.UnifiedTrack{
					ID:     "60nZcImufyMA1MKQY3dcCH", // Real Spotify ID for "Happy"
					Name:   "Happy",
					Artist: "Pharrell Williams",
					Album:  "G I R L",
					Source: "spotify",
				},
				MoodScore:   0.98,
			},
			{
				Track: models.UnifiedTrack{
					ID:     "0BxE4FqsDD1Ot4YuBXwn8F", // Real Spotify ID for "Can't Stop the Feeling!"
					Name:   "Can't Stop the Feeling!",
					Artist: "Justin Timberlake",
					Album:  "Trolls (Original Motion Picture Soundtrack)",
					Source: "spotify",
				},
				MoodScore:   0.96,
			},
			{
				Track: models.UnifiedTrack{
					ID:     "32OlwWuMpZ6b0aN2RZOeMS", // Real Spotify ID for "Uptown Funk"
					Name:   "Uptown Funk",
					Artist: "Mark Ronson ft. Bruno Mars",
					Album:  "Uptown Special",
					Source: "spotify",
				},
				MoodScore:   0.94,
			},
			{
				Track: models.UnifiedTrack{
					ID:     "1WkMMavIMc4JZ8cfMmxHkI", // Real Spotify ID for "Good as Hell"
					Name:   "Good as Hell",
					Artist: "Lizzo",
					Album:  "Cuz I Love You",
					Source: "spotify",
				},
				MoodScore:   0.92,
			},
			{
				Track: models.UnifiedTrack{
					ID:     "0CFuMybe6s77w6QQrJjW7d", // Real Spotify ID for "I'm Gonna Be (500 Miles)"
					Name:   "I'm Gonna Be (500 Miles)",
					Artist: "The Proclaimers",
					Album:  "Sunshine on Leith",
					Source: "spotify",
				},
				MoodScore:   0.90,
			},
			{
				Track: models.UnifiedTrack{
					ID:     "5T8EDUDqKcs6OSOwEsfqG7", // Real Spotify ID for "Don't Stop Me Now"
					Name:   "Don't Stop Me Now",
					Artist: "Queen",
					Album:  "Jazz",
					Source: "spotify",
				},
				MoodScore:   0.88,
			},
			{
				Track: models.UnifiedTrack{
					ID:     "2RlgNHKcydI9sayD2Df2xp", // Real Spotify ID for "Mr. Blue Sky"
					Name:   "Mr. Blue Sky",
					Artist: "Electric Light Orchestra",
					Album:  "Out of the Blue",
					Source: "spotify",
				},
				MoodScore:   0.86,
			},
			{
				Track: models.UnifiedTrack{
					ID:     "3PPogGhAUjr4FLGzEFGzJI", // Real Spotify ID for "Best Day of My Life"
					Name:   "Best Day of My Life",
					Artist: "American Authors",
					Album:  "Oh, What a Life",
					Source: "spotify",
				},
				MoodScore:   0.84,
			},
		},
		"angry": {
			{
				Track: models.UnifiedTrack{
					ID:     "2OzEKCmOoWhyuB8nHi8xhv", // Real Spotify ID for "Break Stuff"
					Name:   "Break Stuff",
					Artist: "Limp Bizkit",
					Album:  "Significant Other",
					Source: "spotify",
				},
				MoodScore:   0.95,
			},
			{
				Track: models.UnifiedTrack{
					ID:     "0yp7ORA8XPNO4kvNj5EYdx", // Real Spotify ID for "Bodies"
					Name:   "Bodies",
					Artist: "Drowning Pool",
					Album:  "Sinner",
					Source: "spotify",
				},
				MoodScore:   0.92,
			},
		},
	}
	
	// Get suggestions for the mood
	suggestions, exists := moodSuggestions[mood]
	if !exists {
		// Default suggestions if mood not found
		suggestions = moodSuggestions["sad"]
	}
	
	// Return up to limit suggestions
	if len(suggestions) > limit {
		return suggestions[:limit]
	}
	
	return suggestions
}

// createEmpatheticResponse creates an empathetic response based on mood
func (h *LyricsHandler) createEmpatheticResponse(mood, originalQuery string) string {
	responses := map[string]string{
		"lonely":    "I hear you're feeling disconnected right now. Sometimes music can be a companion when we feel alone. Here are some songs that explore similar feelings and might resonate with you:",
		"sad":       "I understand you're going through a difficult time. Music has a way of expressing what we can't always put into words. These songs might help you process these feelings:",
		"happy":     "It's wonderful that you're feeling so positive! Let's keep that energy going with some uplifting tracks that match your mood:",
		"angry":     "I can sense your frustration. Sometimes we need music that matches our intensity and helps us release these feelings. Here are some powerful tracks for you:",
		"anxious":   "I understand you're feeling overwhelmed. These songs might help you find some calm or at least know you're not alone in feeling this way:",
		"nostalgic": "Ah, feeling nostalgic... Music has a unique way of taking us back. Here are some songs that capture that bittersweet feeling of remembering:",
		"energetic": "You're full of energy! Let's channel that into some high-powered tracks that'll keep you motivated:",
		"calm":      "Finding your peace... Here are some tranquil songs to help maintain that serene state of mind:",
	}
	
	response, exists := responses[mood]
	if !exists {
		response = fmt.Sprintf("I can sense you're feeling %s. Music has a way of connecting with our emotions. Here are some songs that might resonate with how you're feeling:", mood)
	}
	
	return response
}