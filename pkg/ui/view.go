package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"stak/internal/models"
)

// Clean, minimal styles
var (
	headerClean = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Padding(0, 1)

	todoHeaderClean = headerClean.Copy().
		Background(lipgloss.Color("#FFA500"))

	stakHeaderClean = headerClean.Copy().
		Background(lipgloss.Color("#7D56F4"))

	contentClean = lipgloss.NewStyle().
		Padding(1, 2)

	inputClean = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), true, false, false, false).
		BorderForeground(lipgloss.Color("#666666")).
		Padding(0, 2)

	statusClean = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666666")).
		Align(lipgloss.Center)

	selectedEntryClean = lipgloss.NewStyle().
		Background(lipgloss.Color("#444444")).
		Foreground(lipgloss.Color("#FFFFFF"))
)

// Main view function - clean and stable
func (m Model) View() string {
	if !m.ready {
		return "Loading..."
	}

	// Fixed dimensions - prevents all jumping
	totalHeight := m.height
	headerHeight := 1
	statusHeight := 1
	inputHeight := 3
	
	contentHeight := totalHeight - headerHeight - statusHeight - inputHeight
	if contentHeight < 3 {
		contentHeight = 3
	}

	var sections []string

	// 1. Header (1 line, fixed)
	sections = append(sections, m.renderHeaderClean())

	// 2. Content (fixed height)
	if m.showHelp {
		sections = append(sections, m.renderHelpClean(contentHeight))
	} else if m.currentMode == todoListMode && m.todoList != nil {
		sections = append(sections, m.renderTodoListClean(contentHeight))
	} else {
		sections = append(sections, m.renderEntriesClean(contentHeight))
	}

	// 3. Input (3 lines fixed) or empty space
	if m.currentMode != todoListMode {
		sections = append(sections, m.renderInputClean())
	} else {
		sections = append(sections, strings.Repeat("\n", inputHeight))
	}

	// 4. Status bar (1 line, fixed)
	sections = append(sections, m.renderStatusClean())

	return strings.Join(sections, "")
}

func (m Model) renderHeaderClean() string {
	var title string
	var style lipgloss.Style

	switch m.currentMode {
	case todoMode:
		title = "TODO"
		style = todoHeaderClean
	case todoListMode:
		title = "TODOS"
		style = todoHeaderClean
	case scratchpadMode:
		title = "STAK"
		style = stakHeaderClean
	case todayMode:
		title = "TODAY"
		style = stakHeaderClean
	case searchMode:
		title = fmt.Sprintf("SEARCH: %s", m.searchQuery)
		style = stakHeaderClean
	default:
		title = "STAK"
		style = stakHeaderClean
	}

	return style.Width(m.width).Align(lipgloss.Center).Render(title)
}

func (m Model) renderEntriesClean(height int) string {
	if len(m.entries) == 0 {
		var emptyText string
		switch m.currentMode {
		case todoMode:
			emptyText = "No todos yet. Start typing to add one."
		case scratchpadMode:
			emptyText = "No entries yet. Start typing to add one."
		case todayMode:
			emptyText = "No entries for today. Start typing to add one."
		default:
			emptyText = "No entries found."
		}

		return contentClean.
			Width(m.width).
			Height(height).
			Align(lipgloss.Center, lipgloss.Center).
			Render(emptyText)
	}

	// Show entries IRC/chat style - oldest at top, newest at bottom
	var renderedEntries []string
	
	maxVisible := height - 2 // Account for padding
	if maxVisible < 1 {
		maxVisible = 1
	}

	start := len(m.entries) - maxVisible
	if start < 0 {
		start = 0
	}

	// Render from oldest to newest (chat style)
	for i := start; i < len(m.entries); i++ {
		entry := m.entries[i]
		selected := (i == m.selectedIdx)
		content := m.renderEntryClean(entry, selected)
		renderedEntries = append(renderedEntries, content)
	}

	entriesText := strings.Join(renderedEntries, "\n")
	
	return contentClean.
		Width(m.width).
		Height(height).
		Render(entriesText)
}

func (m Model) renderEntryClean(entry models.Entry, selected bool) string {
	timestamp := entry.CreatedAt.Format("15:04")
	
	var content string
	switch entry.Type {
	case models.TypeTodo:
		if entry.TodoStatus == models.TodoCompleted {
			content = "✓ " + entry.Content
		} else {
			content = "□ " + entry.Content
		}
	default:
		content = entry.Content
	}

	line := fmt.Sprintf("%s %s", timestamp, content)
	
	if selected {
		return selectedEntryClean.Render(line)
	}
	
	return line
}

func (m Model) renderHelpClean(height int) string {
	help := strings.Join(m.commands, "\n")
	return contentClean.
		Width(m.width).
		Height(height).
		Render(help)
}

func (m Model) renderTodoListClean(height int) string {
	if m.todoList == nil {
		return contentClean.
			Width(m.width).
			Height(height).
			Align(lipgloss.Center, lipgloss.Center).
			Render("No todos available")
	}

	view := m.todoList.View()
	return contentClean.
		Width(m.width).
		Height(height).
		Render(view)
}

func (m Model) renderInputClean() string {
	var prompt string
	var promptStyle lipgloss.Style
	
	// Define prompt styles
	todoPromptStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFA500")).
		Bold(true)
		
	stakPromptStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7D56F4")).
		Bold(true)
		
	searchPromptStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00BFFF")).
		Bold(true)
	
	switch m.currentMode {
	case todoMode:
		prompt = "✓ todo › "
		promptStyle = todoPromptStyle
	case scratchpadMode:
		prompt = "⚡ stak › "
		promptStyle = stakPromptStyle
	case todayMode:
		prompt = "📅 today › "
		promptStyle = stakPromptStyle
	case searchMode:
		prompt = "🔍 search › "
		promptStyle = searchPromptStyle
	default:
		prompt = "› "
		promptStyle = stakPromptStyle
	}

	m.textInput.Prompt = promptStyle.Render(prompt)
	
	return inputClean.
		Width(m.width).
		Render(m.textInput.View())
}

// Minimal status bar
func (m Model) renderStatusClean() string {
	status := "Shift+Tab: toggle mode"
	
	return statusClean.
		Width(m.width).
		Render(status)
}