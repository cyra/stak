package ui

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"stak/internal/config"
	"stak/internal/models"
	"stak/pkg/categorizer"
	"stak/pkg/extractor"
	"stak/pkg/search"
	"stak/pkg/storage"
)

type mode int

const (
	scratchpadMode mode = iota
	todoMode
	todoListMode
	searchMode
	todayMode
)

type Model struct {
	config       *config.Config
	storage      *storage.Storage
	categoriser  *categorizer.Categoriser
	searcher     *search.FuzzySearcher
	extractor    *extractor.LinkExtractor
	textInput    textinput.Model
	todoList     *TodoListModel
	entries      []models.Entry
	currentMode  mode
	width        int
	height       int
	ready        bool
	commands      []string
	slashCommands []string
	selectedIdx  int
	searchQuery  string
	showHelp     bool
}

func NewModel() *Model {
	cfg := config.DefaultConfig()
	return NewModelWithConfig(cfg)
}

func NewModelWithConfig(cfg *config.Config) *Model {
	storage := storage.New(cfg)
	categoriser := categorizer.New()
	searcher := search.NewFuzzySearcher()
	extractor := extractor.NewLinkExtractor()

	ti := textinput.New()
	ti.Placeholder = "Enter your thoughts, links, todos..."
	ti.Focus()
	ti.CharLimit = 500
	ti.Width = 50
	ti.ShowSuggestions = true

	return &Model{
		config:      cfg,
		storage:     storage,
		categoriser: categoriser,
		searcher:    searcher,
		extractor:   extractor,
		textInput:   ti,
		entries:     []models.Entry{},
		currentMode: scratchpadMode,
		commands: []string{
			"Shift+Tab - Toggle between scratchpad and todo mode",
			"/todos - Interactive todo list with checkboxes",
			"/today - Show today's entries",
			"/search <query> - Search all entries", 
			"/s <query> - Search all entries",
			"/sl <query> - Search links only",
			"/help - Show this help",
			"/quit - Exit stak",
		},
		slashCommands: []string{
			"/todos",
			"/today",
			"/search ",
			"/s ",
			"/sl ",
			"/help",
			"/quit",
		},
		selectedIdx: -1,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		m.loadTodayEntries(),
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
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit

		case tea.KeyEsc:
			if m.showHelp {
				m.showHelp = false
				return m, nil
			}
			if m.currentMode == searchMode || m.currentMode == todayMode || m.currentMode == todoListMode {
				m.currentMode = scratchpadMode
				return m, m.loadTodayEntries()
			}

		case tea.KeyShiftTab:
			if m.currentMode == scratchpadMode {
				m.currentMode = todoMode
			} else if m.currentMode == todoMode {
				m.currentMode = scratchpadMode
			}
			m.selectedIdx = -1
			return m, m.loadFilteredEntries()

		case tea.KeyEnter:
			if m.showHelp {
				m.showHelp = false
				return m, nil
			}
			return m.handleEnter()

		case tea.KeyUp:
			if m.selectedIdx > 0 {
				m.selectedIdx--
			}

		case tea.KeyDown:
			if m.selectedIdx < len(m.entries)-1 {
				m.selectedIdx++
			}

		case tea.KeyTab:
			if (m.currentMode == todoMode || m.currentMode == todayMode) && m.selectedIdx >= 0 && m.selectedIdx < len(m.entries) {
				return m.toggleTodo()
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
		cmds = append(cmds, m.loadFilteredEntries())

	case todoToggledMsg:
		// Save the toggled todo entry
		if err := m.storage.SaveEntry(msg.entry); err == nil {
			// Refresh the current view if we're not in todo list mode
			if m.currentMode != todoListMode {
				cmds = append(cmds, m.loadFilteredEntries())
			}
		}

	case exitTodoListMsg:
		m.currentMode = scratchpadMode
		return m, m.loadTodayEntries()
	}

	// Handle todo list updates if in todo list mode
	if m.currentMode == todoListMode && m.todoList != nil {
		var todoCmd tea.Cmd
		todoModel, todoCmd := m.todoList.Update(msg)
		if todoList, ok := todoModel.(*TodoListModel); ok {
			m.todoList = todoList
		}
		cmds = append(cmds, todoCmd)
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	cmds = append(cmds, cmd)

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
	args := parts[1:]

	switch command {
	case "/quit", "/q":
		return m, tea.Quit

	case "/help", "/h":
		m.showHelp = !m.showHelp
		m.textInput.SetValue("")
		return m, nil

	case "/today", "/t":
		m.currentMode = todayMode
		m.textInput.SetValue("")
		return m, m.loadFilteredEntries()

	case "/todos":
		m.currentMode = todoListMode
		m.textInput.SetValue("")
		// Load today's entries and create todo list
		entries, _ := m.storage.LoadTodayEntries()
		m.todoList = NewTodoListModel(entries)
		return m, nil

	case "/search", "/s":
		if len(args) > 0 {
			query := strings.Join(args, " ")
			m.currentMode = searchMode
			m.searchQuery = query
			m.textInput.SetValue("")
			return m, m.searchEntries(query, false)
		}

	case "/sl":
		if len(args) > 0 {
			query := strings.Join(args, " ")
			m.currentMode = searchMode
			m.searchQuery = query
			m.textInput.SetValue("")
			return m, m.searchEntries(query, true)
		}
	}

	m.textInput.SetValue("")
	return m, nil
}

func (m Model) addEntry(content string) (tea.Model, tea.Cmd) {
	entry := models.NewEntry(content)
	
	// If we're in todo mode, force everything to be a todo
	if m.currentMode == todoMode {
		entry.Type = models.TypeTodo
		entry.TodoStatus = models.TodoPending
		entry.Tags = []string{"todo", "task"}
	} else {
		// Use normal categorization for other modes
		m.categoriser.CategoriseEntry(entry)
	}

	if entry.Type == models.TypeLink && entry.URL != "" {
		go func() {
			if title, err := m.extractor.GetURLTitle(entry.URL); err == nil {
				entry.URLTitle = title
				m.storage.SaveEntry(entry)
			}
		}()
	}

	if err := m.storage.SaveEntry(entry); err != nil {
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
	entry := models.NewEntry(content)
	m.categoriser.CategoriseEntry(entry)

	if err := m.storage.SaveEntryForTomorrow(entry); err != nil {
		return m, nil
	}

	m.textInput.SetValue("")
	return m, func() tea.Msg {
		return entryAddedMsg{}
	}
}


