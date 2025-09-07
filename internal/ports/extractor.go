package ports

// ExtractorPort defines the interface for link extraction and metadata
type ExtractorPort interface {
	GetURLTitle(url string) (string, error)
}