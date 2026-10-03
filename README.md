# Listly

Manage your daily tasks right in the terminal.

This easy to use TUI allows you to efficiently organize your TODO list with vim-like keybinds.

## TODO:

- [x] model datatypes
- [x] display tui
- [x] make tui appealing to look at (thanks charm cli)
- [x] choose and connect to a local db
- [x] store sessions in db
- [x] welcome page / session organizer
- [x] "daily session" option that doesn't need saved
- [x] save lists / session with keybind ctrl-s
- [x] saving popup (name your session)
- [x] create columns (todo / in progress / done)
- [x] create form (for todos)
- [x] move todos to other columns (when complete or in progress)
- [x] delete todos
- [x] undo list (with u keybind and a stack)
- [x] allow tab & shift-tab movement between cols
- [x] fix bug where keybinds are registered in filter or new task
- [x] ? displays keybind guide
- [x] finalize intuitive keybinds

## Keybinds

### Welcome Screen
| Key | Action |
|-----|--------|
| `↑` / `k` | Navigate up |
| `↓` / `j` | Navigate down |
| `enter` | Open selected session |
| `n` | Create new session (prompts for name) |
| `d` | Open daily session |
| `q` / `ctrl+c` | Quit |

### Normal Mode (Task Board)
| Key | Action |
|-----|--------|
| `h` / `left` / `shift+tab` | Previous column |
| `l` / `right` / `tab` | Next column |
| `enter` | Move task to next column |
| `d` | Delete selected task |
| `n` | New task (opens form) |
| `u` | Undo last action |
| `ctrl+s` | Save session (prompts for name) |
| `/` | Filter mode |
| `?` | Show help |
| `q` / `ctrl+c` | Quit |

### Filtering Mode
| Key | Action |
|-----|--------|
| `esc` / `enter` | Exit filter mode |

### Task Creation Form
| Key | Action |
|-----|--------|
| `enter` | Switch between title/description, or create task |
| `ctrl+c` | Cancel |

### Session Naming Popup
| Key | Action |
|-----|--------|
| `enter` | Save with entered name |
| `esc` / `ctrl+c` | Cancel |

## Features

- **Persistent sessions**: All sessions saved to SQLite database at `~/.local/share/listly/listly.db`
- **Daily session**: Special session that doesn't auto-save (press `d` on welcome screen)
- **Auto-save**: Changes saved automatically after each action (except daily sessions)
- **Manual save**: Press `ctrl+s` to save with a custom name
- **Undo**: Press `u` to undo last action (create, move, delete) - up to 50 actions
- **Vim-style navigation**: `h/j/k/l`, `tab/shift+tab` for column switching

## Building

```bash
go build
./listly
```

Requires Go 1.23+ and a C compiler (for SQLite).