package main

import (
	"backend/config"
	"backend/middleware"
	"backend/repositories"
	"backend/server/database"
	"backend/server/handlers"
	"backend/services/genius"
	"backend/services/ollama"
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	// Initialize database
	db, err := database.InitDB()
	if err != nil {
		log.Fatal("Database connection failed:", err)
	}
	defer db.Close()

	// Setup database tables
	if err := setupDatabase(db); err != nil {
		log.Fatal("Error setting up database:", err)
	}

	// Initialize services
	geniusService := genius.New(genius.Config{
		AccessToken: cfg.Genius.AccessToken,
	})

	// Note: Spotify service can be initialized here if needed for future features
	// spotifyService := spotify.New(spotify.Config{
	//     ClientID:     cfg.Spotify.ClientID,
	//     ClientSecret: cfg.Spotify.ClientSecret,
	// })

	ollamaService := ollama.New(ollama.Config{
		BaseURL:     cfg.Ollama.BaseURL,
		Model:       cfg.Ollama.Model,
		Temperature: cfg.Ollama.Temperature,
		TopP:        cfg.Ollama.TopP,
		TopK:        cfg.Ollama.TopK,
	})

	// Check if Ollama is available
	if err := ollamaService.IsAvailable(); err != nil {
		log.Fatalf("Error connecting to Ollama: %v - Make sure Ollama is running with 'ollama serve'", err)
	}
	log.Println("Successfully connected to Ollama server")

	// Initialize repositories
	musicRepo := repositories.NewMusicRepository(geniusService)

	// Initialize handlers
	lyricsHandler := handlers.NewLyricsHandler(musicRepo, ollamaService)
	chatHandler := handlers.NewChatHandler(db)

	// Setup routes
	router := setupRoutes(lyricsHandler, chatHandler)

	// Apply middleware
	handler := middleware.Recovery(middleware.Logging(router))

	// Setup CORS
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"http://localhost:3000", "http://127.0.0.1:3000"},
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type", "Authorization"},
	})

	// Start server
	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	log.Printf("Server starting on %s", addr)
	log.Printf("Make sure Ollama is running: ollama serve")
	log.Printf("Make sure you have the model: ollama pull %s", cfg.Ollama.Model)
	
	if err := http.ListenAndServe(addr, c.Handler(handler)); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}

// setupRoutes configures all HTTP routes
func setupRoutes(lyricsHandler *handlers.LyricsHandler, chatHandler *handlers.ChatHandler) *mux.Router {
	r := mux.NewRouter()

	// API routes
	api := r.PathPrefix("/api").Subrouter()

	// Global chat routes
	api.HandleFunc("/messages", chatHandler.GetMessages).Methods("GET")
	api.HandleFunc("/messages", chatHandler.PostMessage).Methods("POST")

	// Music and lyrics routes
	api.HandleFunc("/now-playing", lyricsHandler.UpdateNowPlaying).Methods("POST")
	api.HandleFunc("/now-playing", lyricsHandler.GetNowPlaying).Methods("GET")
	api.HandleFunc("/history", lyricsHandler.GetPlayHistory).Methods("GET")
	api.HandleFunc("/chat", lyricsHandler.HandleChat).Methods("POST")

	// Health check
	api.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "OK")
	}).Methods("GET")

	return r
}

// setupDatabase creates the necessary database tables
func setupDatabase(db *sql.DB) error {
	// Create table for global messages if it doesn't exist
	query := `
        CREATE TABLE IF NOT EXISTS global_messages (
            id SERIAL PRIMARY KEY,
            user_email VARCHAR(255) NOT NULL,
            username VARCHAR(255) NOT NULL,
            message_text TEXT NOT NULL,
            created_at TIMESTAMP WITH TIME ZONE NOT NULL
        );

		-- Create indexes for better performance
		CREATE INDEX IF NOT EXISTS idx_global_messages_created_at ON global_messages(created_at DESC);
		CREATE INDEX IF NOT EXISTS idx_global_messages_user_email ON global_messages(user_email);
    `
	
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	log.Println("Database tables set up successfully")
	return nil
}