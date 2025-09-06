# stak 

A smart terminal scratchpad built with Go and Charm Bracelet's TUI libraries. stak is your intelligent dumping ground for thoughts, todos, links, and code snippets - with powerful categorisation and search capabilities.

## ✨ Features

### Smart Categorisation
- **Automatic detection** of content types: todos, links, code, questions, meetings, notes
- **Multi-layered tagging** system with domain and context awareness  
- **Intelligent parsing** that learns from your content patterns

### Powerful Search
- **Fuzzy search** across all entries with ranked results
- **Link-specific search** (`/sl <query>`)  
- **Tag-based filtering** and content matching
- **Recency-based sorting** for better relevance

### Beautiful Interface
- **Clean TUI** built with Bubbletea and Lipgloss
- **Syntax highlighting** for different content types
- **Smooth navigation** with keyboard shortcuts
- **Real-time visual feedback** for todos and links

### Data Management  
- **Local markdown files** organised by date
- **YAML frontmatter** for structured metadata
- **Git-friendly** storage format
- **No cloud dependencies** - your data stays local

## 🚀 Quick Start

### Installation

```bash
git clone <your-repo>
cd stak
go build -o stak ./cmd/stak
```

### Usage

Start stak:
```bash
./stak
```

## 📖 Commands

| Command | Description |
|---------|-------------|
| `/today` or `/t` | Show today's entries |
| `/search <query>` or `/s <query>` | Search all entries |  
| `/sl <query>` | Search links only |
| `/help` or `/h` | Show help |
| `/quit` or `/q` | Exit stak |

## ⌨️ Keyboard Shortcuts

- **Enter** - Add entry or execute command
- **↑/↓** - Navigate entries  
- **Tab** - Toggle todo completion (in today view)
- **Esc** - Go back / exit modes
- **Ctrl+C** - Quit

## 💡 Usage Examples

### Adding Content

Just type and press Enter:

```
Fix authentication bug in user service
```
→ *Automatically categorised as todo*

```  
https://go.dev/blog/error-handling-and-go
```
→ *Automatically detected as link with metadata extraction*

```
How do I implement middleware in Go?
```
→ *Automatically tagged as question*

### Using Commands

```
/today
```
View and manage today's entries

```
/s golang error handling  
```
Find all entries related to Go error handling

```
/sl github
```
Find all GitHub links you've saved

## 🏗️ Architecture

### Project Structure
```
stak/
├── cmd/stak/           # Main application entry point
├── internal/           # Private application code  
│   ├── config/        # Configuration management
│   └── models/        # Data models and types
├── pkg/               # Public packages
│   ├── categorizer/   # Smart content categorisation
│   ├── extractor/     # Link metadata extraction  
│   ├── search/        # Fuzzy search implementation
│   ├── storage/       # File-based persistence
│   └── ui/           # TUI interface components
└── README.md
```

### Data Storage

Entries are stored in `~/.stak/` as markdown files:

```
~/.stak/
├── 2024-01-15.md
├── 2024-01-16.md
└── ...
```

Each file contains:
- YAML frontmatter with structured metadata
- Human-readable markdown content
- Automatic backups via git

### Smart Categorisation System

**Layer 1: Content Type Detection**
- URLs → `link` 
- `[ ]` or `TODO:` → `todo`
- Code blocks → `code`
- Questions (`?`) → `question`
- Meeting keywords → `meeting`

**Layer 2: Domain Detection**  
- `work`, `personal`, `project`, `learning`

**Layer 3: Auto-tagging**
- Programming languages: `go`, `javascript`, `python`
- Context clues: `bug`, `feature`, `urgent`

## 🛠️ Development

### Dependencies
- [Bubbletea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Bubbles](https://github.com/charmbracelet/bubbles) - UI components  
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Styling
- [Glamour](https://github.com/charmbracelet/glamour) - Markdown rendering
- [Log](https://github.com/charmbracelet/log) - Structured logging

### Building from Source

```bash
git clone <repo>
cd stak
go mod download
go build -o stak ./cmd/stak
```

### Running Tests

```bash  
go test ./...
```

## 🔮 Future Enhancements

- [ ] Plugin system for custom categorisers
- [ ] Export to various formats (JSON, CSV, PDF)
- [ ] Sync across devices (optional cloud backup)
- [ ] Integration with external tools (Slack, Notion, etc.)
- [ ] Advanced analytics and insights
- [ ] Custom themes and colour schemes

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 Licence

This project is licensed under the MIT Licence - see the [LICENCE](LICENCE) file for details.

## 🙏 Acknowledgements

- [Charm Bracelet](https://charm.sh) for the amazing TUI toolkit
- The Go community for excellent tooling and libraries
- Everyone who contributes ideas and feedback

---

**stak** - *Your intelligent terminal scratchpad* 🚀
