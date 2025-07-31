package models

import (
	"sync"
	"time"
)

// NowPlaying represents the currently playing song
type NowPlaying struct {
	TrackID   string `json:"track_id"`
	TrackName string `json:"track_name"`
	Artist    string `json:"artist"`
	Album     string `json:"album"`
	Lyrics    string `json:"lyrics,omitempty"`
	UpdatedAt time.Time `json:"updated_at"`
	mutex     sync.RWMutex
}

// NewNowPlaying creates a new NowPlaying instance
func NewNowPlaying() *NowPlaying {
	return &NowPlaying{}
}

// Update safely updates the currently playing track
func (np *NowPlaying) Update(track SpotifyTrack) {
	np.mutex.Lock()
	defer np.mutex.Unlock()
	
	np.TrackID = track.ID
	np.TrackName = track.Name
	np.Artist = track.Artist
	np.Album = track.Album
	np.Lyrics = "" // Reset lyrics for new track
	np.UpdatedAt = time.Now()
}

// UpdateLyrics safely updates the lyrics
func (np *NowPlaying) UpdateLyrics(lyrics string) {
	np.mutex.Lock()
	defer np.mutex.Unlock()
	np.Lyrics = lyrics
}

// Get safely returns a copy of the current playing track
func (np *NowPlaying) Get() NowPlaying {
	np.mutex.RLock()
	defer np.mutex.RUnlock()
	
	return NowPlaying{
		TrackID:   np.TrackID,
		TrackName: np.TrackName,
		Artist:    np.Artist,
		Album:     np.Album,
		Lyrics:    np.Lyrics,
		UpdatedAt: np.UpdatedAt,
	}
}

// IsEmpty checks if there's no song currently playing
func (np *NowPlaying) IsEmpty() bool {
	np.mutex.RLock()
	defer np.mutex.RUnlock()
	
	return np.TrackID == "" || np.TrackName == ""
}

// GetInfo returns formatted song information
func (np *NowPlaying) GetInfo() string {
	np.mutex.RLock()
	defer np.mutex.RUnlock()
	
	if np.TrackName == "" || np.Artist == "" {
		return ""
	}
	return np.TrackName + " by " + np.Artist
}