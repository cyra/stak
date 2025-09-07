package ui

import (
	"fmt"
	"strings"
	"time"

	"stak/internal/models"

	"github.com/charmbracelet/lipgloss"
)

// Status bar styles following Lip Gloss example
var (
	statusNugget = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Padding(0, 1)

	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#343433", Dark: "#C1C6B2"}).
			Background(lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#353533"})

	statusStyle = lipgloss.NewStyle().
			Inherit(statusBarStyle).
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#FF5F87")).
			Padding(0, 1).
			MarginRight(1)

	timeStyle = statusNugget.
			Background(lipgloss.Color("#A550DF")).
			Align(lipgloss.Right)

	statusText = lipgloss.NewStyle().Inherit(statusBarStyle)

	contentClean = lipgloss.NewStyle().
			Padding(1, 2)


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

	// 1. Content (fixed height) - apply consistent borders
	if m.showHelp {
		content := m.renderHelpClean()
		sections = append(sections, m.addConsistentBorder(content, m.width, contentHeight, false))
	} else if m.currentMode == calendarMode {
		sections = append(sections, m.renderCalendarView(contentHeight))
	} else {
		// For STAK and TODO modes, apply border to the main content area
		content := m.renderEntriesClean()
		isFocused := m.currentMode == todoMode && !m.textInput.Focused() // Focused when navigating todos
		sections = append(sections, m.addConsistentBorder(content, m.width, contentHeight, isFocused))
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
	// Status key (mode)
	var statusKey string
	switch m.currentMode {
	case todoMode:
		if m.editingTodoIdx >= 0 {
			statusKey = "EDITING"
		} else {
			statusKey = "TODO"
		}
	case stakMode:
		statusKey = "STAK"
	case calendarMode:
		statusKey = "CALENDAR"
	default:
		statusKey = "STAK"
	}

	// Context information
	var contextText string
	switch m.currentMode {
	case todoMode:
		if m.editingTodoIdx < 0 {
			completed := 0
			for _, entry := range m.entries {
				if entry.Type == models.TypeTodo && entry.TodoStatus == models.TodoCompleted {
					completed++
				}
			}
			contextText = fmt.Sprintf("%d/%d completed", completed, len(m.entries))
		} else {
			contextText = "Editing todo item"
		}
	case stakMode:
		today := time.Now().Format("2006-01-02")
		contextText = fmt.Sprintf("%s.md • %d entries", today, len(m.entries))
	case calendarMode:
		var paneText string
		switch m.activePane {
		case inputPane:
			paneText = "INPUT"
		case entriesPane:
			paneText = "NOTES"
		case datePickerPane:
			paneText = "CALENDAR"
		}
		contextText = fmt.Sprintf("%s • %s", m.selectedDate.Format("January 2006"), paneText)
	default:
		contextText = fmt.Sprintf("%d entries", len(m.entries))
	}

	// Time or error
	var timeText string
	if m.errorMessage != "" && time.Since(m.errorTime) < 5*time.Second {
		timeText = m.errorMessage
	} else {
		now := time.Now()
		timeText = now.Format("15:04")
	}

	// Render components following Lip Gloss example pattern
	w := lipgloss.Width

	statusKeyRendered := statusStyle.Render(statusKey)
	timeRendered := timeStyle.Render(timeText)
	contextRendered := statusText.
		Width(m.width - w(statusKeyRendered) - w(timeRendered)).
		Render(contextText)

	// Join horizontally like in the example
	bar := lipgloss.JoinHorizontal(lipgloss.Top,
		statusKeyRendered,
		contextRendered,
		timeRendered,
	)

	return statusBarStyle.Width(m.width).Render(bar)
}

func (m Model) renderEntriesClean() string {
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

		// Don't apply sizing here - let the border function handle it
		return emptyText
	}

	// Show entries IRC/chat style - oldest at top, newest at bottom
	var renderedEntries []string

	// Render all entries - let the border function handle height constraints
	for i := 0; i < len(m.entries); i++ {
		entry := m.entries[i]
		selected := (i == m.selectedIdx)
		content := m.renderEntryClean(entry, selected)
		renderedEntries = append(renderedEntries, content)
	}

	entriesText := strings.Join(renderedEntries, "\n")

	// Don't apply sizing here - let the border function handle it
	return entriesText
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

func (m Model) renderHelpClean() string {
	help := strings.Join(m.commands, "\n")
	// Don't apply sizing here - let the border function handle it
	return help
}


func (m Model) renderInputClean() string {
	// Make sure we have a visible input with proper styling
	inputView := m.textInput.View()

	// Highlight border if active pane in calendar mode
	borderColor := "#666666"
	if m.currentMode == calendarMode && m.activePane == inputPane {
		borderColor = "#FFA500"
	} else if m.textInput.Focused() {
		borderColor = "#00FF00" // Green for focused input
	}

	// Claude Code style rounded border
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(borderColor)).
		Padding(0, 1).
		Width(m.width - 2). // Account for border and padding
		Render(inputView)
}

// Help bar using Bubble Tea help component
func (m Model) renderHelpBar() string {
	return m.help.View(m.keys)
}

// Calendar view with tab-navigable panes
func (m Model) renderCalendarView(height int) string {
	// Check minimum terminal size for calendar mode
	minWidth := 100
	minHeight := 20

	if m.width < minWidth || height < minHeight {
		errorText := fmt.Sprintf("Terminal too small. Calendar mode needs at least %dx%d characters.", minWidth, minHeight)
		return contentClean.
			Width(m.width).
			Height(height).
			Align(lipgloss.Center, lipgloss.Center).
			Render(errorText)
	}

	// Fixed width for calendar pane - make it a bit wider for better appearance
	// Account for two separate borders: each takes 4 characters (2 border + 2 padding)
	availableWidth := m.width // 4 for left border + 4 for right border
	calendarFixedWidth := 40  // Fixed width for calendar pane
	leftWidth := availableWidth - calendarFixedWidth

	// Ensure minimum width for notes pane
	if leftWidth < 40 {
		leftWidth = 40
		calendarFixedWidth = availableWidth - leftWidth
	}

	// Left side: entries for selected date (with focus indication)
	leftContent := m.renderSelectedDateEntries()
	leftContent = m.addConsistentBorder(leftContent, leftWidth, height, m.activePane == entriesPane)

	// Right side: just the calendar, taking up the full height
	calendarContent := m.renderDatePicker()
	calendarContent = m.addConsistentBorder(calendarContent, calendarFixedWidth, height, m.activePane == datePickerPane)

	// Manually combine left and right content to eliminate gaps
	// Split both contents into lines and combine them line by line
	leftLines := strings.Split(leftContent, "\n")
	rightLines := strings.Split(calendarContent, "\n")

	// Ensure both have the same number of lines
	maxLines := len(leftLines)
	if len(rightLines) > maxLines {
		maxLines = len(rightLines)
	}

	var combinedLines []string
	for i := 0; i < maxLines; i++ {
		leftLine := ""
		rightLine := ""

		if i < len(leftLines) {
			leftLine = leftLines[i]
		}
		if i < len(rightLines) {
			rightLine = rightLines[i]
		}

		// Combine lines directly without any spacing
		combinedLines = append(combinedLines, leftLine+rightLine)
	}

	return strings.Join(combinedLines, "\n")
}

func (m Model) addConsistentBorder(content string, width, height int, isFocused bool) string {
	// Always apply border to maintain consistent dimensions
	borderColor := "#444444" // Darker gray border for better visibility
	if isFocused {
		borderColor = "#FFA500" // Orange border when focused
	}

	// Ensure content fills the full height by padding it to the required height
	contentLines := strings.Split(content, "\n")
	requiredLines := height - 4 // Account for border (2) + padding (2)

	// Safety check: ensure requiredLines is at least 1
	if requiredLines < 1 {
		requiredLines = 1
	}

	// If content has fewer lines than required, pad with empty lines
	for len(contentLines) < requiredLines {
		contentLines = append(contentLines, "")
	}

	// If content has more lines than required, truncate
	if len(contentLines) > requiredLines {
		contentLines = contentLines[:requiredLines]
	}

	paddedContent := strings.Join(contentLines, "\n")

	// Apply border with proper dimensions
	// The width parameter should be the content width (terminal width - border space)
	// RoundedBorder adds 2 characters, Padding(1,1) adds 2 more = 4 total
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(borderColor)).
		Width(width-2).   // Content width (terminal width - border/padding)
		Height(height-4). // Content height (terminal height - border/padding)
		Padding(1, 1).    // Padding inside the border
		Render(paddedContent)
}


// Render calendar grid for current month
func (m Model) renderCalendarGrid() string {
	now := m.selectedDate
	year := now.Year()
	month := now.Month()

	// Calendar header with better spacing for wider pane
	monthName := month.String() + " " + fmt.Sprintf("%d", year)
	header := lipgloss.NewStyle().Bold(true).Align(lipgloss.Center).Render(monthName)

	// Days of week header with better spacing
	daysHeader := " Su Mo Tu We Th Fr Sa "

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
		currentLine = append(currentLine, "   ") // Match the spacing of days
	}

	// Add days of the month with better spacing
	for day := 1; day <= daysInMonth; day++ {
		dayStr := fmt.Sprintf(" %2d", day) // Add space before day for better alignment

		// Check if this day has entries
		dayDate := time.Date(year, month, day, 0, 0, 0, 0, now.Location())
		dateKey := dayDate.Format("2006-01-02")
		hasEntries := false
		if entries, exists := m.calendarEntries[dateKey]; exists && len(entries) > 0 {
			hasEntries = true
		}

		// Apply styling based on selection and entries
		if day == now.Day() {
			// Selected day - use orange background with black text
			dayStr = lipgloss.NewStyle().
				Background(lipgloss.Color("#FFA500")).
				Foreground(lipgloss.Color("#000000")).
				Bold(true).
				Render(dayStr)
		} else if hasEntries {
			// Days with entries - use orange text
			dayStr = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFA500")).
				Bold(true).
				Render(dayStr)
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

	// Ensure we always have 6 weeks (42 days) for consistent height
	// Add empty lines if needed to maintain consistent calendar height
	for len(lines) < 8 { // 1 header + 1 days header + 6 weeks
		lines = append(lines, "                    ") // Empty line with consistent width
	}

	return strings.Join(lines, "\n")
}

// Render entries for the currently selected date
func (m Model) renderSelectedDateEntries() string {
	selectedDateStr := m.selectedDate.Format("Monday, January 2, 2006")

	header := lipgloss.NewStyle().Bold(true).Render(selectedDateStr)

	if len(m.entries) == 0 {
		emptyText := "No entries for this date"
		content := header + "\n\n" + lipgloss.NewStyle().Faint(true).Render(emptyText)
		return content
	}

	var entryLines []string
	entryLines = append(entryLines, header, "")

	// Render all entries - let the border function handle height constraints
	for i := 0; i < len(m.entries); i++ {
		entry := m.entries[i]
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
	// Don't apply sizing here - let the border function handle it
	return content
}

// Helper function to wrap entries content with consistent sizing

// Render the datepicker component
func (m Model) renderDatePicker() string {
	header := lipgloss.NewStyle().Bold(true).Render("Calendar")

	// Use our custom calendar grid with highlighting
	calendarGrid := m.renderCalendarGrid() // -2 for header spacing

	content := header + "\n\n" + calendarGrid

	// Don't apply sizing here - let the border function handle it
	return content
}
