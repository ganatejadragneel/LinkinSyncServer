package models_test

import (
	"backend/server/models"
	"testing"
)

func TestUnifiedTrack_ToSpotifyTrack(t *testing.T) {
	unifiedTrack := models.UnifiedTrack{
		ID:         "test123",
		Name:       "Test Song",
		Artist:     "Test Artist",
		Album:      "Test Album",
		Source:     "spotify",
		PreviewURL: "https://preview.url",
	}

	spotifyTrack := unifiedTrack.ToSpotifyTrack()

	if spotifyTrack.ID != unifiedTrack.ID {
		t.Errorf("Expected ID %s, got %s", unifiedTrack.ID, spotifyTrack.ID)
	}
	if spotifyTrack.Name != unifiedTrack.Name {
		t.Errorf("Expected Name %s, got %s", unifiedTrack.Name, spotifyTrack.Name)
	}
	if spotifyTrack.Artist != unifiedTrack.Artist {
		t.Errorf("Expected Artist %s, got %s", unifiedTrack.Artist, spotifyTrack.Artist)
	}
	if spotifyTrack.Album != unifiedTrack.Album {
		t.Errorf("Expected Album %s, got %s", unifiedTrack.Album, spotifyTrack.Album)
	}
	if spotifyTrack.PreviewURL != unifiedTrack.PreviewURL {
		t.Errorf("Expected PreviewURL %s, got %s", unifiedTrack.PreviewURL, spotifyTrack.PreviewURL)
	}
}

func TestFromSpotifyTrack(t *testing.T) {
	spotifyTrack := models.SpotifyTrack{
		ID:         "test123",
		Name:       "Test Song",
		Artist:     "Test Artist",
		Album:      "Test Album",
		PreviewURL: "https://preview.url",
	}

	unifiedTrack := models.FromSpotifyTrack(spotifyTrack)

	if unifiedTrack.ID != spotifyTrack.ID {
		t.Errorf("Expected ID %s, got %s", spotifyTrack.ID, unifiedTrack.ID)
	}
	if unifiedTrack.Name != spotifyTrack.Name {
		t.Errorf("Expected Name %s, got %s", spotifyTrack.Name, unifiedTrack.Name)
	}
	if unifiedTrack.Artist != spotifyTrack.Artist {
		t.Errorf("Expected Artist %s, got %s", spotifyTrack.Artist, unifiedTrack.Artist)
	}
	if unifiedTrack.Album != spotifyTrack.Album {
		t.Errorf("Expected Album %s, got %s", spotifyTrack.Album, unifiedTrack.Album)
	}
	if unifiedTrack.Source != "spotify" {
		t.Errorf("Expected Source spotify, got %s", unifiedTrack.Source)
	}
	if unifiedTrack.PreviewURL != spotifyTrack.PreviewURL {
		t.Errorf("Expected PreviewURL %s, got %s", spotifyTrack.PreviewURL, unifiedTrack.PreviewURL)
	}
}

func TestFromYouTubeTrack(t *testing.T) {
	id := "yt123"
	name := "YouTube Song"
	artist := "YouTube Artist"
	album := "YouTube Album"
	imageURL := "https://image.url"
	duration := 240

	unifiedTrack := models.FromYouTubeTrack(id, name, artist, album, imageURL, duration)

	if unifiedTrack.ID != id {
		t.Errorf("Expected ID %s, got %s", id, unifiedTrack.ID)
	}
	if unifiedTrack.Name != name {
		t.Errorf("Expected Name %s, got %s", name, unifiedTrack.Name)
	}
	if unifiedTrack.Artist != artist {
		t.Errorf("Expected Artist %s, got %s", artist, unifiedTrack.Artist)
	}
	if unifiedTrack.Album != album {
		t.Errorf("Expected Album %s, got %s", album, unifiedTrack.Album)
	}
	if unifiedTrack.Source != "youtube" {
		t.Errorf("Expected Source youtube, got %s", unifiedTrack.Source)
	}
	if unifiedTrack.ImageURL != imageURL {
		t.Errorf("Expected ImageURL %s, got %s", imageURL, unifiedTrack.ImageURL)
	}
	if unifiedTrack.Duration != duration {
		t.Errorf("Expected Duration %d, got %d", duration, unifiedTrack.Duration)
	}
	expectedURL := "https://music.youtube.com/watch?v=" + id
	if unifiedTrack.ExternalURL != expectedURL {
		t.Errorf("Expected ExternalURL %s, got %s", expectedURL, unifiedTrack.ExternalURL)
	}
}