package services_test

import (
	"backend/services/ollama"
	"testing"
)

func TestOllamaConfig_DefaultConfig(t *testing.T) {
	config := ollama.DefaultConfig()
	
	if config.BaseURL != "http://localhost:11434" {
		t.Errorf("Expected BaseURL http://localhost:11434, got %s", config.BaseURL)
	}
	
	if config.Model != "llama3.2:3b" {
		t.Errorf("Expected Model llama3.2:3b, got %s", config.Model)
	}
	
	if config.Temperature != 0.7 {
		t.Errorf("Expected Temperature 0.7, got %f", config.Temperature)
	}
	
	if config.TopP != 0.9 {
		t.Errorf("Expected TopP 0.9, got %f", config.TopP)
	}
	
	if config.TopK != 40 {
		t.Errorf("Expected TopK 40, got %d", config.TopK)
	}
}

func TestOllamaService_New(t *testing.T) {
	config := ollama.DefaultConfig()
	service := ollama.New(config)
	
	if service == nil {
		t.Fatal("New should not return nil")
	}
}

// Note: These tests would require a running Ollama instance to test actual functionality
// For unit tests, we'd typically mock the HTTP client, but that would require refactoring
// the service to accept an HTTP client interface. For now, we test the configuration
// and basic instantiation.

func TestOllamaService_CustomConfig(t *testing.T) {
	config := ollama.Config{
		BaseURL:     "http://custom:11434",
		Model:       "custom-model",
		Temperature: 0.5,
		TopP:        0.8,
		TopK:        30,
	}
	
	service := ollama.New(config)
	
	if service == nil {
		t.Fatal("New should not return nil with custom config")
	}
	
	// Test that the service was created with the custom config
	// (We can't directly test the config values without exposing them,
	// but we can ensure the service was created successfully)
}