package ports

import "stak/internal/models"

// SearchPort defines the interface for search operations
type SearchPort interface {
	Search(query string, entries []models.Entry) []models.Entry
}