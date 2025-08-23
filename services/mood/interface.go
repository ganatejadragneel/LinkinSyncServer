package mood

import (
	"backend/server/models"
)

// Service defines the interface for mood analysis and song matching operations
type Service interface {
	// DetectMood analyzes user message for emotional content
	DetectMood(message string) (*models.MoodAnalysis, error)
	
	// MatchSongsToMood finds songs that match the detected mood
	MatchSongsToMood(mood *models.MoodAnalysis, userTracks []models.UnifiedTrack, limit int) ([]models.MoodBasedRecommendation, error)
	
	// GetLyricsWithMood fetches lyrics and analyzes their mood
	GetLyricsWithMood(trackName, artistName string) (*LyricsWithMood, error)
	
	// SaveUserMoodHistory saves user's mood and played songs to history file
	SaveUserMoodHistory(userID string, mood string, playedSongs []string) error
	
	// GetUserMoodHistory retrieves user's mood history
	GetUserMoodHistory(userID string) ([]UserMoodEntry, error)
}

// LyricsWithMood represents lyrics with mood analysis
type LyricsWithMood struct {
	Lyrics       string
	MoodAnalysis *models.MoodAnalysis
	Themes       []string
}

// UserMoodEntry represents a single mood history entry
type UserMoodEntry struct {
	Timestamp    string   `json:"timestamp"`
	DetectedMood string   `json:"detected_mood"`
	PlayedSongs  []string `json:"played_songs"`
}

// MoodKeywords defines keywords associated with different moods
var MoodKeywords = map[string][]string{
	"sad": {"cry", "tears", "broken", "hurt", "pain", "lost", "miss", "gone", "alone", "empty"},
	"happy": {"joy", "smile", "laugh", "bright", "sunshine", "celebrate", "love", "wonderful", "amazing", "blessed"},
	"angry": {"rage", "fury", "hate", "mad", "pissed", "scream", "fight", "burn", "destroy", "revenge"},
	"lonely": {"alone", "nobody", "isolated", "forgotten", "abandoned", "solitary", "empty", "belong", "disconnected"},
	"anxious": {"worry", "fear", "nervous", "panic", "stress", "overwhelmed", "restless", "uncertain", "doubt"},
	"nostalgic": {"remember", "memories", "past", "used to", "once", "old days", "reminisce", "looking back", "childhood"},
	"energetic": {"pump", "hype", "energy", "power", "strength", "unstoppable", "fire", "ready", "go", "motivation"},
	"calm": {"peace", "quiet", "serene", "tranquil", "relax", "breathe", "gentle", "soft", "still", "harmony"},
}

// RelatedMoods defines moods that are similar or related
var RelatedMoods = map[string][]string{
	"sad": {"melancholic", "depressed", "sorrowful", "grief"},
	"happy": {"joyful", "excited", "cheerful", "elated"},
	"angry": {"frustrated", "bitter", "defiant", "aggressive"},
	"lonely": {"isolated", "disconnected", "yearning", "longing"},
	"anxious": {"worried", "tense", "uneasy", "stressed"},
	"nostalgic": {"sentimental", "wistful", "reflective", "bittersweet"},
	"energetic": {"pumped", "motivated", "dynamic", "vigorous"},
	"calm": {"peaceful", "relaxed", "meditative", "zen"},
}