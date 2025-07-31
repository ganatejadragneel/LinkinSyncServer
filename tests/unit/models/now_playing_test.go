package models_test

import (
	"backend/server/models"
	"testing"
	"time"
)

func TestNewNowPlaying(t *testing.T) {
	np := models.NewNowPlaying()
	
	if np == nil {
		t.Fatal("NewNowPlaying should not return nil")
	}
	
	if !np.IsEmpty() {
		t.Error("New NowPlaying should be empty initially")
	}
}

func TestNowPlaying_Update(t *testing.T) {
	np := models.NewNowPlaying()
	track := models.SpotifyTrack{
		ID:     "test123",
		Name:   "Test Song",
		Artist: "Test Artist",
		Album:  "Test Album",
	}
	
	np.Update(track)
	
	result := np.Get()
	if result.TrackID != track.ID {
		t.Errorf("Expected TrackID %s, got %s", track.ID, result.TrackID)
	}
	if result.TrackName != track.Name {
		t.Errorf("Expected TrackName %s, got %s", track.Name, result.TrackName)
	}
	if result.Artist != track.Artist {
		t.Errorf("Expected Artist %s, got %s", track.Artist, result.Artist)
	}
	if result.Album != track.Album {
		t.Errorf("Expected Album %s, got %s", track.Album, result.Album)
	}
	if result.Lyrics != "" {
		t.Error("Lyrics should be empty after update")
	}
	if result.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should be set")
	}
}

func TestNowPlaying_UpdateLyrics(t *testing.T) {
	np := models.NewNowPlaying()
	track := models.SpotifyTrack{
		ID:     "test123",
		Name:   "Test Song",
		Artist: "Test Artist",
		Album:  "Test Album",
	}
	
	np.Update(track)
	lyrics := "Test lyrics content"
	np.UpdateLyrics(lyrics)
	
	result := np.Get()
	if result.Lyrics != lyrics {
		t.Errorf("Expected Lyrics %s, got %s", lyrics, result.Lyrics)
	}
}

func TestNowPlaying_IsEmpty(t *testing.T) {
	np := models.NewNowPlaying()
	
	// Should be empty initially
	if !np.IsEmpty() {
		t.Error("New NowPlaying should be empty")
	}
	
	// Should not be empty after update
	track := models.SpotifyTrack{
		ID:     "test123",
		Name:   "Test Song",
		Artist: "Test Artist",
		Album:  "Test Album",
	}
	np.Update(track)
	
	if np.IsEmpty() {
		t.Error("NowPlaying should not be empty after update")
	}
}

func TestNowPlaying_GetInfo(t *testing.T) {
	np := models.NewNowPlaying()
	
	// Should return empty string initially
	if np.GetInfo() != "" {
		t.Error("GetInfo should return empty string for empty NowPlaying")
	}
	
	// Should return formatted string after update
	track := models.SpotifyTrack{
		ID:     "test123",
		Name:   "Test Song",
		Artist: "Test Artist",
		Album:  "Test Album",
	}
	np.Update(track)
	
	expected := "Test Song by Test Artist"
	if np.GetInfo() != expected {
		t.Errorf("Expected GetInfo %s, got %s", expected, np.GetInfo())
	}
}

func TestNowPlaying_ConcurrentAccess(t *testing.T) {
	np := models.NewNowPlaying()
	track := models.SpotifyTrack{
		ID:     "test123",
		Name:   "Test Song",
		Artist: "Test Artist",
		Album:  "Test Album",
	}
	
	// Test concurrent reads and writes
	done := make(chan bool, 10)
	
	// Start 5 goroutines updating
	for i := 0; i < 5; i++ {
		go func(i int) {
			track.ID = "test" + string(rune(i+'0'))
			np.Update(track)
			done <- true
		}(i)
	}
	
	// Start 5 goroutines reading
	for i := 0; i < 5; i++ {
		go func() {
			_ = np.Get()
			_ = np.IsEmpty()
			_ = np.GetInfo()
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
}