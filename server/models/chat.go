package models

// ChatRequest represents a chat request from the user
type ChatRequest struct {
	Query string `json:"query"`
}

// ChatResponse represents a response to a chat request
type ChatResponse struct {
	Answer string `json:"answer"`
	Error  string `json:"error,omitempty"`
}