package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
	"stak/internal/config"
	"stak/internal/models"
)

type Storage struct {
	config *config.Config
}

func New(cfg *config.Config) *Storage {
	return &Storage{
		config: cfg,
	}
}

func (s *Storage) Initialize() error {
	return s.config.EnsureDataDir()
}

func (s *Storage) SaveEntry(entry *models.Entry) error {
	date := entry.CreatedAt.Format(s.config.DateFormat)
	dayFile, err := s.loadDayFile(date)
	if err != nil {
		dayFile = &models.DayFile{
			Date:    entry.CreatedAt,
			Entries: []models.Entry{},
		}
	}

	found := false
	for i, existing := range dayFile.Entries {
		if existing.ID == entry.ID {
			dayFile.Entries[i] = *entry
			found = true
			break
		}
	}

	if !found {
		dayFile.Entries = append(dayFile.Entries, *entry)
	}

	return s.saveDayFile(date, dayFile)
}

func (s *Storage) LoadTodayEntries() ([]models.Entry, error) {
	today := time.Now().Format(s.config.DateFormat)
	dayFile, err := s.loadDayFile(today)
	if err != nil {
		return []models.Entry{}, nil
	}
	return dayFile.Entries, nil
}

func (s *Storage) LoadAllEntries() ([]models.Entry, error) {
	var allEntries []models.Entry
	
	files, err := filepath.Glob(filepath.Join(s.config.DataDir, "*.md"))
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		dayFile, err := s.loadDayFileFromPath(file)
		if err != nil {
			continue
		}
		allEntries = append(allEntries, dayFile.Entries...)
	}

	return allEntries, nil
}

func (s *Storage) SearchEntries(query string) ([]models.Entry, error) {
	allEntries, err := s.LoadAllEntries()
	if err != nil {
		return nil, err
	}

	var results []models.Entry
	queryLower := strings.ToLower(query)

	for _, entry := range allEntries {
		if s.matchesQuery(entry, queryLower) {
			results = append(results, entry)
		}
	}

	return results, nil
}

func (s *Storage) SearchLinks(query string) ([]models.Entry, error) {
	allEntries, err := s.LoadAllEntries()
	if err != nil {
		return nil, err
	}

	var results []models.Entry
	queryLower := strings.ToLower(query)

	for _, entry := range allEntries {
		if entry.Type == models.TypeLink && s.matchesQuery(entry, queryLower) {
			results = append(results, entry)
		}
	}

	return results, nil
}

func (s *Storage) loadDayFile(date string) (*models.DayFile, error) {
	filePath := filepath.Join(s.config.DataDir, date+".md")
	return s.loadDayFileFromPath(filePath)
}

func (s *Storage) loadDayFileFromPath(filePath string) (*models.DayFile, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	parts := strings.SplitN(string(content), "---", 3)
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid markdown file format")
	}

	var dayFile models.DayFile
	if err := yaml.Unmarshal([]byte(parts[1]), &dayFile); err != nil {
		return nil, err
	}

	return &dayFile, nil
}

func (s *Storage) saveDayFile(date string, dayFile *models.DayFile) error {
	filePath := filepath.Join(s.config.DataDir, date+".md")
	
	yamlData, err := yaml.Marshal(dayFile)
	if err != nil {
		return err
	}

	content := fmt.Sprintf("---\n%s---\n\n# %s\n\n", string(yamlData), dayFile.Date.Format("January 2, 2006"))
	
	for _, entry := range dayFile.Entries {
		content += s.formatEntryAsMarkdown(entry)
	}

	return os.WriteFile(filePath, []byte(content), 0644)
}

func (s *Storage) formatEntryAsMarkdown(entry models.Entry) string {
	var md strings.Builder
	
	md.WriteString(fmt.Sprintf("## %s\n\n", entry.CreatedAt.Format("15:04")))
	
	if entry.Type == models.TypeTodo {
		checkbox := "[ ]"
		if entry.TodoStatus == models.TodoCompleted {
			checkbox = "[x]"
		}
		md.WriteString(fmt.Sprintf("- %s %s\n", checkbox, entry.Content))
	} else {
		md.WriteString(fmt.Sprintf("%s\n", entry.Content))
	}
	
	if entry.URL != "" {
		if entry.URLTitle != "" {
			md.WriteString(fmt.Sprintf("\n[%s](%s)\n", entry.URLTitle, entry.URL))
		} else {
			md.WriteString(fmt.Sprintf("\n%s\n", entry.URL))
		}
	}
	
	if len(entry.Tags) > 0 {
		md.WriteString(fmt.Sprintf("\n*Tags: %s*\n", strings.Join(entry.Tags, ", ")))
	}
	
	md.WriteString("\n---\n\n")
	
	return md.String()
}

func (s *Storage) matchesQuery(entry models.Entry, query string) bool {
	contentMatch := strings.Contains(strings.ToLower(entry.Content), query)
	
	tagMatch := false
	for _, tag := range entry.Tags {
		if strings.Contains(strings.ToLower(tag), query) {
			tagMatch = true
			break
		}
	}
	
	urlMatch := strings.Contains(strings.ToLower(entry.URL), query) || 
	           strings.Contains(strings.ToLower(entry.URLTitle), query)
	
	return contentMatch || tagMatch || urlMatch
}