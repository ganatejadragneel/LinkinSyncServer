package ollama

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Config holds Ollama service configuration
type Config struct {
	BaseURL     string
	Model       string
	Temperature float64
	TopP        float64
	TopK        int
}

// DefaultConfig returns a default configuration
func DefaultConfig() Config {
	return Config{
		BaseURL:     "http://localhost:11434",
		Model:       "llama3.2:3b",
		Temperature: 0.7,
		TopP:        0.9,
		TopK:        40,
	}
}

// service implements the Ollama Service interface
type service struct {
	config     Config
	httpClient *http.Client
}

// New creates a new Ollama service
func New(config Config) Service {
	return &service{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// IsAvailable checks if the Ollama service is available
func (s *service) IsAvailable() error {
	resp, err := s.httpClient.Get(s.config.BaseURL + "/api/tags")
	if err != nil {
		return fmt.Errorf("ollama server not available at %s: %w", s.config.BaseURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ollama server not responding correctly, status: %d", resp.StatusCode)
	}

	return nil
}

// AnalyzeLyrics analyzes lyrics based on a user query
func (s *service) AnalyzeLyrics(query, lyrics, songInfo string) (string, error) {
	prompt := s.buildLyricsPrompt(query, lyrics, songInfo)
	return s.generate(prompt)
}

// GenerateResponse generates a general response without lyrics context
func (s *service) GenerateResponse(prompt string) (string, error) {
	return s.generate(prompt)
}

// buildLyricsPrompt creates a prompt for lyrics analysis
func (s *service) buildLyricsPrompt(query, lyrics, songInfo string) string {
	return fmt.Sprintf(`You are analyzing "%s". Answer in EXACTLY 2 short paragraphs only. Be concise.

Question: %s

Keep it brief - maximum 4-5 sentences per paragraph. Focus only on the most important points.`, songInfo, query)
}

// generate sends a request to Ollama and returns the response
func (s *service) generate(prompt string) (string, error) {
	// Prepare the request
	req := Request{
		Model:  s.config.Model,
		Prompt: prompt,
		Stream: false,
		Options: map[string]interface{}{
			"temperature": s.config.Temperature,
			"top_p":       s.config.TopP,
			"top_k":       s.config.TopK,
		},
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Send request to Ollama
	resp, err := s.httpClient.Post(
		s.config.BaseURL+"/api/generate",
		"application/json",
		bytes.NewBuffer(reqBody),
	)
	if err != nil {
		return "", fmt.Errorf("failed to send request to Ollama: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ollama API failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var ollamaResp Response
	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if ollamaResp.Error != "" {
		return "", fmt.Errorf("ollama error: %s", ollamaResp.Error)
	}

	return ollamaResp.Response, nil
}