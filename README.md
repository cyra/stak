# stak

intelligent terminal scratchpad for quick thoughts, todos, links

## install

### From source
```bash
git clone https://github.com/cyra/stak.git
cd stak
go build -o stak ./cmd/stak
./stak
```

### From go install
```bash
go install github.com/cyra/stak/cmd/stak@latest
```

## usage

```bash
stak                    # start interactive mode
stak -dir ./notes       # custom notes directory  
stak -create-config     # generate sample config
```

## modes

- **scratchpad** - capture anything quickly
- **todo** - force all entries as todos
- **interactive todos** - checkbox interface via `/todos`
- **search** - find stuff with `/search` or `/s`

## slash commands

```
/todos          interactive todo list
/today          show today's entries  
/search <query> fuzzy search everything
/s <query>      same but shorter
/sl <query>     search links only
/help           show commands
/quit           exit
```

## features

- smart categorization (todos, links, notes)
- irc-style chat ui with newest entries at bottom
- autocomplete for slash commands
- bubbletea terminal interface
- local markdown storage
- tomorrow entries support

## config

optional yaml config at `~/.stak/config.yaml`:

```yaml
data_dir: "./notes"
log_level: "info"
theme: "default" 
date_format: "2006-01-02"
auto_save: true
fuzzy_search: true
```

## architecture  

hexagonal architecture with ports/adapters pattern for clean separation of concerns and easy testing

built with go + bubbletea + lipgloss
