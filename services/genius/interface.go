package genius

// Service defines the interface for Genius/lyrics operations
type Service interface {
	GetLyrics(trackName, artistName string) (string, error)
}