package models

// ChatRequest represents a chat request from the user
type ChatRequest struct {
	Query string `json:"query"`
}

// ChatResponse represents a response to a chat request
type ChatResponse struct {
	Answer          string                   `json:"answer"`
	Error           string                   `json:"error,omitempty"`
	Type            string                   `json:"type,omitempty"`            // "text" | "song_request" | "mood_recommendation"
	SongQuery       *SongQuery               `json:"song_query,omitempty"`      // Only present when Type is "song_request"
	MoodAnalysis    *MoodAnalysis            `json:"mood_analysis,omitempty"`   // Present when mood is detected
	Recommendations *MoodRecommendations     `json:"recommendations,omitempty"` // Present when Type is "mood_recommendation"
}

// SongQuery represents a parsed song request
type SongQuery struct {
	Query  string `json:"query"`  // The song name or search query
	Artist string `json:"artist,omitempty"` // Optional artist name
}

// MoodAnalysis represents the detected mood from user's message
type MoodAnalysis struct {
	PrimaryMood  string   `json:"primary_mood"`  // Main emotion detected
	MoodScore    float64  `json:"mood_score"`    // Confidence score 0-1
	EmotionTags  []string `json:"emotion_tags"`  // Related emotions/themes
}

// MoodRecommendations represents mood-based song recommendations
type MoodRecommendations struct {
	FromLibrary []MoodBasedRecommendation `json:"from_library"` // Songs from user's playlists (5)
	Suggested   []MoodBasedRecommendation `json:"suggested"`    // General suggestions (10)
}

// MoodBasedRecommendation represents a single mood-matched song recommendation
type MoodBasedRecommendation struct {
	Track       UnifiedTrack `json:"track"`
	MatchReason string       `json:"match_reason,omitempty"` // Why this song matches the mood
	MoodScore   float64      `json:"mood_score"`             // How well it matches (0-1)
}