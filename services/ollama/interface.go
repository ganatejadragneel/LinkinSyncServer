package ollama

// Service defines the interface for Ollama/AI operations
type Service interface {
	// AnalyzeLyrics analyzes lyrics based on a user query
	AnalyzeLyrics(query, lyrics, songInfo string) (string, error)
	
	// GenerateResponse generates a general response without lyrics context
	GenerateResponse(prompt string) (string, error)
	
	// IsAvailable checks if the Ollama service is available
	IsAvailable() error
}