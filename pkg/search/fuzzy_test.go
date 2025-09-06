package search

import (
	"testing"
	"time"

	"stak/internal/models"
)

func TestFuzzySearch(t *testing.T) {
	searcher := NewFuzzySearcher()

	entries := []models.Entry{
		{
			ID:        "1",
			Content:   "Fix authentication bug in Go service",
			Type:      models.TypeTodo,
			Tags:      []string{"golang", "bug", "todo"},
			CreatedAt: time.Now(),
		},
		{
			ID:        "2",
			Content:   "https://go.dev/blog/error-handling",
			URL:       "https://go.dev/blog/error-handling",
			URLTitle:  "Error Handling in Go",
			Type:      models.TypeLink,
			Tags:      []string{"golang", "link", "learning"},
			CreatedAt: time.Now().Add(-1 * time.Hour),
		},
		{
			ID:        "3",
			Content:   "How to implement middleware in JavaScript?",
			Type:      models.TypeQuestion,
			Tags:      []string{"javascript", "question"},
			CreatedAt: time.Now().Add(-2 * time.Hour),
		},
	}

	searcher.SetEntries(entries)

	tests := []struct {
		name            string
		query           string
		expectedResults int
		shouldContainID string
	}{
		{
			name:            "Search for 'go' should find Go-related entries",
			query:           "go",
			expectedResults: 2,
			shouldContainID: "1",
		},
		{
			name:            "Search for 'javascript' should find JS entry",
			query:           "javascript",
			expectedResults: 1,
			shouldContainID: "3",
		},
		{
			name:            "Empty query should return all entries",
			query:           "",
			expectedResults: 3,
		},
		{
			name:            "Search for non-existent term",
			query:           "nonexistent",
			expectedResults: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := searcher.RankedSearch(tt.query)

			if len(results) != tt.expectedResults {
				t.Errorf("expected %d results, got %d", tt.expectedResults, len(results))
			}

			if tt.shouldContainID != "" {
				found := false
				for _, result := range results {
					if result.ID == tt.shouldContainID {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected to find entry with ID %s", tt.shouldContainID)
				}
			}
		})
	}
}

func TestSearchLinks(t *testing.T) {
	searcher := NewFuzzySearcher()

	entries := []models.Entry{
		{
			ID:      "1",
			Content: "Regular note",
			Type:    models.TypeNote,
		},
		{
			ID:      "2",
			Content: "https://github.com/golang/go",
			Type:    models.TypeLink,
			Tags:    []string{"golang", "github"},
		},
		{
			ID:      "3",
			Content: "https://docs.docker.com",
			Type:    models.TypeLink,
			Tags:    []string{"docker", "documentation"},
		},
	}

	searcher.SetEntries(entries)

	results := searcher.SearchLinks("golang")
	
	if len(results) != 1 {
		t.Errorf("expected 1 link result, got %d", len(results))
	}

	if len(results) > 0 && results[0].ID != "2" {
		t.Errorf("expected link with ID '2', got ID '%s'", results[0].ID)
	}
}