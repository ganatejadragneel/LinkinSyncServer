package models_test

import (
	"backend/server/models"
	"testing"
	"time"
)

func TestNewPlayHistory(t *testing.T) {
	maxItems := 5
	ph := models.NewPlayHistory(maxItems)
	
	if ph == nil {
		t.Fatal("NewPlayHistory should not return nil")
	}
	
	items := ph.GetItems()
	if len(items) != 0 {
		t.Error("New PlayHistory should have no items initially")
	}
}

func TestPlayHistory_Add(t *testing.T) {
	ph := models.NewPlayHistory(3)
	
	track1 := models.SpotifyTrack{
		ID:     "track1",
		Name:   "Song 1",
		Artist: "Artist 1",
		Album:  "Album 1",
	}
	
	track2 := models.SpotifyTrack{
		ID:     "track2",
		Name:   "Song 2",
		Artist: "Artist 2",
		Album:  "Album 2",
	}
	
	ph.Add(track1)
	items := ph.GetItems()
	
	if len(items) != 1 {
		t.Errorf("Expected 1 item, got %d", len(items))
	}
	
	if items[0].TrackID != track1.ID {
		t.Errorf("Expected TrackID %s, got %s", track1.ID, items[0].TrackID)
	}
	
	if items[0].PlayedAt.IsZero() {
		t.Error("PlayedAt should be set")
	}
	
	ph.Add(track2)
	items = ph.GetItems()
	
	if len(items) != 2 {
		t.Errorf("Expected 2 items, got %d", len(items))
	}
	
	// Most recent should be first
	if items[0].TrackID != track2.ID {
		t.Errorf("Expected first item to be %s, got %s", track2.ID, items[0].TrackID)
	}
	if items[1].TrackID != track1.ID {
		t.Errorf("Expected second item to be %s, got %s", track1.ID, items[1].TrackID)
	}
}

func TestPlayHistory_MaxItems(t *testing.T) {
	maxItems := 2
	ph := models.NewPlayHistory(maxItems)
	
	// Add more items than max
	for i := 1; i <= 5; i++ {
		track := models.SpotifyTrack{
			ID:     "track" + string(rune(i+'0')),
			Name:   "Song " + string(rune(i+'0')),
			Artist: "Artist " + string(rune(i+'0')),
			Album:  "Album " + string(rune(i+'0')),
		}
		ph.Add(track)
	}
	
	items := ph.GetItems()
	if len(items) != maxItems {
		t.Errorf("Expected %d items, got %d", maxItems, len(items))
	}
	
	// Should have the most recent items
	if items[0].TrackID != "track5" {
		t.Errorf("Expected first item to be track5, got %s", items[0].TrackID)
	}
	if items[1].TrackID != "track4" {
		t.Errorf("Expected second item to be track4, got %s", items[1].TrackID)
	}
}

func TestPlayHistory_GetItemsReturnsACopy(t *testing.T) {
	ph := models.NewPlayHistory(5)
	
	track := models.SpotifyTrack{
		ID:     "test123",
		Name:   "Test Song",
		Artist: "Test Artist",
		Album:  "Test Album",
	}
	
	ph.Add(track)
	items1 := ph.GetItems()
	items2 := ph.GetItems()
	
	// Modify one slice
	if len(items1) > 0 {
		items1[0].TrackName = "Modified"
	}
	
	// Other slice should be unaffected
	if items2[0].TrackName == "Modified" {
		t.Error("GetItems should return a copy, not the original slice")
	}
	
	if items2[0].TrackName != track.Name {
		t.Errorf("Expected TrackName %s, got %s", track.Name, items2[0].TrackName)
	}
}

func TestPlayHistory_ConcurrentAccess(t *testing.T) {
	ph := models.NewPlayHistory(10)
	
	done := make(chan bool, 10)
	
	// Start 5 goroutines adding items
	for i := 0; i < 5; i++ {
		go func(i int) {
			track := models.SpotifyTrack{
				ID:     "track" + string(rune(i+'0')),
				Name:   "Song " + string(rune(i+'0')),
				Artist: "Artist " + string(rune(i+'0')),
				Album:  "Album " + string(rune(i+'0')),
			}
			ph.Add(track)
			done <- true
		}(i)
	}
	
	// Start 5 goroutines reading items
	for i := 0; i < 5; i++ {
		go func() {
			_ = ph.GetItems()
			done <- true
		}()
	}
	
	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("Timeout waiting for goroutines")
		}
	}
	
	// Verify we have some items
	items := ph.GetItems()
	if len(items) == 0 {
		t.Error("Should have some items after concurrent operations")
	}
}