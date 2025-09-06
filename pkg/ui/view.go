package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"stak/internal/models"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	inputStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7D56F4")).
			Padding(1)

	entryStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(lipgloss.Color("#383838")).
			Padding(0, 1).
			Margin(0, 0, 1, 0)

	selectedEntryStyle = entryStyle.Copy().
				BorderForeground(lipgloss.Color("#7D56F4")).
				Background(lipgloss.Color("#2A2A2A"))

	todoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFA500"))

	completedTodoStyle = lipgloss.NewStyle().
				Strikethrough(true).
				Foreground(lipgloss.Color("#666666"))

	linkStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00CED1")).
			Underline(true)

	tagStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#98FB98")).
			Italic(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			Border(lipgloss.RoundedBorder()).
			Padding(1).
			Margin(1, 0)
)

func (m Model) View() string {
	if !m.ready {
		return "Loading..."
	}

	var sections []string

	sections = append(sections, m.renderHeader())

	if m.showHelp {
		sections = append(sections, m.renderHelp())
	} else {
		sections = append(sections, m.renderEntries())
	}

	sections = append(sections, m.renderInput())
	sections = append(sections, m.renderFooter())

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m Model) renderHeader() string {
	title := "stak"
	switch m.currentMode {
	case todayMode:
		title += " - Today"
	case searchMode:
		title += fmt.Sprintf(" - Search: %s", m.searchQuery)
	}

	return titleStyle.Render(title)
}

func (m Model) renderHelp() string {
	help := strings.Join(m.commands, "\n")
	return helpStyle.Render(help)
}

func (m Model) renderEntries() string {
	if len(m.entries) == 0 {
		empty := "No entries found."
		if m.currentMode == normalMode || m.currentMode == todayMode {
			empty = "No entries for today. Start typing to add your first entry!"
		}
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666")).
			Italic(true).
			Margin(2, 0).
			Render(empty)
	}

	var renderedEntries []string
	maxVisible := m.height - 10

	start := 0
	end := len(m.entries)

	if len(m.entries) > maxVisible {
		if m.selectedIdx >= maxVisible/2 {
			start = m.selectedIdx - maxVisible/2
			if start+maxVisible > len(m.entries) {
				start = len(m.entries) - maxVisible
			}
		}
		end = start + maxVisible
		if end > len(m.entries) {
			end = len(m.entries)
		}
	}

	for i := start; i < end; i++ {
		entry := m.entries[i]
		style := entryStyle
		if i == m.selectedIdx {
			style = selectedEntryStyle
		}

		content := m.renderEntry(entry)
		renderedEntries = append(renderedEntries, style.Render(content))
	}

	return strings.Join(renderedEntries, "\n")
}

func (m Model) renderEntry(entry models.Entry) string {
	var parts []string

	timestamp := entry.CreatedAt.Format("15:04")
	timeStr := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render(timestamp)

	typeIcon := m.getTypeIcon(entry.Type)
	
	var contentStr string
	switch entry.Type {
	case models.TypeTodo:
		if entry.TodoStatus == models.TodoCompleted {
			contentStr = completedTodoStyle.Render("✓ " + entry.Content)
		} else {
			contentStr = todoStyle.Render("□ " + entry.Content)
		}
	case models.TypeLink:
		contentStr = linkStyle.Render(entry.Content)
		if entry.URL != "" {
			contentStr += "\n  " + lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Render(entry.URL)
		}
	default:
		contentStr = entry.Content
	}

	parts = append(parts, fmt.Sprintf("%s %s %s", timeStr, typeIcon, contentStr))

	if len(entry.Tags) > 0 {
		tagsStr := tagStyle.Render("#" + strings.Join(entry.Tags, " #"))
		parts = append(parts, "  "+tagsStr)
	}

	return strings.Join(parts, "\n")
}

func (m Model) getTypeIcon(entryType models.EntryType) string {
	switch entryType {
	case models.TypeTodo:
		return "☐"
	case models.TypeLink:
		return "🔗"
	case models.TypeCode:
		return "💻"
	case models.TypeQuestion:
		return "❓"
	case models.TypeMeeting:
		return "👥"
	case models.TypeIdea:
		return "💡"
	default:
		return "📝"
	}
}

func (m Model) renderInput() string {
	placeholder := "Enter your thoughts, links, todos..."
	if m.currentMode == todayMode {
		placeholder = "Add to today's entries (Tab to toggle todos)"
	} else if m.currentMode == searchMode {
		placeholder = "Press Esc to go back"
	}

	m.textInput.Placeholder = placeholder
	return inputStyle.Render(m.textInput.View())
}

func (m Model) renderFooter() string {
	var parts []string
	
	if m.currentMode == todayMode && len(m.entries) > 0 {
		todos := 0
		completed := 0
		for _, entry := range m.entries {
			if entry.Type == models.TypeTodo {
				todos++
				if entry.TodoStatus == models.TodoCompleted {
					completed++
				}
			}
		}
		if todos > 0 {
			progress := fmt.Sprintf("Todos: %d/%d completed", completed, todos)
			parts = append(parts, progress)
		}
	}
	
	help := "ESC: back • Ctrl+C: quit • /help: commands"
	if m.currentMode == todayMode {
		help = "↑↓: navigate • Tab: toggle todo • " + help
	}
	
	parts = append(parts, help)
	
	footer := strings.Join(parts, " • ")
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666666")).
		Render(footer)
}