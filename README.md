# LinkinSync Backend

LinkinSync Backend is a Go-based server that powers a music-focused social platform with real-time chat and AI-powered lyrics analysis.

## Features

- **Global Chat**: Real-time messaging system for users to connect and discuss music
- **Lyrics Analysis**: AI-powered analysis of song lyrics using Ollama LLM
- **Music Information Management**: Track currently playing songs and maintain play history
- **Spotify Integration**: Get song information via Spotify API
- **Genius Integration**: Fetch and analyze lyrics via Genius API

## Tech Stack

- **Go**: Primary programming language
- **PostgreSQL**: Database for storing chat messages
- **Gorilla/Mux**: HTTP router and dispatcher
- **Ollama**: Local LLM for AI-powered lyrics analysis
- **Third-Party APIs**:
    - Spotify API for music information
    - Genius API for lyrics data

## API Endpoints

### Global Chat
- `GET /api/messages`: Fetch all chat messages
- `POST /api/messages`: Post a new chat message

### Music and Lyrics
- `POST /api/now-playing`: Update the currently playing song
- `GET /api/now-playing`: Get details of the currently playing song
- `GET /api/history`: Get the recent playback history
- `POST /api/chat`: Send a query about lyrics to the AI assistant

## Setup Instructions

### Prerequisites
- Go 1.19 or higher
- PostgreSQL
- Ollama (with llama2 model)
- Spotify Developer Account
- Genius API Client

### Configuration

1. Clone the repository
   ```bash
   git clone <repository-url>
   cd linkin-sync/backend
   ```

2. Create a `.env` file in the root directory with the following variables:
   ```
   # Database settings
   DB_HOST=localhost
   DB_PORT=5432
   DB_USER=your_db_user
   DB_PASSWORD=your_db_password
   DB_NAME=linkinsync

   # Server settings
   PORT=8080

   # Spotify API credentials
   SPOTIFY_CLIENT_ID=your_spotify_client_id
   SPOTIFY_CLIENT_SECRET=your_spotify_client_secret

   # Genius API credentials
   GENIUS_ACCESS_TOKEN=your_genius_access_token
   ```

3. Install Ollama and the llama2 model
   ```bash
   # Install Ollama from https://ollama.ai/
   # Then pull the llama2 model
   ollama pull llama2
   ```

4. Create the PostgreSQL database
   ```sql
   CREATE DATABASE linkinsync;
   ```

### Running the Server

1. Install dependencies
   ```bash
   go mod download
   ```

2. Start the server
   ```bash
   go run server/main.go
   ```

3. The server will be available at http://localhost:8080

## Database Schema

### Global Messages Table
```sql
CREATE TABLE global_messages (
    id SERIAL PRIMARY KEY,
    user_email VARCHAR(255) NOT NULL,
    username VARCHAR(255) NOT NULL,
    message_text TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL
);
```

## Deployment

For production deployment:

1. Build the binary
   ```bash
   go build -o linkinsync-server server/main.go
   ```

2. Run the binary
   ```bash
   ./linkinsync-server
   ```

3. For proper production setup, consider using a process manager like systemd or PM2.

## License

[MIT License](LICENSE)
