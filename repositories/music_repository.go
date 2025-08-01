package repositories

import (
	"backend/server/models"
	"backend/services/genius"
	"fmt"
)

// MusicRepository manages music-related data
type MusicRepository struct {
	nowPlaying   *models.NowPlaying
	playHistory  *models.PlayHistory
	lyricsCache  map[string]string // Simple in-memory cache for lyrics
	geniusService genius.Service
}

// NewMusicRepository creates a new music repository
func NewMusicRepository(geniusService genius.Service) *MusicRepository {
	return &MusicRepository{
		nowPlaying:    models.NewNowPlaying(),
		playHistory:   models.NewPlayHistory(10), // Keep last 10 tracks
		lyricsCache:   make(map[string]string),
		geniusService: geniusService,
	}
}

// UpdateNowPlaying updates the currently playing track from SpotifyTrack
func (r *MusicRepository) UpdateNowPlaying(track models.SpotifyTrack) {
	r.nowPlaying.Update(track)
	r.playHistory.Add(track)
}

// UpdateNowPlayingUnified updates the currently playing track from UnifiedTrack
func (r *MusicRepository) UpdateNowPlayingUnified(track models.UnifiedTrack) {
	r.nowPlaying.UpdateUnified(track)
	r.playHistory.AddUnified(track)
}

// GetNowPlaying returns the currently playing track
func (r *MusicRepository) GetNowPlaying() models.NowPlaying {
	return r.nowPlaying.Get()
}

// IsPlaying checks if a track is currently playing
func (r *MusicRepository) IsPlaying() bool {
	return !r.nowPlaying.IsEmpty()
}

// GetPlayHistory returns the play history
func (r *MusicRepository) GetPlayHistory() []models.PlayHistoryItem {
	return r.playHistory.GetItems()
}

// GetLyricsForCurrentSong fetches lyrics for the current song
func (r *MusicRepository) GetLyricsForCurrentSong() (string, error) {
	current := r.nowPlaying.Get()
	
	// Check if we have a current song
	if current.TrackName == "" || current.Artist == "" {
		return "", fmt.Errorf("no song is currently playing")
	}

	// Check if we already have lyrics cached
	if current.Lyrics != "" {
		return current.Lyrics, nil
	}

	// Check memory cache
	cacheKey := fmt.Sprintf("%s|%s", current.TrackName, current.Artist)
	if lyrics, ok := r.lyricsCache[cacheKey]; ok {
		r.nowPlaying.UpdateLyrics(lyrics)
		return lyrics, nil
	}

	// Fetch lyrics from Genius
	lyrics, err := r.geniusService.GetLyrics(current.TrackName, current.Artist)
	if err != nil {
		return "", fmt.Errorf("failed to fetch lyrics: %w", err)
	}

	// Cache the lyrics
	r.lyricsCache[cacheKey] = lyrics
	r.nowPlaying.UpdateLyrics(lyrics)

	return lyrics, nil
}

// GetCurrentSongInfo returns formatted information about the current song
func (r *MusicRepository) GetCurrentSongInfo() string {
	return r.nowPlaying.GetInfo()
}