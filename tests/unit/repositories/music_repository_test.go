package repositories_test

import (
	"backend/repositories"
	"backend/server/models"
	"backend/tests/mocks"
	"errors"
	"testing"
)

func TestNewMusicRepository(t *testing.T) {
	mockGenius := &mocks.MockGeniusService{}
	repo := repositories.NewMusicRepository(mockGenius)
	
	if repo == nil {
		t.Fatal("NewMusicRepository should not return nil")
	}
}

func TestMusicRepository_UpdateNowPlaying(t *testing.T) {
	mockGenius := &mocks.MockGeniusService{}
	repo := repositories.NewMusicRepository(mockGenius)
	
	track := models.SpotifyTrack{
		ID:     "test123",
		Name:   "Test Song",
		Artist: "Test Artist",
		Album:  "Test Album",
	}
	
	repo.UpdateNowPlaying(track)
	
	// Verify track is stored
	if !repo.IsPlaying() {
		t.Error("Should be playing after UpdateNowPlaying")
	}
	
	nowPlaying := repo.GetNowPlaying()
	if nowPlaying.TrackID != track.ID {
		t.Errorf("Expected TrackID %s, got %s", track.ID, nowPlaying.TrackID)
	}
	if nowPlaying.TrackName != track.Name {
		t.Errorf("Expected TrackName %s, got %s", track.Name, nowPlaying.TrackName)
	}
	if nowPlaying.Source != "spotify" {
		t.Errorf("Expected Source spotify, got %s", nowPlaying.Source)
	}
	
	// Verify it's in play history
	history := repo.GetPlayHistory()
	if len(history) != 1 {
		t.Errorf("Expected 1 item in history, got %d", len(history))
	}
	if history[0].TrackID != track.ID {
		t.Errorf("Expected history TrackID %s, got %s", track.ID, history[0].TrackID)
	}
	if history[0].Source != "spotify" {
		t.Errorf("Expected history Source spotify, got %s", history[0].Source)
	}
}

func TestMusicRepository_UpdateNowPlayingUnified(t *testing.T) {
	mockGenius := &mocks.MockGeniusService{}
	repo := repositories.NewMusicRepository(mockGenius)
	
	track := models.UnifiedTrack{
		ID:     "yt123",
		Name:   "Test YouTube Song",
		Artist: "Test YouTube Artist",
		Album:  "Test YouTube Album",
		Source: "youtube",
	}
	
	repo.UpdateNowPlayingUnified(track)
	
	// Verify track is stored
	if !repo.IsPlaying() {
		t.Error("Should be playing after UpdateNowPlayingUnified")
	}
	
	nowPlaying := repo.GetNowPlaying()
	if nowPlaying.TrackID != track.ID {
		t.Errorf("Expected TrackID %s, got %s", track.ID, nowPlaying.TrackID)
	}
	if nowPlaying.TrackName != track.Name {
		t.Errorf("Expected TrackName %s, got %s", track.Name, nowPlaying.TrackName)
	}
	if nowPlaying.Source != "youtube" {
		t.Errorf("Expected Source youtube, got %s", nowPlaying.Source)
	}
	
	// Verify it's in play history
	history := repo.GetPlayHistory()
	if len(history) != 1 {
		t.Errorf("Expected 1 item in history, got %d", len(history))
	}
	if history[0].TrackID != track.ID {
		t.Errorf("Expected history TrackID %s, got %s", track.ID, history[0].TrackID)
	}
	if history[0].Source != "youtube" {
		t.Errorf("Expected history Source youtube, got %s", history[0].Source)
	}
}

func TestMusicRepository_IsPlaying(t *testing.T) {
	mockGenius := &mocks.MockGeniusService{}
	repo := repositories.NewMusicRepository(mockGenius)
	
	// Should not be playing initially
	if repo.IsPlaying() {
		t.Error("Should not be playing initially")
	}
	
	// Should be playing after update
	track := models.SpotifyTrack{
		ID:     "test123",
		Name:   "Test Song",
		Artist: "Test Artist",
		Album:  "Test Album",
	}
	repo.UpdateNowPlaying(track)
	
	if !repo.IsPlaying() {
		t.Error("Should be playing after UpdateNowPlaying")
	}
}

func TestMusicRepository_GetCurrentSongInfo(t *testing.T) {
	mockGenius := &mocks.MockGeniusService{}
	repo := repositories.NewMusicRepository(mockGenius)
	
	// Should return empty string initially
	if repo.GetCurrentSongInfo() != "" {
		t.Error("Should return empty string when not playing")
	}
	
	// Should return formatted string after update
	track := models.SpotifyTrack{
		ID:     "test123",
		Name:   "Test Song",
		Artist: "Test Artist",
		Album:  "Test Album",
	}
	repo.UpdateNowPlaying(track)
	
	expected := "Test Song by Test Artist"
	if repo.GetCurrentSongInfo() != expected {
		t.Errorf("Expected %s, got %s", expected, repo.GetCurrentSongInfo())
	}
}

func TestMusicRepository_GetLyricsForCurrentSong(t *testing.T) {
	mockLyrics := "Test lyrics content"
	mockGenius := &mocks.MockGeniusService{
		GetLyricsFunc: func(trackName, artistName string) (string, error) {
			if trackName == "Test Song" && artistName == "Test Artist" {
				return mockLyrics, nil
			}
			return "", errors.New("song not found")
		},
	}
	repo := repositories.NewMusicRepository(mockGenius)
	
	// Should return error when no song is playing
	_, err := repo.GetLyricsForCurrentSong()
	if err == nil {
		t.Error("Should return error when no song is playing")
	}
	
	// Update with a track
	track := models.SpotifyTrack{
		ID:     "test123",
		Name:   "Test Song",
		Artist: "Test Artist",
		Album:  "Test Album",
	}
	repo.UpdateNowPlaying(track)
	
	// Should fetch lyrics
	lyrics, err := repo.GetLyricsForCurrentSong()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if lyrics != mockLyrics {
		t.Errorf("Expected lyrics %s, got %s", mockLyrics, lyrics)
	}
	
	// Should use cached lyrics on second call (verify by changing mock behavior)
	mockGenius.GetLyricsFunc = func(trackName, artistName string) (string, error) {
		return "Different lyrics", nil
	}
	
	lyrics2, err := repo.GetLyricsForCurrentSong()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if lyrics2 != mockLyrics {
		t.Errorf("Expected cached lyrics %s, got %s", mockLyrics, lyrics2)
	}
}

func TestMusicRepository_GetLyricsError(t *testing.T) {
	mockError := errors.New("genius API error")
	mockGenius := &mocks.MockGeniusService{
		GetLyricsFunc: func(trackName, artistName string) (string, error) {
			return "", mockError
		},
	}
	repo := repositories.NewMusicRepository(mockGenius)
	
	track := models.SpotifyTrack{
		ID:     "test123",
		Name:   "Test Song",
		Artist: "Test Artist",
		Album:  "Test Album",
	}
	repo.UpdateNowPlaying(track)
	
	_, err := repo.GetLyricsForCurrentSong()
	if err == nil {
		t.Error("Should return error when Genius service fails")
	}
}

func TestMusicRepository_PlayHistory(t *testing.T) {
	mockGenius := &mocks.MockGeniusService{}
	repo := repositories.NewMusicRepository(mockGenius)
	
	// Add multiple tracks (mix of Spotify and YouTube)
	tracks := []models.UnifiedTrack{
		{ID: "track1", Name: "Song 1", Artist: "Artist 1", Album: "Album 1", Source: "spotify"},
		{ID: "yt1", Name: "YouTube Song", Artist: "YouTube Artist", Album: "YouTube Album", Source: "youtube"},
		{ID: "track3", Name: "Song 3", Artist: "Artist 3", Album: "Album 3", Source: "spotify"},
	}
	
	for _, track := range tracks {
		repo.UpdateNowPlayingUnified(track)
	}
	
	history := repo.GetPlayHistory()
	if len(history) != 3 {
		t.Errorf("Expected 3 items in history, got %d", len(history))
	}
	
	// Most recent should be first
	if history[0].TrackID != "track3" {
		t.Errorf("Expected first item to be track3, got %s", history[0].TrackID)
	}
	if history[0].Source != "spotify" {
		t.Errorf("Expected first item source to be spotify, got %s", history[0].Source)
	}
	if history[1].TrackID != "yt1" {
		t.Errorf("Expected second item to be yt1, got %s", history[1].TrackID)
	}
	if history[1].Source != "youtube" {
		t.Errorf("Expected second item source to be youtube, got %s", history[1].Source)
	}
	if history[2].TrackID != "track1" {
		t.Errorf("Expected last item to be track1, got %s", history[2].TrackID)
	}
	if history[2].Source != "spotify" {
		t.Errorf("Expected last item source to be spotify, got %s", history[2].Source)
	}
}