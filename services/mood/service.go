package mood

import (
	"backend/server/models"
	"backend/services/genius"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// AIService defines the interface for AI services (both Ollama and OpenAI)
type AIService interface {
	GenerateResponse(prompt string) (string, error)
}

// service implements the Mood Service interface
type service struct {
	geniusService genius.Service
	aiService     AIService  // Can be either Ollama or OpenAI
	lyricsCache   map[string]*LyricsWithMood
	cacheMutex    sync.RWMutex
	dataDir       string
}

// New creates a new Mood service
func New(geniusService genius.Service, aiService AIService, dataDir string) Service {
	// Create data directory if it doesn't exist
	moodHistoryDir := filepath.Join(dataDir, "mood_history")
	os.MkdirAll(moodHistoryDir, 0755)
	
	return &service{
		geniusService: geniusService,
		aiService:     aiService,
		lyricsCache:   make(map[string]*LyricsWithMood),
		dataDir:       dataDir,
	}
}

// DetectMood analyzes user message for emotional content
func (s *service) DetectMood(message string) (*models.MoodAnalysis, error) {
	// Create mood detection prompt
	prompt := fmt.Sprintf(`Analyze the following message for emotional content and mood. Return a JSON response with:
- primary_mood: The main emotion detected (must be one of: sad, happy, angry, lonely, anxious, nostalgic, energetic, calm)
- mood_score: Confidence score between 0 and 1
- emotion_tags: Array of related emotions/themes

Important: Respond ONLY with valid JSON, no additional text.

User message: "%s"`, message)

	// Get response from AI service
	response, err := s.aiService.GenerateResponse(prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to detect mood: %w", err)
	}

	// Parse JSON response
	var moodAnalysis models.MoodAnalysis
	if err := json.Unmarshal([]byte(response), &moodAnalysis); err != nil {
		// If JSON parsing fails, try to extract mood manually
		return s.fallbackMoodDetection(message), nil
	}

	return &moodAnalysis, nil
}

// fallbackMoodDetection provides basic mood detection if AI fails
func (s *service) fallbackMoodDetection(message string) *models.MoodAnalysis {
	lowerMessage := strings.ToLower(message)
	
	// Check for mood keywords
	for mood, keywords := range MoodKeywords {
		matchCount := 0
		for _, keyword := range keywords {
			if strings.Contains(lowerMessage, keyword) {
				matchCount++
			}
		}
		
		if matchCount > 0 {
			score := float64(matchCount) / float64(len(keywords))
			if score > 0.2 { // If at least 20% of keywords match
				return &models.MoodAnalysis{
					PrimaryMood: mood,
					MoodScore:   score,
					EmotionTags: RelatedMoods[mood],
				}
			}
		}
	}
	
	// Default to neutral/uncertain
	return &models.MoodAnalysis{
		PrimaryMood: "calm",
		MoodScore:   0.3,
		EmotionTags: []string{"neutral", "uncertain"},
	}
}

// MatchSongsToMood finds songs that match the detected mood
func (s *service) MatchSongsToMood(mood *models.MoodAnalysis, userTracks []models.UnifiedTrack, limit int) ([]models.MoodBasedRecommendation, error) {
	var recommendations []models.MoodBasedRecommendation
	var wg sync.WaitGroup
	var mutex sync.Mutex
	
	// Process tracks concurrently but with rate limiting
	semaphore := make(chan struct{}, 5) // Process max 5 tracks at a time
	
	for _, track := range userTracks {
		if len(recommendations) >= limit {
			break
		}
		
		wg.Add(1)
		go func(t models.UnifiedTrack) {
			defer wg.Done()
			
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			
			// Get lyrics and analyze mood
			lyricsData, err := s.GetLyricsWithMood(t.Name, t.Artist)
			if err != nil {
				return // Skip this track if we can't get lyrics
			}
			
			// Calculate mood match score
			matchScore := s.calculateMoodMatch(mood, lyricsData.MoodAnalysis)
			
			if matchScore > 0.5 { // Only include if match score is above threshold
				mutex.Lock()
				recommendations = append(recommendations, models.MoodBasedRecommendation{
					Track:       t,
					MoodScore:   matchScore,
					MatchReason: s.generateMatchReason(mood.PrimaryMood, lyricsData.Themes),
				})
				mutex.Unlock()
			}
		}(track)
	}
	
	wg.Wait()
	
	// Sort by mood score (highest first)
	// Simple bubble sort for now
	for i := 0; i < len(recommendations); i++ {
		for j := i + 1; j < len(recommendations); j++ {
			if recommendations[j].MoodScore > recommendations[i].MoodScore {
				recommendations[i], recommendations[j] = recommendations[j], recommendations[i]
			}
		}
	}
	
	// Return top matches up to limit
	if len(recommendations) > limit {
		recommendations = recommendations[:limit]
	}
	
	return recommendations, nil
}

// GetLyricsWithMood fetches lyrics and analyzes their mood
func (s *service) GetLyricsWithMood(trackName, artistName string) (*LyricsWithMood, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("%s-%s", strings.ToLower(trackName), strings.ToLower(artistName))
	
	s.cacheMutex.RLock()
	if cached, exists := s.lyricsCache[cacheKey]; exists {
		s.cacheMutex.RUnlock()
		return cached, nil
	}
	s.cacheMutex.RUnlock()
	
	// Fetch lyrics from Genius
	lyrics, err := s.geniusService.GetLyrics(trackName, artistName)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch lyrics: %w", err)
	}
	
	// Analyze mood of lyrics
	moodPrompt := fmt.Sprintf(`Analyze the mood and themes of these song lyrics. Return a JSON response with:
- primary_mood: The main emotion (must be one of: sad, happy, angry, lonely, anxious, nostalgic, energetic, calm)
- mood_score: Confidence score between 0 and 1
- emotion_tags: Array of related emotions
- themes: Array of main themes in the song

Important: Respond ONLY with valid JSON.

Lyrics:
%s`, lyrics)

	response, err := s.aiService.GenerateResponse(moodPrompt)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze lyrics mood: %w", err)
	}
	
	// Parse response
	var analysis struct {
		PrimaryMood string   `json:"primary_mood"`
		MoodScore   float64  `json:"mood_score"`
		EmotionTags []string `json:"emotion_tags"`
		Themes      []string `json:"themes"`
	}
	
	if err := json.Unmarshal([]byte(response), &analysis); err != nil {
		// Fallback analysis
		analysis.PrimaryMood = "calm"
		analysis.MoodScore = 0.5
		analysis.EmotionTags = []string{"uncertain"}
		analysis.Themes = []string{"general"}
	}
	
	result := &LyricsWithMood{
		Lyrics: lyrics,
		MoodAnalysis: &models.MoodAnalysis{
			PrimaryMood: analysis.PrimaryMood,
			MoodScore:   analysis.MoodScore,
			EmotionTags: analysis.EmotionTags,
		},
		Themes: analysis.Themes,
	}
	
	// Cache the result
	s.cacheMutex.Lock()
	s.lyricsCache[cacheKey] = result
	s.cacheMutex.Unlock()
	
	return result, nil
}

// calculateMoodMatch calculates how well two moods match
func (s *service) calculateMoodMatch(userMood, songMood *models.MoodAnalysis) float64 {
	// Direct mood match
	if userMood.PrimaryMood == songMood.PrimaryMood {
		return 0.9 + (songMood.MoodScore * 0.1) // 90-100% match
	}
	
	// Check if moods are related
	relatedMoods, exists := RelatedMoods[userMood.PrimaryMood]
	if exists {
		for _, related := range relatedMoods {
			if related == songMood.PrimaryMood {
				return 0.7 + (songMood.MoodScore * 0.2) // 70-90% match
			}
		}
	}
	
	// Check emotion tag overlap
	overlapCount := 0
	for _, userTag := range userMood.EmotionTags {
		for _, songTag := range songMood.EmotionTags {
			if userTag == songTag {
				overlapCount++
			}
		}
	}
	
	if overlapCount > 0 {
		overlapScore := float64(overlapCount) / float64(len(userMood.EmotionTags))
		return 0.5 + (overlapScore * 0.3) // 50-80% match based on overlap
	}
	
	return 0.2 // Low match
}

// generateMatchReason generates a reason why a song matches the mood
func (s *service) generateMatchReason(mood string, themes []string) string {
	themeStr := ""
	if len(themes) > 0 {
		themeStr = themes[0]
		if len(themes) > 1 {
			themeStr += " and " + themes[1]
		}
	}
	
	reasons := map[string]string{
		"sad":       "This song captures feelings of sadness and melancholy",
		"happy":     "This uplifting song matches your positive energy",
		"angry":     "This song channels frustration and intensity",
		"lonely":    "This song explores themes of isolation and longing",
		"anxious":   "This song reflects feelings of uncertainty and tension",
		"nostalgic": "This song brings back memories and reflection",
		"energetic": "This high-energy track matches your motivated mood",
		"calm":      "This peaceful song promotes relaxation and tranquility",
	}
	
	reason := reasons[mood]
	if themeStr != "" {
		reason += fmt.Sprintf(" through themes of %s", themeStr)
	}
	
	return reason
}

// SaveUserMoodHistory saves user's mood and played songs to history file
func (s *service) SaveUserMoodHistory(userID string, mood string, playedSongs []string) error {
	historyFile := filepath.Join(s.dataDir, "mood_history", fmt.Sprintf("user_%s_mood_history.txt", userID))
	
	// Create entry
	entry := fmt.Sprintf("%s|%s|%s\n", 
		time.Now().Format(time.RFC3339),
		mood,
		strings.Join(playedSongs, ","))
	
	// Append to file
	f, err := os.OpenFile(historyFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open history file: %w", err)
	}
	defer f.Close()
	
	if _, err := f.WriteString(entry); err != nil {
		return fmt.Errorf("failed to write history: %w", err)
	}
	
	// Check if compaction is needed (every 30 days)
	go s.compactHistoryIfNeeded(historyFile)
	
	return nil
}

// GetUserMoodHistory retrieves user's mood history
func (s *service) GetUserMoodHistory(userID string) ([]UserMoodEntry, error) {
	historyFile := filepath.Join(s.dataDir, "mood_history", fmt.Sprintf("user_%s_mood_history.txt", userID))
	
	// Check if file exists
	if _, err := os.Stat(historyFile); os.IsNotExist(err) {
		return []UserMoodEntry{}, nil
	}
	
	// Read file
	content, err := ioutil.ReadFile(historyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read history: %w", err)
	}
	
	// Parse entries
	var entries []UserMoodEntry
	lines := strings.Split(string(content), "\n")
	
	for _, line := range lines {
		if line == "" {
			continue
		}
		
		parts := strings.Split(line, "|")
		if len(parts) != 3 {
			continue
		}
		
		songs := []string{}
		if parts[2] != "" {
			songs = strings.Split(parts[2], ",")
		}
		
		entries = append(entries, UserMoodEntry{
			Timestamp:    parts[0],
			DetectedMood: parts[1],
			PlayedSongs:  songs,
		})
	}
	
	return entries, nil
}

// compactHistoryIfNeeded compacts history file if it's older than 30 days
func (s *service) compactHistoryIfNeeded(historyFile string) {
	// Implementation for 30-day compaction would go here
	// For now, this is a placeholder
}