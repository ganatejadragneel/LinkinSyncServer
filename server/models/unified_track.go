package models

// UnifiedTrack represents a track from any supported music service
type UnifiedTrack struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Artist     string `json:"artist"`
	Album      string `json:"album,omitempty"`
	Source     string `json:"source"`      // "spotify" or "youtube"
	PreviewURL string `json:"preview_url,omitempty"`
	ExternalURL string `json:"external_url,omitempty"`
	Duration   int    `json:"duration,omitempty"` // duration in seconds
	ImageURL   string `json:"image_url,omitempty"`
}

// ToSpotifyTrack converts UnifiedTrack to SpotifyTrack for backward compatibility
func (t UnifiedTrack) ToSpotifyTrack() SpotifyTrack {
	return SpotifyTrack{
		ID:         t.ID,
		Name:       t.Name,
		Artist:     t.Artist,
		Album:      t.Album,
		PreviewURL: t.PreviewURL,
	}
}

// FromSpotifyTrack creates UnifiedTrack from SpotifyTrack
func FromSpotifyTrack(track SpotifyTrack) UnifiedTrack {
	return UnifiedTrack{
		ID:         track.ID,
		Name:       track.Name,
		Artist:     track.Artist,
		Album:      track.Album,
		Source:     "spotify",
		PreviewURL: track.PreviewURL,
	}
}

// FromYouTubeTrack creates UnifiedTrack from YouTube track data
func FromYouTubeTrack(id, name, artist, album, imageURL string, duration int) UnifiedTrack {
	return UnifiedTrack{
		ID:       id,
		Name:     name,
		Artist:   artist,
		Album:    album,
		Source:   "youtube",
		Duration: duration,
		ImageURL: imageURL,
		ExternalURL: "https://music.youtube.com/watch?v=" + id,
	}
}