package mocks

import (
	"backend/server/models"
	"backend/services/mood"
)

// MockMoodService implements mood.Service for testing
type MockMoodService struct {
	DetectMoodFunc       func(message string) (*models.MoodAnalysis, error)
	MatchSongsToMoodFunc func(moodAnalysis *models.MoodAnalysis, userTracks []models.UnifiedTrack, limit int) ([]models.MoodBasedRecommendation, error)
	GetLyricsWithMoodFunc func(trackName, artistName string) (*mood.LyricsWithMood, error)
	SaveUserMoodHistoryFunc func(userID string, mood string, playedSongs []string) error
	GetUserMoodHistoryFunc func(userID string) ([]mood.UserMoodEntry, error)
}

// Ensure MockMoodService implements mood.Service
var _ mood.Service = (*MockMoodService)(nil)

// DetectMood calls the mock function if set, otherwise returns default values
func (m *MockMoodService) DetectMood(message string) (*models.MoodAnalysis, error) {
	if m.DetectMoodFunc != nil {
		return m.DetectMoodFunc(message)
	}
	return &models.MoodAnalysis{
		PrimaryMood:  "happy",
		MoodScore:    0.8,
		EmotionTags:  []string{"energetic", "positive"},
	}, nil
}

// MatchSongsToMood calls the mock function if set, otherwise returns default values
func (m *MockMoodService) MatchSongsToMood(moodAnalysis *models.MoodAnalysis, userTracks []models.UnifiedTrack, limit int) ([]models.MoodBasedRecommendation, error) {
	if m.MatchSongsToMoodFunc != nil {
		return m.MatchSongsToMoodFunc(moodAnalysis, userTracks, limit)
	}
	return []models.MoodBasedRecommendation{}, nil
}

// GetLyricsWithMood calls the mock function if set, otherwise returns default values
func (m *MockMoodService) GetLyricsWithMood(trackName, artistName string) (*mood.LyricsWithMood, error) {
	if m.GetLyricsWithMoodFunc != nil {
		return m.GetLyricsWithMoodFunc(trackName, artistName)
	}
	return &mood.LyricsWithMood{
		Lyrics: "Mock lyrics for " + trackName,
		MoodAnalysis: &models.MoodAnalysis{
			PrimaryMood: "happy",
			MoodScore:   0.8,
			EmotionTags: []string{"positive"},
		},
		Themes: []string{"love", "happiness"},
	}, nil
}

// SaveUserMoodHistory calls the mock function if set, otherwise returns nil
func (m *MockMoodService) SaveUserMoodHistory(userID string, mood string, playedSongs []string) error {
	if m.SaveUserMoodHistoryFunc != nil {
		return m.SaveUserMoodHistoryFunc(userID, mood, playedSongs)
	}
	return nil
}

// GetUserMoodHistory calls the mock function if set, otherwise returns empty slice
func (m *MockMoodService) GetUserMoodHistory(userID string) ([]mood.UserMoodEntry, error) {
	if m.GetUserMoodHistoryFunc != nil {
		return m.GetUserMoodHistoryFunc(userID)
	}
	return []mood.UserMoodEntry{}, nil
}