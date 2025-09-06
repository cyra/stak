package models

import (
	"time"
)

type EntryType string

const (
	TypeNote     EntryType = "note"
	TypeTodo     EntryType = "todo"
	TypeLink     EntryType = "link"
	TypeCode     EntryType = "code"
	TypeQuestion EntryType = "question"
	TypeMeeting  EntryType = "meeting"
	TypeIdea     EntryType = "idea"
)

type TodoStatus string

const (
	TodoPending   TodoStatus = "pending"
	TodoCompleted TodoStatus = "completed"
	TodoCancelled TodoStatus = "cancelled"
)

type Entry struct {
	ID          string            `yaml:"id" json:"id"`
	Content     string            `yaml:"content" json:"content"`
	Type        EntryType         `yaml:"type" json:"type"`
	Tags        []string          `yaml:"tags,omitempty" json:"tags,omitempty"`
	URL         string            `yaml:"url,omitempty" json:"url,omitempty"`
	URLTitle    string            `yaml:"url_title,omitempty" json:"url_title,omitempty"`
	TodoStatus  TodoStatus        `yaml:"todo_status,omitempty" json:"todo_status,omitempty"`
	CreatedAt   time.Time         `yaml:"created_at" json:"created_at"`
	UpdatedAt   time.Time         `yaml:"updated_at" json:"updated_at"`
	Metadata    map[string]string `yaml:"metadata,omitempty" json:"metadata,omitempty"`
}

type DayFile struct {
	Date    time.Time `yaml:"date" json:"date"`
	Entries []Entry   `yaml:"entries" json:"entries"`
}

func NewEntry(content string) *Entry {
	now := time.Now()
	return &Entry{
		ID:        generateID(),
		Content:   content,
		CreatedAt: now,
		UpdatedAt: now,
		Tags:      []string{},
		Metadata:  make(map[string]string),
	}
}

func generateID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(6)
}

func randomString(n int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}