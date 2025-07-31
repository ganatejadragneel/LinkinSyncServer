# Testing Guide

## Overview

This project includes comprehensive testing with unit tests, integration tests, and mocks to ensure code quality and reliability.

## Test Structure

```
tests/
├── unit/                    # Unit tests
│   ├── models/             # Model tests
│   ├── services/           # Service tests
│   ├── repositories/       # Repository tests
│   └── handlers/           # Handler tests
├── integration/            # Integration tests
│   └── api_test.go        # API endpoint tests
└── mocks/                  # Mock implementations
    ├── genius_mock.go     # Genius service mock
    ├── ollama_mock.go     # Ollama service mock
    └── spotify_mock.go    # Spotify service mock
```

## Running Tests

### All Tests
```bash
make test
```

### Unit Tests Only
```bash
make test-unit
```

### Integration Tests Only
```bash
make test-integration
```

### Specific Package Tests
```bash
make test-models
make test-handlers
make test-repositories
make test-services
```

### Test Coverage
```bash
make test-coverage
```
This generates `coverage.html` with a detailed coverage report.

## Test Categories

### 1. Unit Tests

#### Model Tests (`tests/unit/models/`)
- **NowPlaying**: Tests thread-safe operations, state management, and data integrity
- **PlayHistory**: Tests concurrent access, size limits, and data ordering

#### Repository Tests (`tests/unit/repositories/`)
- **MusicRepository**: Tests music data management, lyrics caching, and service integration

#### Handler Tests (`tests/unit/handlers/`)
- **LyricsHandler**: Tests HTTP request handling, validation, and response formatting

#### Service Tests (`tests/unit/services/`)
- **OllamaService**: Tests configuration and service creation

### 2. Integration Tests

#### API Tests (`tests/integration/`)
- **Complete workflows**: Tests entire request/response cycles
- **Error scenarios**: Tests invalid inputs and error handling
- **Multi-step processes**: Tests complex interactions between components

### 3. Mock Services

Located in `tests/mocks/`, these provide controlled, predictable behavior for testing:

- **MockGeniusService**: Simulates lyrics fetching
- **MockOllamaService**: Simulates AI chat responses
- **MockSpotifyService**: Simulates Spotify API calls

## Writing Tests

### Unit Test Example

```go
func TestNowPlaying_Update(t *testing.T) {
    np := models.NewNowPlaying()
    track := models.SpotifyTrack{
        ID:     "test123",
        Name:   "Test Song",
        Artist: "Test Artist",
        Album:  "Test Album",
    }
    
    np.Update(track)
    
    result := np.Get()
    if result.TrackID != track.ID {
        t.Errorf("Expected TrackID %s, got %s", track.ID, result.TrackID)
    }
}
```

### Integration Test Example

```go
func TestAPI_NowPlayingWorkflow(t *testing.T) {
    server := setupTestServer()
    defer server.Close()
    
    // Test POST
    track := models.SpotifyTrack{/* ... */}
    body, _ := json.Marshal(track)
    resp, err := http.Post(server.URL+"/api/now-playing", "application/json", bytes.NewBuffer(body))
    
    // Verify response
    if resp.StatusCode != http.StatusOK {
        t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
    }
}
```

### Mock Usage Example

```go
func TestWithMock(t *testing.T) {
    mockGenius := &mocks.MockGeniusService{
        GetLyricsFunc: func(trackName, artistName string) (string, error) {
            return "Test lyrics", nil
        },
    }
    
    repo := repositories.NewMusicRepository(mockGenius)
    // Test repository with predictable mock behavior
}
```

## Test Best Practices

### 1. Test Naming
- Use descriptive names: `TestNowPlaying_Update_SetsCorrectValues`
- Include the scenario: `TestHandler_InvalidJSON_ReturnsBadRequest`

### 2. Test Structure
```go
func TestFunction_Scenario_ExpectedBehavior(t *testing.T) {
    // Arrange - Set up test data
    
    // Act - Execute the function
    
    // Assert - Verify the results
}
```

### 3. Error Testing
Always test both success and failure scenarios:
```go
func TestFunction_Success(t *testing.T) { /* ... */ }
func TestFunction_Error(t *testing.T) { /* ... */ }
```

### 4. Concurrent Testing
Test thread safety for concurrent code:
```go
func TestConcurrentAccess(t *testing.T) {
    done := make(chan bool, 10)
    
    for i := 0; i < 5; i++ {
        go func() {
            // Concurrent operation
            done <- true
        }()
    }
    
    // Wait for completion with timeout
    for i := 0; i < 5; i++ {
        select {
        case <-done:
        case <-time.After(time.Second):
            t.Fatal("Timeout")
        }
    }
}
```

## Continuous Integration

Tests are designed to run in CI/CD environments:

### GitHub Actions Example
```yaml
- name: Run Tests
  run: |
    go test -v -race -coverprofile=coverage.out ./tests/...
    go tool cover -html=coverage.out -o coverage.html
```

### Test Environment Variables
For integration tests that need external services, use environment variables:
```go
if testing.Short() {
    t.Skip("Skipping integration test in short mode")
}
```

## Coverage Goals

- **Unit Tests**: Aim for >90% coverage on business logic
- **Integration Tests**: Cover all API endpoints and major workflows
- **Critical Paths**: 100% coverage on security and data integrity functions

## Performance Testing

Use benchmarks for performance-critical code:
```go
func BenchmarkFunction(b *testing.B) {
    for i := 0; i < b.N; i++ {
        // Function to benchmark
    }
}
```

Run with:
```bash
make benchmark
```

## Debugging Tests

### Verbose Output
```bash
go test -v ./tests/unit/models/
```

### Run Specific Test
```bash
go test -run TestNowPlaying_Update ./tests/unit/models/
```

### Test with Race Detection
```bash
go test -race ./tests/...
```

## Mock Development

When adding new services, create corresponding mocks:

1. Implement the service interface
2. Add configurable function fields
3. Provide default behavior
4. Document mock behavior in tests

This ensures consistent, reliable testing across the entire application.