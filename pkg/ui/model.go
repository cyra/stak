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
	normalMode mode = iota
	searchMode
	todayMode
)

type Model struct {
	config       *config.Config
	storage      *storage.Storage
	categorizer  *categorizer.Categorizer
	searcher     *search.FuzzySearcher
	extractor    *extractor.LinkExtractor
	textInput    textinput.Model
	entries      []models.Entry
	currentMode  mode
	width        int
	height       int
	ready        bool
	commands     []string
	selectedIdx  int
	searchQuery  string
	showHelp     bool
}

func NewModel() *Model {
	cfg := config.DefaultConfig()
	storage := storage.New(cfg)
	categorizer := categorizer.New()
	searcher := search.NewFuzzySearcher()
	extractor := extractor.NewLinkExtractor()

	ti := textinput.New()
	ti.Placeholder = "Enter your thoughts, links, todos..."
	ti.Focus()
	ti.CharLimit = 500
	ti.Width = 50

	return &Model{
		config:      cfg,
		storage:     storage,
		categorizer: categorizer,
		searcher:    searcher,
		extractor:   extractor,
		textInput:   ti,
		entries:     []models.Entry{},
		currentMode: normalMode,
		commands: []string{
			"/today - Show today's entries",
			"/search <query> - Search all entries", 
			"/s <query> - Search all entries",
			"/sl <query> - Search links only",
			"/help - Show this help",
			"/quit - Exit stak",
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
		m.textInput.Width = msg.Width - 4
		m.ready = true

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit

		case tea.KeyEsc:
			if m.currentMode != normalMode {
				m.currentMode = normalMode
				m.showHelp = false
				return m, m.loadTodayEntries()
			}

		case tea.KeyEnter:
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
			if m.currentMode == todayMode && m.selectedIdx >= 0 && m.selectedIdx < len(m.entries) {
				return m.toggleTodo()
			}
		}

	case entriesLoadedMsg:
		m.entries = msg.entries
		if len(m.entries) > 0 && m.selectedIdx < 0 {
			m.selectedIdx = 0
		}

	case entryAddedMsg:
		cmds = append(cmds, m.loadTodayEntries())
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	cmds = append(cmds, cmd)

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
		return m, m.loadTodayEntries()

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
	m.categorizer.CategorizeEntry(entry)

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
	m.storage.SaveEntry(entry)

	return m, m.loadTodayEntries()
}

func (m *Model) Storage() *storage.Storage {
	return m.storage
}