package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"stak/internal/models"
)

// IRC-style status bar styles
var (
	statusBarStyle = lipgloss.NewStyle().
		Background(lipgloss.Color("#2D2D2D")).
		Foreground(lipgloss.Color("#FFFFFF")).
		Padding(0, 1)

	modeSegmentStyle = lipgloss.NewStyle().
		Background(lipgloss.Color("#5F87AF")).
		Foreground(lipgloss.Color("#FFFFFF")).
		Bold(true).
		Padding(0, 1)

	contextSegmentStyle = lipgloss.NewStyle().
		Background(lipgloss.Color("#875F87")).
		Foreground(lipgloss.Color("#FFFFFF")).
		Padding(0, 1)

	timeSegmentStyle = lipgloss.NewStyle().
		Background(lipgloss.Color("#5F875F")).
		Foreground(lipgloss.Color("#FFFFFF")).
		Padding(0, 1)

	contentClean = lipgloss.NewStyle().
		Padding(1, 2)

	contextBarStyle = lipgloss.NewStyle().
		Background(lipgloss.Color("#3D3D3D")).
		Foreground(lipgloss.Color("#AAAAAA")).
		Padding(0, 1)

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
	contextBarHeight := 1
	statusBarHeight := 1
	inputHeight := 3
	
	contentHeight := totalHeight - contextBarHeight - statusBarHeight - inputHeight
	if contentHeight < 3 {
		contentHeight = 3
	}

	var sections []string

	// 1. Content (fixed height)
	if m.showHelp {
		sections = append(sections, m.renderHelpClean(contentHeight))
	} else if m.currentMode == calendarMode {
		sections = append(sections, m.renderCalendarView(contentHeight))
	} else {
		sections = append(sections, m.renderEntriesClean(contentHeight))
	}

	// 2. IRC-style status bar (directly above input)
	sections = append(sections, m.renderIRCStatusBar())

	// 3. Input
	sections = append(sections, m.renderInputClean())

	// 4. Help bar (bottom)
	sections = append(sections, m.renderHelpBar())

	return strings.Join(sections, "\n")
}

func (m Model) renderIRCStatusBar() string {
	// Mode segment
	var modeText string
	switch m.currentMode {
	case todoMode:
		modeText = "TODO"
	case stakMode:
		modeText = "STAK"
	case calendarMode:
		modeText = "CALENDAR"
	default:
		modeText = "STAK"
	}
	modeSegment := modeSegmentStyle.Render(modeText)

	// Context segment (or error message)
	var contextText string
	var contextSegment string
	
	// Check if we have an error to show (and it's less than 5 seconds old)
	if m.errorMessage != "" && time.Since(m.errorTime) < 5*time.Second {
		// Show error in red
		errorStyle := lipgloss.NewStyle().
			Background(lipgloss.Color("#FF0000")).
			Foreground(lipgloss.Color("#FFFFFF")).
			Padding(0, 1)
		contextSegment = errorStyle.Render(m.errorMessage)
	} else {
		// Clear old error
		if m.errorMessage != "" && time.Since(m.errorTime) >= 5*time.Second {
			// Note: In a real implementation, we'd clear this in the model,
			// but for now just don't show it
		}
		
		// Show normal context
		switch m.currentMode {
		case todoMode:
			if m.editingTodoIdx >= 0 {
				contextText = "EDITING TODO"
			} else {
				completed := 0
				for _, entry := range m.entries {
					if entry.Type == models.TypeTodo && entry.TodoStatus == models.TodoCompleted {
						completed++
					}
				}
				contextText = fmt.Sprintf("%d/%d done", completed, len(m.entries))
			}
		case calendarMode:
			contextText = fmt.Sprintf("%s • %d entries", m.selectedDate.Format("January 2006"), len(m.entries))
		default:
			contextText = fmt.Sprintf("%d entries", len(m.entries))
		}
		contextSegment = contextSegmentStyle.Render(contextText)
	}

	// Time segment
	now := time.Now()
	timeText := now.Format("15:04 Mon Jan 2")
	timeSegment := timeSegmentStyle.Render(timeText)

	// Spacer to push time to the right
	segments := modeSegment + " " + contextSegment
	usedWidth := lipgloss.Width(segments) + lipgloss.Width(timeSegment)
	spacerWidth := m.width - usedWidth - 2 // -2 for padding
	if spacerWidth < 0 {
		spacerWidth = 0
	}
	spacer := statusBarStyle.Width(spacerWidth).Render("")

	return statusBarStyle.Width(m.width).Render(segments + spacer + timeSegment)
}

func (m Model) renderEntriesClean(height int) string {
	if len(m.entries) == 0 {
		var emptyText string
		switch m.currentMode {
		case todoMode:
			emptyText = "No todos yet. Start typing to add one."
		case stakMode:
			emptyText = "No entries yet. Start typing to add one."
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
		if m.currentMode == todoMode && !m.textInput.Focused() {
			// Add visual indicator for navigation mode
			line = "› " + line
		}
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
	// Make sure we have a visible input with proper styling
	inputView := m.textInput.View()
	
	// Highlight border if active pane in calendar mode
	borderColor := "#666666"
	if m.currentMode == calendarMode && m.activePane == inputPane {
		borderColor = "#FFA500"
	}
	
	// Claude Code style rounded border
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(borderColor)).
		Padding(0, 1).
		Width(m.width - 4). // Account for border and padding
		Render(inputView)
}

// Help bar using Bubble Tea help component
func (m Model) renderHelpBar() string {
	return m.help.View(m.keys)
}

// Calendar view with tab-navigable panes
func (m Model) renderCalendarView(height int) string {
	// Check minimum terminal size for calendar mode
	minWidth := 80
	minHeight := 20
	
	if m.width < minWidth || height < minHeight {
		errorText := fmt.Sprintf("Terminal too small. Calendar mode needs at least %dx%d characters.", minWidth, minHeight)
		return contentClean.
			Width(m.width).
			Height(height).
			Align(lipgloss.Center, lipgloss.Center).
			Render(errorText)
	}
	
	// Split the width - 40% left content, 60% right for calendar and notes
	leftWidth := int(float64(m.width) * 0.4)
	rightWidth := m.width - leftWidth - 2 // -2 for spacing and potential borders
	
	if leftWidth < 20 {
		leftWidth = 20
		rightWidth = m.width - 22 // Adjust right width accordingly
	}
	if rightWidth < 30 {
		rightWidth = 30
		leftWidth = m.width - 32 // Adjust left width accordingly
	}
	
	// Ensure widths don't exceed available space
	if leftWidth + rightWidth + 2 > m.width {
		leftWidth = int(float64(m.width-2) * 0.4)
		rightWidth = m.width - leftWidth - 2
	}
	
	// Left side: entries for selected date (with focus indication)
	leftContent := m.renderSelectedDateEntries(leftWidth, height)
	if m.activePane == entriesPane {
		leftContent = m.addFocusBorder(leftContent, leftWidth, height)
	}
	
	// Right side: just the calendar, taking up the full height
	calendarContent := m.renderDatePicker(rightWidth, height)
	if m.activePane == datePickerPane {
		calendarContent = m.addFocusBorder(calendarContent, rightWidth, height)
	}
	
	// No vertical join needed - just use the calendar content
	rightContent := calendarContent
	
	// Combine left and right horizontally
	leftBox := contentClean.Width(leftWidth).Height(height).Render(leftContent)
	// Don't apply additional height constraint to rightBox since rightContent already has proper sizing
	rightBox := lipgloss.NewStyle().Width(rightWidth).Render(rightContent)
	
	return lipgloss.JoinHorizontal(lipgloss.Top, leftBox, rightBox)
}

func (m Model) addFocusBorder(content string, width, height int) string {
	focusStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#FFA500")).
		Width(width).
		Height(height).
		Align(lipgloss.Center, lipgloss.Center)
	return focusStyle.Render(content)
}

// Render calendar grid for current month
func (m Model) renderCalendarGrid(width, height int) string {
	now := m.selectedDate
	year := now.Year()
	month := now.Month()
	
	// Calendar header
	monthName := month.String() + " " + fmt.Sprintf("%d", year)
	header := lipgloss.NewStyle().Bold(true).Align(lipgloss.Center).Render(monthName)
	
	// Days of week header
	daysHeader := "Su Mo Tu We Th Fr Sa"
	
	// Get first day of month and number of days
	firstDay := time.Date(year, month, 1, 0, 0, 0, 0, now.Location())
	lastDay := firstDay.AddDate(0, 1, -1)
	daysInMonth := lastDay.Day()
	startWeekday := int(firstDay.Weekday())
	
	// Build calendar grid
	var lines []string
	lines = append(lines, header)
	lines = append(lines, daysHeader)
	
	// Calendar days
	var currentLine []string
	
	// Add empty spaces for days before month starts
	for i := 0; i < startWeekday; i++ {
		currentLine = append(currentLine, "  ")
	}
	
	// Add days of the month
	for day := 1; day <= daysInMonth; day++ {
		dayStr := fmt.Sprintf("%2d", day)
		
		// Check if this day has entries
		dayDate := time.Date(year, month, day, 0, 0, 0, 0, now.Location())
		dateKey := dayDate.Format("2006-01-02")
		if entries, exists := m.calendarEntries[dateKey]; exists && len(entries) > 0 {
			dayStr = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFA500")).Bold(true).Render(dayStr)
		}
		
		// Highlight selected day
		if day == now.Day() {
			dayStr = lipgloss.NewStyle().Background(lipgloss.Color("#444444")).Render(dayStr)
		}
		
		currentLine = append(currentLine, dayStr)
		
		// Start new line after Saturday
		if len(currentLine) == 7 {
			lines = append(lines, strings.Join(currentLine, " "))
			currentLine = []string{}
		}
	}
	
	// Add remaining days if needed
	if len(currentLine) > 0 {
		lines = append(lines, strings.Join(currentLine, " "))
	}
	
	return strings.Join(lines, "\n")
}

// Render entries for the currently selected date
func (m Model) renderSelectedDateEntries(width, height int) string {
	selectedDateStr := m.selectedDate.Format("Monday, January 2, 2006")
	
	header := lipgloss.NewStyle().Bold(true).Render(selectedDateStr)
	
	if len(m.entries) == 0 {
		emptyText := "No entries for this date"
		content := header + "\n\n" + lipgloss.NewStyle().Faint(true).Render(emptyText)
		return m.wrapEntriesContent(content, width, height)
	}
	
	var entryLines []string
	entryLines = append(entryLines, header, "")
	
	for i, entry := range m.entries {
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
		
		// Highlight selected entry (if any)
		if i == m.selectedIdx {
			line = selectedEntryClean.Render(line)
		}
		
		entryLines = append(entryLines, line)
	}
	
	content := strings.Join(entryLines, "\n")
	return m.wrapEntriesContent(content, width, height)
}

// Helper function to wrap entries content with proper sizing
func (m Model) wrapEntriesContent(content string, width, height int) string {
	// Adjust sizing based on whether focus border will be applied
	actualWidth := width - 2  // Account for potential border
	actualHeight := height - 2
	if m.activePane == entriesPane {
		actualWidth = width - 4 // Extra space for focus border
		actualHeight = height - 4
	}
	
	return lipgloss.NewStyle().
		Width(actualWidth).
		Height(actualHeight).
		Padding(1, 1).
		Render(content)
}

// Render the datepicker component
func (m Model) renderDatePicker(width, height int) string {
	header := lipgloss.NewStyle().Bold(true).Render("Calendar")
	
	// Get the datepicker view
	pickerView := m.datePicker.View()
	
	content := header + "\n\n" + pickerView
	
	// Adjust sizing based on whether focus border will be applied
	actualWidth := width - 2  // Account for potential border
	actualHeight := height - 2
	if m.activePane == datePickerPane {
		actualWidth = width - 4 // Extra space for focus border
		actualHeight = height - 4
	}
	
	return lipgloss.NewStyle().
		Width(actualWidth).
		Height(actualHeight).
		Padding(1, 1).
		Render(content)
}

