package ui

import (
	"stak/internal/models"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type entriesLoadedMsg struct {
	entries []models.Entry
}

type filteredEntriesLoadedMsg struct {
	entries []models.Entry
	mode    mode
}

type calendarEntriesLoadedMsg struct {
	calendarEntries map[string][]models.Entry
	selectedDate    time.Time
}

type entryAddedMsg struct{}

func (m Model) loadTodayEntries() tea.Cmd {
	return func() tea.Msg {
		entries, err := m.entryService.LoadTodayEntries()
		if err != nil {
			return entriesLoadedMsg{entries: []models.Entry{}}
		}
		return entriesLoadedMsg{entries: entries}
	}
}

func (m Model) searchEntries(query string, linksOnly bool) tea.Cmd {
	return func() tea.Msg {
		entries, err := m.entryService.SearchEntries(query, linksOnly)
		if err != nil {
			return entriesLoadedMsg{entries: []models.Entry{}}
		}

		return entriesLoadedMsg{entries: entries}
	}
}

func (m Model) loadFilteredEntries() tea.Cmd {
	currentMode := m.currentMode // Capture current mode
	return func() tea.Msg {
		var entries []models.Entry
		var err error

		switch currentMode {
		case todoMode:
			entries, err = m.entryService.LoadFilteredEntries(models.TypeTodo)
		case stakMode:
			entries, err = m.entryService.LoadTodayEntries()
		default:
			entries, err = m.entryService.LoadTodayEntries()
		}

		if err != nil {
			return filteredEntriesLoadedMsg{entries: []models.Entry{}, mode: currentMode}
		}

		return filteredEntriesLoadedMsg{entries: entries, mode: currentMode}
	}
}

func (m Model) loadCalendarEntries() tea.Cmd {
	selectedDate := m.selectedDate
	return func() tea.Msg {
		// Load entries for the current month
		startOfMonth := time.Date(selectedDate.Year(), selectedDate.Month(), 1, 0, 0, 0, 0, selectedDate.Location())
		_ = startOfMonth // TODO: Use for date range queries when implemented

		// Load all entries to populate the calendar
		entries, err := m.entryService.LoadAllEntries()
		if err != nil {
			return calendarEntriesLoadedMsg{
				calendarEntries: make(map[string][]models.Entry),
				selectedDate:    selectedDate,
			}
		}

		// Group entries by date
		calendarEntries := make(map[string][]models.Entry)
		for _, entry := range entries {
			dateKey := entry.CreatedAt.Format("2006-01-02")
			calendarEntries[dateKey] = append(calendarEntries[dateKey], entry)
		}

		return calendarEntriesLoadedMsg{
			calendarEntries: calendarEntries,
			selectedDate:    selectedDate,
		}
	}
}

// Load entries for a specific date
func (m Model) loadEntriesForDate(date time.Time) tea.Cmd {
	return func() tea.Msg {
		// Load all entries and filter by date
		allEntries, err := m.entryService.LoadAllEntries()
		if err != nil {
			return entriesLoadedMsg{entries: []models.Entry{}}
		}

		// Filter entries for the specific date
		dateKey := date.Format("2006-01-02")
		var dayEntries []models.Entry
		for _, entry := range allEntries {
			entryDateKey := entry.CreatedAt.Format("2006-01-02")
			if entryDateKey == dateKey {
				dayEntries = append(dayEntries, entry)
			}
		}

		return entriesLoadedMsg{entries: dayEntries}
	}
}
