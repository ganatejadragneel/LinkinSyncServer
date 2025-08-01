package models

import (
	"sync"
	"time"
)

// PlayHistoryItem represents a single song in play history
type PlayHistoryItem struct {
	TrackID   string    `json:"track_id"`
	TrackName string    `json:"track_name"`
	Artist    string    `json:"artist"`
	Album     string    `json:"album"`
	Source    string    `json:"source,omitempty"`
	PlayedAt  time.Time `json:"played_at"`
}

// PlayHistory stores recently played tracks
type PlayHistory struct {
	items []PlayHistoryItem
	mutex sync.RWMutex
	maxItems int
}

// NewPlayHistory creates a new PlayHistory instance
func NewPlayHistory(maxItems int) *PlayHistory {
	return &PlayHistory{
		items: make([]PlayHistoryItem, 0),
		maxItems: maxItems,
	}
}

// Add adds a new track from SpotifyTrack to the history
func (ph *PlayHistory) Add(track SpotifyTrack) {
	ph.AddUnified(FromSpotifyTrack(track))
}

// AddUnified adds a new UnifiedTrack to the history
func (ph *PlayHistory) AddUnified(track UnifiedTrack) {
	ph.mutex.Lock()
	defer ph.mutex.Unlock()
	
	item := PlayHistoryItem{
		TrackID:   track.ID,
		TrackName: track.Name,
		Artist:    track.Artist,
		Album:     track.Album,
		Source:    track.Source,
		PlayedAt:  time.Now(),
	}
	
	// Add to beginning
	ph.items = append([]PlayHistoryItem{item}, ph.items...)
	
	// Keep only maxItems
	if len(ph.items) > ph.maxItems {
		ph.items = ph.items[:ph.maxItems]
	}
}

// GetItems returns a copy of all history items
func (ph *PlayHistory) GetItems() []PlayHistoryItem {
	ph.mutex.RLock()
	defer ph.mutex.RUnlock()
	
	// Return a copy to prevent external modification
	itemsCopy := make([]PlayHistoryItem, len(ph.items))
	copy(itemsCopy, ph.items)
	return itemsCopy
}