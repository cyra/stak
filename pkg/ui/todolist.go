package ui

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"stak/internal/models"
)

var (
	todoItemStyle = lipgloss.NewStyle().
			PaddingLeft(4)

	selectedTodoItemStyle = lipgloss.NewStyle().
				PaddingLeft(2).
				Foreground(lipgloss.Color("170"))

	todoCompletedStyle = lipgloss.NewStyle().
				Strikethrough(true).
				Foreground(lipgloss.Color("#666666"))

	todoPendingStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFA500"))
)

type todoItem struct {
	entry *models.Entry
}

func (i todoItem) FilterValue() string {
	return i.entry.Content
}

type todoItemDelegate struct{}

func (d todoItemDelegate) Height() int                             { return 1 }
func (d todoItemDelegate) Spacing() int                            { return 0 }
func (d todoItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d todoItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(todoItem)
	if !ok {
		return
	}

	entry := i.entry
	var checkbox, content string

	// Render checkbox
	if entry.TodoStatus == models.TodoCompleted {
		checkbox = "✓"
		content = todoCompletedStyle.Render(entry.Content)
	} else {
		checkbox = "☐"
		content = todoPendingStyle.Render(entry.Content)
	}

	// Style based on selection
	str := fmt.Sprintf("%s %s", checkbox, content)
	
	fn := todoItemStyle.Render
	if index == m.Index() {
		fn = func(...string) string {
			return selectedTodoItemStyle.Render("> " + str)
		}
	}

	fmt.Fprint(w, fn(str))
}

type TodoListModel struct {
	list     list.Model
	entries  []models.Entry
	quitting bool
}

func NewTodoListModel(entries []models.Entry) *TodoListModel {
	// Filter only todo entries
	var todoEntries []models.Entry
	for _, entry := range entries {
		if entry.Type == models.TypeTodo {
			todoEntries = append(todoEntries, entry)
		}
	}

	items := make([]list.Item, len(todoEntries))
	for i := range todoEntries {
		items[i] = todoItem{entry: &todoEntries[i]}
	}

	const defaultWidth = 80
	const defaultHeight = 20

	l := list.New(items, todoItemDelegate{}, defaultWidth, defaultHeight)
	l.Title = "Today's Todos"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFA500"))
	l.Styles.PaginationStyle = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	l.Styles.HelpStyle = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)

	return &TodoListModel{
		list:    l,
		entries: todoEntries,
	}
}

func (m TodoListModel) Init() tea.Cmd {
	return nil
}

func (m TodoListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "q", "esc":
			return m, func() tea.Msg {
				return exitTodoListMsg{}
			}

		case "enter", " ", "tab":
			i, ok := m.list.SelectedItem().(todoItem)
			if ok {
				// Toggle todo status
				if i.entry.TodoStatus == models.TodoCompleted {
					i.entry.TodoStatus = models.TodoPending
				} else {
					i.entry.TodoStatus = models.TodoCompleted
				}
				
				// Update the entry in our slice
				if m.list.Index() < len(m.entries) {
					m.entries[m.list.Index()] = *i.entry
				}
				
				return m, func() tea.Msg {
					return todoToggledMsg{entry: i.entry}
				}
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m TodoListModel) View() string {
	if m.quitting {
		return "Bye!\n"
	}
	
	if len(m.entries) == 0 {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666")).
			Italic(true).
			Margin(2, 0).
			Render("No todos for today. Add some todos first!")
	}
	
	return "\n" + m.list.View()
}

func (m *TodoListModel) SetSize(width, height int) {
	m.list.SetSize(width, height)
}

type todoToggledMsg struct {
	entry *models.Entry
}

type exitTodoListMsg struct{}