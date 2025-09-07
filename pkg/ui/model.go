package ui

import (
	"fmt"
	"strings"
	"time"

	"stak/internal/application"
	"stak/internal/config"
	"stak/internal/models"
	"stak/pkg/categorizer"
	"stak/pkg/extractor"
	"stak/pkg/search"
	"stak/pkg/storage"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type mode int

const (
	stakMode mode = iota // Renamed from scratchpadMode
	todoMode
	calendarMode
)

type calendarPane int

const (
	inputPane calendarPane = iota
	entriesPane
	datePickerPane
)

// Key bindings for help
type keyMap struct {
	Up       key.Binding
	Down     key.Binding
	Tab      key.Binding
	ShiftTab key.Binding
	Enter    key.Binding
	Edit     key.Binding
	Quit     key.Binding
	Help     key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Tab, k.ShiftTab},
		{k.Enter, k.Edit, k.Help, k.Quit},
	}
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up"),
		key.WithHelp("↑", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down"),
		key.WithHelp("↓", "down"),
	),
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next pane"),
	),
	ShiftTab: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "toggle mode"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select/toggle"),
	),
	Edit: key.NewBinding(
		key.WithKeys("e"),
		key.WithHelp("e", "edit todo"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c", "q"),
		key.WithHelp("ctrl+c/q", "quit"),
	),
}

type Model struct {
	config        *config.Config
	storage       *storage.Storage // Keep for direct access needs
	entryService  *application.EntryService
	textInput     textinput.Model
	todoList      *TodoListModel
	entries       []models.Entry
	currentMode   mode
	width         int
	height        int
	ready         bool
	commands      []string
	slashCommands []string
	selectedIdx   int
	searchQuery   string
	showHelp      bool
	// Calendar mode fields
	selectedDate    time.Time
	calendarEntries map[string][]models.Entry // date -> entries
	activePane      calendarPane              // for tab navigation in calendar mode
	help            help.Model
	keys            keyMap
	// TODO editing state
	editingTodoIdx  int    // -1 when not editing
	originalContent string // backup for cancel
	// Error handling
	errorMessage string    // Error message to show in status bar
	errorTime    time.Time // When error was shown
}

func NewModel() *Model {
	cfg := config.DefaultConfig()
	return NewModelWithConfig(cfg)
}

func NewModelWithConfig(cfg *config.Config) *Model {
	// Create dependencies
	storage := storage.New(cfg)
	categoriser := categorizer.New()
	searcher := search.NewFuzzySearcher()
	extractor := extractor.NewLinkExtractor()

	// Create application service
	entryService := application.NewEntryService(storage, categoriser, extractor, searcher)

	ti := textinput.New()
	ti.Placeholder = "Enter your thoughts, links, todos..."
	ti.Focus()
	ti.CharLimit = 500
	ti.Width = 50
	ti.ShowSuggestions = true

	// No longer using external datepicker

	// Initialize help
	h := help.New()
	h.ShowAll = false // Start with short help

	model := &Model{
		config:       cfg,
		storage:      storage,
		entryService: entryService,
		textInput:    ti,
		entries:      []models.Entry{},
		currentMode:  stakMode,
		commands: []string{
			"Shift+Tab - Toggle between STAK and TODO mode",
			"/todos - Switch to TODO mode",
			"/cal - Calendar view with date picker",
			"/help - Show this help",
			"/quit - Exit stak",
		},
		slashCommands: []string{
			"/todos",
			"/cal",
			"/help",
			"/quit",
		},
		selectedIdx: -1,
		// Initialize calendar fields
		selectedDate:    time.Now(),
		calendarEntries: make(map[string][]models.Entry),
		activePane:      inputPane,
		help:            h,
		keys:            keys,
		editingTodoIdx:  -1, // Not editing by default
	}

	model.updatePrompt() // Set initial prompt
	return model
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		m.loadFilteredEntries(),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Ensure minimum width for text input
		inputWidth := msg.Width - 4
		if inputWidth < 20 {
			inputWidth = 20
		}
		m.textInput.Width = inputWidth

		// Update todoList dimensions if it exists
		if m.todoList != nil {
			m.todoList.SetSize(msg.Width, msg.Height-3) // Account for header and footer
		}

		m.ready = true

	case tea.KeyMsg:
		// Check for help key first
		if key.Matches(msg, m.keys.Help) {
			m.help.ShowAll = !m.help.ShowAll
			return m, nil
		}

		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit

		case tea.KeyEsc:
			if m.showHelp {
				m.showHelp = false
				return m, nil
			}
			if m.currentMode == calendarMode {
				m.currentMode = stakMode
				return m, m.loadTodayEntries()
			}

		case tea.KeyShiftTab:
			// Cycle through all 3 modes: STAK → TODO → CALENDAR → STAK
			switch m.currentMode {
			case stakMode:
				m.currentMode = todoMode
			case todoMode:
				m.currentMode = calendarMode
				m.activePane = inputPane // Reset pane navigation
				m.textInput.Focus()      // Make sure input is focused
			case calendarMode:
				m.currentMode = stakMode
				m.activePane = inputPane // Reset pane navigation
				m.textInput.Focus()      // Make sure input is focused
			}
			m.selectedIdx = -1
			m.showHelp = false // Close help dialog when switching modes
			m.updatePrompt()   // Update the prompt when mode changes

			// Load appropriate entries for the new mode
			if m.currentMode == calendarMode {
				return m, m.loadCalendarEntries()
			} else {
				return m, m.loadFilteredEntries()
			}

		case tea.KeyEnter:
			if m.showHelp {
				m.showHelp = false
				return m, nil
			}

			// Handle enter in calendar mode
			if m.currentMode == calendarMode {
				switch m.activePane {
				case inputPane:
					// Handle input commands/entries
					return m.handleEnter()
				case entriesPane:
					// In entries pane, enter does nothing (just browsing)
					return m, nil
				case datePickerPane:
					// In calendar pane, enter could select the date (already selected)
					// or we could add a command to jump to input
					m.activePane = inputPane
					m.textInput.Focus()
					return m, nil
				}
			}

			// Handle enter in TODO mode
			if m.currentMode == todoMode {
				// If editing a todo, save the changes
				if m.editingTodoIdx >= 0 && m.editingTodoIdx < len(m.entries) {
					return m.saveEditingTodo()
				}
				// If a todo is selected (not in input), toggle it
				if !m.textInput.Focused() && m.selectedIdx >= 0 && m.selectedIdx < len(m.entries) {
					return m.toggleTodo()
				}
			}

			return m.handleEnter()

		case tea.KeyUp:
			if m.currentMode == calendarMode {
				// Handle up arrow in calendar mode based on active pane
				switch m.activePane {
				case entriesPane:
					// Navigate through notes entries
					if m.selectedIdx > 0 {
						m.selectedIdx--
					}
				case datePickerPane:
					// Navigate calendar dates spatially (up = day above in grid)
					m.selectedDate = m.getSpatialDate(m.selectedDate, "up")
					// Update entries for new date
					cmds = append(cmds, m.loadEntriesForDate(m.selectedDate))
				}
			} else {
				// Normal up arrow behavior for other modes
				if m.selectedIdx > 0 {
					m.selectedIdx--
				}
			}

		case tea.KeyDown:
			if m.currentMode == calendarMode {
				// Handle down arrow in calendar mode based on active pane
				switch m.activePane {
				case entriesPane:
					// Navigate through notes entries
					if m.selectedIdx < len(m.entries)-1 {
						m.selectedIdx++
					}
				case datePickerPane:
					// Navigate calendar dates spatially (down = day below in grid)
					m.selectedDate = m.getSpatialDate(m.selectedDate, "down")
					// Update entries for new date
					cmds = append(cmds, m.loadEntriesForDate(m.selectedDate))
				}
			} else {
				// Normal down arrow behavior for other modes
				if m.selectedIdx < len(m.entries)-1 {
					m.selectedIdx++
				}
			}

		case tea.KeyLeft:
			if m.currentMode == calendarMode && m.activePane == datePickerPane {
				// Navigate calendar dates spatially (left = day to the left in grid)
				m.selectedDate = m.getSpatialDate(m.selectedDate, "left")
				// Update entries for new date
				cmds = append(cmds, m.loadEntriesForDate(m.selectedDate))
			}

		case tea.KeyRight:
			if m.currentMode == calendarMode && m.activePane == datePickerPane {
				// Navigate calendar dates spatially (right = day to the right in grid)
				m.selectedDate = m.getSpatialDate(m.selectedDate, "right")
				// Update entries for new date
				cmds = append(cmds, m.loadEntriesForDate(m.selectedDate))
			}

		case tea.KeyTab:
			if m.currentMode == calendarMode {
				// Cycle through panes in calendar mode: Input → Entries → DatePicker → Input
				switch m.activePane {
				case inputPane:
					m.activePane = entriesPane
					m.textInput.Blur()
					// Set selectedIdx to last entry if available
					if len(m.entries) > 0 && m.selectedIdx < 0 {
						m.selectedIdx = len(m.entries) - 1
					}
				case entriesPane:
					m.activePane = datePickerPane
					// Clear entry selection when moving to calendar
					m.selectedIdx = -1
				case datePickerPane:
					m.activePane = inputPane
					m.textInput.Focus()
					// Clear entry selection when moving to input
					m.selectedIdx = -1
				}
				return m, nil
			} else if m.currentMode == todoMode {
				// In TODO mode, tab switches between input and todo list navigation
				if m.textInput.Focused() {
					m.textInput.Blur()
					// Focus on todo list - set selectedIdx if not already set
					if len(m.entries) > 0 && m.selectedIdx < 0 {
						m.selectedIdx = 0
					}
				} else {
					m.textInput.Focus()
					m.selectedIdx = -1 // Clear todo selection
				}
				return m, nil
			}
			// STAK mode: Tab does nothing (could add basic completion later)

		default:
			// Handle special keys in TODO mode
			if m.currentMode == todoMode && !m.textInput.Focused() && m.selectedIdx >= 0 && m.selectedIdx < len(m.entries) {
				switch msg.String() {
				case "e":
					// Enter edit mode
					return m.startEditingTodo()
				case "right":
					// TODO: Show context menu (edit/delete)
					// For now, just start editing
					return m.startEditingTodo()
				}
			}

			// Handle escape in edit mode
			if m.currentMode == todoMode && m.editingTodoIdx >= 0 && msg.String() == "esc" {
				return m.cancelEditingTodo()
			}
		}

	case entriesLoadedMsg:
		m.entries = msg.entries
		if len(m.entries) > 0 && m.selectedIdx < 0 {
			m.selectedIdx = len(m.entries) - 1
		}

	case filteredEntriesLoadedMsg:
		// Only update if the mode matches current mode (avoid race conditions)
		if msg.mode == m.currentMode {
			m.entries = msg.entries
			if len(m.entries) > 0 && m.selectedIdx < 0 {
				m.selectedIdx = len(m.entries) - 1
			}
		}

	case entryAddedMsg:
		if m.currentMode == calendarMode {
			// In calendar mode, reload entries for the selected date
			cmds = append(cmds, m.loadEntriesForDate(m.selectedDate))
		} else {
			// In other modes, reload filtered entries
			cmds = append(cmds, m.loadFilteredEntries())
		}

	case calendarEntriesLoadedMsg:
		m.calendarEntries = msg.calendarEntries
		m.selectedDate = msg.selectedDate
		// Set entries for the selected date
		dateKey := m.selectedDate.Format("2006-01-02")
		if dayEntries, exists := m.calendarEntries[dateKey]; exists {
			m.entries = dayEntries
		} else {
			m.entries = []models.Entry{}
		}

	case todoToggledMsg:
		// Save the toggled todo entry
		if err := m.storage.SaveEntry(msg.entry); err == nil {
			// Refresh the current view
			cmds = append(cmds, m.loadFilteredEntries())
		}
	}

	// Remove old todoListMode handling - now integrated into todoMode

	// No longer using external datepicker - using custom calendar grid instead

	// Handle text input updates (only if input pane is active in calendar mode)
	if m.currentMode != calendarMode || m.activePane == inputPane {
		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(msg)
		cmds = append(cmds, cmd)
	}

	// Handle autocomplete for slash commands
	m.updateSuggestions()

	return m, tea.Batch(cmds...)
}

func (m Model) handleEnter() (tea.Model, tea.Cmd) {
	input := strings.TrimSpace(m.textInput.Value())
	if input == "" {
		return m, nil
	}

	if strings.HasPrefix(input, "/") {
		return m.handleCommand(input)
	}

	if strings.HasPrefix(input, "tomorrow") {
		return m.addTomorrowEntry(input)
	}

	return m.addEntry(input)
}

func (m Model) handleCommand(cmd string) (tea.Model, tea.Cmd) {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return m, nil
	}

	command := parts[0]

	switch command {
	case "/quit", "/q":
		return m, tea.Quit

	case "/help", "/h":
		m.showHelp = !m.showHelp
		m.textInput.SetValue("")
		return m, nil

	case "/stak":
		m.currentMode = stakMode
		m.textInput.SetValue("")
		return m, m.loadFilteredEntries()

	case "/cal":
		m.currentMode = calendarMode
		m.activePane = inputPane
		m.textInput.SetValue("")
		m.textInput.Focus()
		return m, m.loadCalendarEntries()

	case "/todos":
		m.currentMode = todoMode
		m.textInput.SetValue("")
		return m, m.loadFilteredEntries()

	case "/todo", "/t":
		// Add todo without switching modes
		args := parts[1:] // Get the text after the command
		if len(args) > 0 {
			todoText := strings.Join(args, " ")
			// Force as todo type
			todoType := models.TypeTodo
			var err error
			if m.currentMode == calendarMode {
				// In calendar mode, create todo for the selected date
				_, err = m.entryService.CreateEntryForDate(todoText, m.selectedDate, &todoType)
			} else {
				// In other modes, create todo for today
				_, err = m.entryService.CreateEntry(todoText, &todoType)
			}
			if err != nil {
				// TODO: Show error in status bar
			}
			m.textInput.SetValue("")
			if m.currentMode == calendarMode {
				return m, m.loadEntriesForDate(m.selectedDate)
			} else {
				return m, m.loadFilteredEntries()
			}
		}
		// If no text provided, show error
		m.errorMessage = "Usage: /todo <text> or /t <text>"
		m.errorTime = time.Now()
		return m, nil

	default:
		// Unknown command
		m.errorMessage = fmt.Sprintf("Unknown command: %s", command)
		m.errorTime = time.Now()
	}

	m.textInput.SetValue("")
	return m, nil
}

func (m Model) addEntry(content string) (tea.Model, tea.Cmd) {
	// Use application service for business logic
	var forceType *models.EntryType
	if m.currentMode == todoMode {
		todoType := models.TypeTodo
		forceType = &todoType
	}

	var err error
	if m.currentMode == calendarMode {
		// In calendar mode, create entry for the selected date
		_, err = m.entryService.CreateEntryForDate(content, m.selectedDate, forceType)
	} else {
		// In other modes, create entry for today
		_, err = m.entryService.CreateEntry(content, forceType)
	}

	if err != nil {
		return m, nil
	}

	m.textInput.SetValue("")
	return m, func() tea.Msg {
		return entryAddedMsg{}
	}
}

func (m Model) toggleTodo() (tea.Model, tea.Cmd) {
	if m.selectedIdx < 0 || m.selectedIdx >= len(m.entries) {
		return m, nil
	}

	entry := &m.entries[m.selectedIdx]
	if entry.Type != models.TypeTodo {
		return m, nil
	}

	if entry.TodoStatus == models.TodoPending {
		entry.TodoStatus = models.TodoCompleted
	} else {
		entry.TodoStatus = models.TodoPending
	}

	entry.UpdatedAt = time.Now()
	if err := m.storage.SaveEntry(entry); err != nil {
		return m, nil
	}

	return m, m.loadFilteredEntries()
}

func (m *Model) Storage() *storage.Storage {
	return m.storage
}

func (m *Model) updatePrompt() {
	m.textInput.Prompt = "> "
}

func (m *Model) updateSuggestions() {
	input := m.textInput.Value()

	// Only show suggestions when input starts with "/"
	if strings.HasPrefix(input, "/") {
		// Filter slash commands based on current input
		var suggestions []string
		for _, cmd := range m.slashCommands {
			if strings.HasPrefix(cmd, input) {
				suggestions = append(suggestions, cmd)
			}
		}
		m.textInput.SetSuggestions(suggestions)
	} else {
		// Clear suggestions when not typing a slash command
		m.textInput.SetSuggestions([]string{})
	}
}

func (m Model) addTomorrowEntry(content string) (tea.Model, tea.Cmd) {
	if err := m.entryService.CreateTomorrowEntry(content); err != nil {
		return m, nil
	}

	m.textInput.SetValue("")
	return m, func() tea.Msg {
		return entryAddedMsg{}
	}
}

// Load todos from the past week
func (m Model) loadWeekTodos() ([]models.Entry, error) {
	now := time.Now()
	weekAgo := now.AddDate(0, 0, -7)

	// For now, load today's entries and filter by date
	// TODO: Extend storage service to support date range queries
	allEntries, err := m.storage.LoadTodayEntries()
	if err != nil {
		return []models.Entry{}, err
	}

	var weekEntries []models.Entry
	for _, entry := range allEntries {
		if entry.Type == models.TypeTodo && entry.CreatedAt.After(weekAgo) {
			weekEntries = append(weekEntries, entry)
		}
	}

	return weekEntries, nil
}

// Load todos from the past month
func (m Model) loadMonthTodos() ([]models.Entry, error) {
	now := time.Now()
	monthAgo := now.AddDate(0, -1, 0)

	// For now, load today's entries and filter by date
	// TODO: Extend storage service to support date range queries
	allEntries, err := m.storage.LoadTodayEntries()
	if err != nil {
		return []models.Entry{}, err
	}

	var monthEntries []models.Entry
	for _, entry := range allEntries {
		if entry.Type == models.TypeTodo && entry.CreatedAt.After(monthAgo) {
			monthEntries = append(monthEntries, entry)
		}
	}

	return monthEntries, nil
}

// Start editing a selected todo
func (m Model) startEditingTodo() (tea.Model, tea.Cmd) {
	if m.selectedIdx < 0 || m.selectedIdx >= len(m.entries) {
		return m, nil
	}

	entry := &m.entries[m.selectedIdx]
	if entry.Type != models.TypeTodo {
		return m, nil
	}

	m.editingTodoIdx = m.selectedIdx
	m.originalContent = entry.Content
	m.textInput.SetValue(entry.Content)
	m.textInput.Focus()

	return m, nil
}

// Save editing todo
func (m Model) saveEditingTodo() (tea.Model, tea.Cmd) {
	if m.editingTodoIdx < 0 || m.editingTodoIdx >= len(m.entries) {
		return m, nil
	}

	newContent := strings.TrimSpace(m.textInput.Value())
	if newContent == "" {
		// Don't save empty content, cancel instead
		return m.cancelEditingTodo()
	}

	// Update the entry
	entry := &m.entries[m.editingTodoIdx]
	entry.Content = newContent
	entry.UpdatedAt = time.Now()

	// Save to storage
	if err := m.storage.SaveEntry(entry); err != nil {
		// TODO: Show error message
		return m.cancelEditingTodo()
	}

	// Exit edit mode
	m.editingTodoIdx = -1
	m.originalContent = ""
	m.textInput.SetValue("")
	m.textInput.Blur()

	return m, m.loadFilteredEntries()
}

// Cancel editing todo
func (m Model) cancelEditingTodo() (tea.Model, tea.Cmd) {
	m.editingTodoIdx = -1
	m.originalContent = ""
	m.textInput.SetValue("")
	m.textInput.Blur()
	return m, nil
}

// getSpatialDate calculates the date that would be in the specified direction
// from the current date in the calendar grid layout
func (m Model) getSpatialDate(currentDate time.Time, direction string) time.Time {
	year := currentDate.Year()
	month := currentDate.Month()
	day := currentDate.Day()

	// Get the first day of the month and calculate the grid position
	firstDay := time.Date(year, month, 1, 0, 0, 0, 0, currentDate.Location())
	startWeekday := int(firstDay.Weekday()) // 0 = Sunday, 1 = Monday, etc.

	// Calculate the current day's position in the grid
	// Grid position: (week, day_of_week)
	currentWeek := (day - 1 + startWeekday) / 7
	currentDayOfWeek := (day - 1 + startWeekday) % 7

	var newWeek, newDayOfWeek int

	switch direction {
	case "up":
		newWeek = currentWeek - 1
		newDayOfWeek = currentDayOfWeek
	case "down":
		newWeek = currentWeek + 1
		newDayOfWeek = currentDayOfWeek
	case "left":
		newWeek = currentWeek
		newDayOfWeek = currentDayOfWeek - 1
		if newDayOfWeek < 0 {
			newWeek--
			newDayOfWeek = 6 // Saturday
		}
	case "right":
		newWeek = currentWeek
		newDayOfWeek = currentDayOfWeek + 1
		if newDayOfWeek > 6 {
			newWeek++
			newDayOfWeek = 0 // Sunday
		}
	default:
		return currentDate
	}

	// Calculate the day number from the grid position
	newDay := newWeek*7 + newDayOfWeek - startWeekday + 1

	// Handle month boundaries
	if newDay < 1 {
		// Go to previous month
		prevMonth := month - 1
		prevYear := year
		if prevMonth < 1 {
			prevMonth = 12
			prevYear--
		}
		// Get the last day of the previous month
		lastDayOfPrevMonth := time.Date(prevYear, prevMonth+1, 0, 0, 0, 0, 0, currentDate.Location()).Day()
		newDay = lastDayOfPrevMonth + newDay
		month = prevMonth
		year = prevYear
	} else {
		// Check if we're going beyond the current month
		lastDayOfMonth := time.Date(year, month+1, 0, 0, 0, 0, 0, currentDate.Location()).Day()
		if newDay > lastDayOfMonth {
			// Go to next month
			newDay = newDay - lastDayOfMonth
			month++
			if month > 12 {
				month = 1
				year++
			}
		}
	}

	return time.Date(year, month, newDay, 0, 0, 0, 0, currentDate.Location())
}
