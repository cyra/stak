package application

import (
	"strings"
	"time"

	"stak/internal/models"
	"stak/internal/ports"
)

type EntryService struct {
	storage     ports.StoragePort
	categorizer ports.CategorizerPort
	extractor   ports.ExtractorPort
	searcher    ports.SearchPort
}

func NewEntryService(
	storage ports.StoragePort,
	categorizer ports.CategorizerPort,
	extractor ports.ExtractorPort,
	searcher ports.SearchPort,
) *EntryService {
	return &EntryService{
		storage:     storage,
		categorizer: categorizer,
		extractor:   extractor,
		searcher:    searcher,
	}
}

func (s *EntryService) CreateEntry(content string, forceType *models.EntryType) (*models.Entry, error) {
	entry := models.NewEntry(content)
	
	if forceType != nil {
		entry.Type = *forceType
		if *forceType == models.TypeTodo {
			entry.TodoStatus = models.TodoPending
			entry.Tags = []string{"todo", "task"}
		}
	} else {
		s.categorizer.CategoriseEntry(entry)
	}
	
	// Handle link extraction asynchronously if needed
	if entry.Type == models.TypeLink && entry.URL != "" {
		go func() {
			if title, err := s.extractor.GetURLTitle(entry.URL); err == nil {
				entry.URLTitle = title
				s.storage.SaveEntry(entry) // Save updated entry with title
			}
		}()
	}
	
	return entry, s.storage.SaveEntry(entry)
}

func (s *EntryService) CreateTomorrowEntry(content string) error {
	entry := models.NewEntry(content)
	s.categorizer.CategoriseEntry(entry)
	return s.storage.SaveEntryForTomorrow(entry)
}

func (s *EntryService) ToggleTodoStatus(entryID string, entries []models.Entry) (*models.Entry, error) {
	for i := range entries {
		if entries[i].ID == entryID && entries[i].Type == models.TypeTodo {
			if entries[i].TodoStatus == models.TodoPending {
				entries[i].TodoStatus = models.TodoCompleted
			} else {
				entries[i].TodoStatus = models.TodoPending
			}
			entries[i].UpdatedAt = time.Now()
			
			err := s.storage.SaveEntry(&entries[i])
			return &entries[i], err
		}
	}
	return nil, nil
}

func (s *EntryService) LoadTodayEntries() ([]models.Entry, error) {
	return s.storage.LoadTodayEntries()
}

func (s *EntryService) LoadFilteredEntries(entryType models.EntryType) ([]models.Entry, error) {
	return s.storage.LoadFilteredEntries(entryType)
}

func (s *EntryService) SearchEntries(query string, linksOnly bool) ([]models.Entry, error) {
	normalizedQuery := strings.ToLower(strings.TrimSpace(query))
	if normalizedQuery == "" {
		return []models.Entry{}, nil
	}
	
	return s.storage.SearchEntries(normalizedQuery, linksOnly)
}