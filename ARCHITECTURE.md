# Backend Architecture

## Overview

The backend has been refactored to follow SOLID principles and maintain a clean, modular architecture.

## Directory Structure

```
backend/
├── config/           # Configuration management
├── middleware/       # HTTP middleware (logging, recovery, etc.)
├── repositories/     # Data access layer
├── server/
│   ├── database/     # Database connection and setup
│   ├── handlers/     # HTTP request handlers
│   ├── models/       # Data models and structures
│   └── main.go       # Application entry point
├── services/         # Business logic services
│   ├── genius/       # Genius API integration for lyrics
│   ├── ollama/       # Ollama AI integration
│   └── spotify/      # Spotify API integration
└── utils/            # Utility functions

```

## Key Principles Applied

### 1. Single Responsibility Principle (SRP)
- Each package has a single, well-defined purpose
- Services handle specific external integrations
- Handlers only deal with HTTP request/response
- Repositories manage data persistence

### 2. Open/Closed Principle (OCP)
- Services are defined through interfaces
- New implementations can be added without modifying existing code
- Middleware can be chained without changing core logic

### 3. Liskov Substitution Principle (LSP)
- All services implement their respective interfaces
- Services can be swapped out for testing or different implementations

### 4. Interface Segregation Principle (ISP)
- Small, focused interfaces for each service
- Clients only depend on methods they use

### 5. Dependency Inversion Principle (DIP)
- Handlers depend on service interfaces, not concrete implementations
- Dependencies are injected through constructors

## Key Components

### Services

#### Spotify Service
- Handles Spotify API authentication
- Fetches track information
- Token management with automatic refresh

#### Genius Service
- Searches for songs on Genius
- Scrapes lyrics from Genius pages
- Handles artist matching for accurate results

#### Ollama Service
- Integrates with local Ollama instance
- Provides lyrics analysis
- Handles general chat queries

### Repositories

#### Music Repository
- Manages current playing track state
- Maintains play history
- Caches lyrics for performance
- Thread-safe operations

### Handlers

#### Lyrics Handler
- Manages all music-related endpoints
- Uses dependency injection for services
- Handles chat interactions about music

#### Chat Handler
- Manages global chat functionality
- Database persistence for messages

### Middleware

- **Logging**: Logs all HTTP requests with timing
- **Recovery**: Catches panics and returns proper error responses

## Configuration

All configuration is managed through environment variables:

```env
# Server
PORT=8080

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=your_user
DB_PASSWORD=your_password
DB_NAME=your_database
DB_SSL_MODE=disable

# Spotify API
SPOTIFY_CLIENT_ID=your_client_id
SPOTIFY_CLIENT_SECRET=your_client_secret

# Genius API
GENIUS_ACCESS_TOKEN=your_access_token

# Ollama
OLLAMA_BASE_URL=http://localhost:11434
OLLAMA_MODEL=llama3.2:3b
```

## Benefits of This Architecture

1. **Testability**: Each component can be tested in isolation
2. **Maintainability**: Clear separation of concerns
3. **Extensibility**: Easy to add new services or handlers
4. **Reusability**: Services can be used across different handlers
5. **Performance**: Built-in caching and efficient state management
6. **Error Handling**: Consistent error handling across the application
7. **Logging**: Comprehensive request logging for debugging