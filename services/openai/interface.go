package openai

// Service defines the interface for OpenAI operations
type Service interface {
	// AnalyzeLyrics analyzes lyrics based on a user query
	AnalyzeLyrics(query, lyrics, songInfo string) (string, error)
	
	// GenerateResponse generates a general response without lyrics context
	GenerateResponse(prompt string) (string, error)
	
	// IsAvailable checks if the OpenAI service is available
	IsAvailable() error
}