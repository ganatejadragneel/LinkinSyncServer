package openai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Config holds OpenAI service configuration
type Config struct {
	APIKey      string
	Model       string
	BaseURL     string
	Temperature float64
	MaxTokens   int
	TopP        float64
}

// DefaultConfig returns a default configuration for OpenAI
func DefaultConfig() Config {
	return Config{
		Model:       "gpt-3.5-turbo",
		BaseURL:     "https://api.openai.com/v1",
		Temperature: 0.7,
		MaxTokens:   500,
		TopP:        0.9,
	}
}

// service implements the OpenAI Service interface
type service struct {
	config     Config
	httpClient *http.Client
}

// New creates a new OpenAI service
func New(config Config) Service {
	return &service{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// IsAvailable checks if the OpenAI service is available
func (s *service) IsAvailable() error {
	if s.config.APIKey == "" {
		return fmt.Errorf("OpenAI API key not provided")
	}
	
	// Test with a simple request
	req := ChatCompletionRequest{
		Model: s.config.Model,
		Messages: []Message{
			{Role: "user", Content: "Test"},
		},
		MaxTokens: func() *int { v := 1; return &v }(),
	}
	
	_, err := s.makeRequest(req)
	if err != nil {
		return fmt.Errorf("OpenAI service not available: %w", err)
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

// generate sends a request to OpenAI and returns the response
func (s *service) generate(prompt string) (string, error) {
	req := ChatCompletionRequest{
		Model: s.config.Model,
		Messages: []Message{
			{Role: "user", Content: prompt},
		},
		Temperature: &s.config.Temperature,
		MaxTokens:   &s.config.MaxTokens,
		TopP:        &s.config.TopP,
	}
	
	resp, err := s.makeRequest(req)
	if err != nil {
		return "", err
	}
	
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response choices returned from OpenAI")
	}
	
	return resp.Choices[0].Message.Content, nil
}

// makeRequest sends a request to OpenAI API
func (s *service) makeRequest(req ChatCompletionRequest) (*ChatCompletionResponse, error) {
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	
	httpReq, err := http.NewRequest("POST", s.config.BaseURL+"/chat/completions", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.config.APIKey))
	
	// Send request
	httpResp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to OpenAI: %w", err)
	}
	defer httpResp.Body.Close()
	
	// Read response body
	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	
	// Parse response
	var openaiResp ChatCompletionResponse
	if err := json.Unmarshal(body, &openaiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	// Check for API errors
	if openaiResp.Error != nil {
		return nil, fmt.Errorf("OpenAI API error: %s", openaiResp.Error.Message)
	}
	
	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenAI API failed with status %d: %s", httpResp.StatusCode, string(body))
	}
	
	return &openaiResp, nil
}