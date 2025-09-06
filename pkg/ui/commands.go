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
		allEntries, err := m.storage.LoadAllEntries()
		if err != nil {
			return entriesLoadedMsg{entries: []models.Entry{}}
		}

		m.searcher.SetEntries(allEntries)
		
		var entries []models.Entry
		if linksOnly {
			entries = m.searcher.SearchLinks(query)
		} else {
			entries = m.searcher.RankedSearch(query)
		}

		return entriesLoadedMsg{entries: entries}
	}
}