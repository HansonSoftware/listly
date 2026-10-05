# Listly

Manage your daily tasks right in the terminal.

This easy to use TUI allows you to efficiently organize your TODO list with vim-like keybinds.

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
- **Daily session**: Press `d` on the welcome screen to open today's note (named e.g. "Oct 05 2026 TODO"). Same day reopens it; a new day creates a new one. It's a normal session and auto-saves.
- **Auto-save**: Changes saved automatically after each action
- **Manual save**: Press `ctrl+s` to save with a custom name
- **Undo**: Press `u` to undo last action (create, move, delete) - up to 50 actions
- **Vim-style navigation**: `h/j/k/l`, `tab/shift+tab` for column switching

## Building

```bash
go build
./listly
```

Requires Go 1.27+.
