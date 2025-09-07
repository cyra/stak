package ports

import "stak/internal/models"

// StoragePort defines the interface for storage operations
type StoragePort interface {
	Initialize() error
	SaveEntry(entry *models.Entry) error
	SaveEntryForTomorrow(entry *models.Entry) error
	LoadTodayEntries() ([]models.Entry, error)
	LoadFilteredEntries(entryType models.EntryType) ([]models.Entry, error)
	SearchEntries(query string, linksOnly bool) ([]models.Entry, error)
}