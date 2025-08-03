package models

// ChatRequest represents a chat request from the user
type ChatRequest struct {
	Query string `json:"query"`
}

// ChatResponse represents a response to a chat request
type ChatResponse struct {
	Answer    string     `json:"answer"`
	Error     string     `json:"error,omitempty"`
	Type      string     `json:"type,omitempty"`       // "text" | "song_request"
	SongQuery *SongQuery `json:"song_query,omitempty"` // Only present when Type is "song_request"
}

// SongQuery represents a parsed song request
type SongQuery struct {
	Query  string `json:"query"`  // The song name or search query
	Artist string `json:"artist,omitempty"` // Optional artist name
}