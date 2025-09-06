package categorizer

import (
	"testing"

	"stak/internal/models"
)

func TestCategoriseEntry(t *testing.T) {
	categoriser := New()

	tests := []struct {
		name           string
		content        string
		expectedType   models.EntryType
		expectedTags   []string
		expectedStatus models.TodoStatus
	}{
		{
			name:         "URL should be categorised as link",
			content:      "Check out https://go.dev for Go documentation",
			expectedType: models.TypeLink,
			expectedTags: []string{"link", "web", "reference"},
		},
		{
			name:           "Todo with checkbox should be categorised as todo",
			content:        "- [ ] Fix authentication bug",
			expectedType:   models.TypeTodo,
			expectedTags:   []string{"todo", "task"},
			expectedStatus: models.TodoPending,
		},
		{
			name:         "Code snippet should be categorised as code",
			content:      "```go\nfunc main() {\n  fmt.Println(\"Hello\")\n}\n```",
			expectedType: models.TypeCode,
			expectedTags: []string{"code", "golang"},
		},
		{
			name:         "Question should be categorised as question",
			content:      "How do I implement middleware in Go?",
			expectedType: models.TypeQuestion,
			expectedTags: []string{"question", "inquiry", "golang"},
		},
		{
			name:         "Meeting content should be categorised as meeting",
			content:      "Standup meeting at 9am tomorrow",
			expectedType: models.TypeMeeting,
			expectedTags: []string{"meeting", "discussion"},
		},
		{
			name:         "Plain note should be categorised as note",
			content:      "This is just a regular note",
			expectedType: models.TypeNote,
			expectedTags: []string{"note"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := models.NewEntry(tt.content)
			categoriser.CategoriseEntry(entry)

			if entry.Type != tt.expectedType {
				t.Errorf("expected type %v, got %v", tt.expectedType, entry.Type)
			}

			if tt.expectedStatus != "" && entry.TodoStatus != tt.expectedStatus {
				t.Errorf("expected todo status %v, got %v", tt.expectedStatus, entry.TodoStatus)
			}

			for _, expectedTag := range tt.expectedTags {
				found := false
				for _, tag := range entry.Tags {
					if tag == expectedTag {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected tag %v not found in tags %v", expectedTag, entry.Tags)
				}
			}
		})
	}
}