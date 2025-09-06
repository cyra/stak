package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"stak/internal/models"
)

type entriesLoadedMsg struct {
	entries []models.Entry
}

type entryAddedMsg struct{}

func (m Model) loadTodayEntries() tea.Cmd {
	return func() tea.Msg {
		entries, err := m.storage.LoadTodayEntries()
		if err != nil {
			return entriesLoadedMsg{entries: []models.Entry{}}
		}
		return entriesLoadedMsg{entries: entries}
	}
}

func (m Model) searchEntries(query string, linksOnly bool) tea.Cmd {
	return func() tea.Msg {
		var entries []models.Entry
		var err error

		if linksOnly {
			entries, err = m.storage.SearchLinks(query)
		} else {
			entries, err = m.storage.SearchEntries(query)
		}

		if err != nil {
			return entriesLoadedMsg{entries: []models.Entry{}}
		}
		return entriesLoadedMsg{entries: entries}
	}
}