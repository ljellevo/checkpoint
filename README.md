# checkpoint

A fast, lightweight CLI tool for creating and restoring working-tree snapshots in any Git repository — without committing.

Checkpoints capture all tracked file changes and untracked files at a point in time, stored as diffs inside `.git/checkpoints/`. You can freely restore, branch off, and experiment without touching your Git history.

---

## Usage

### Create a checkpoint

```bash
checkpoint
checkpoint -m "before refactoring auth"
```

Creates a snapshot of the current working tree (modified tracked files + untracked files). Prints a confirmation with the checkpoint ID and timestamp.

### Restore the previous checkpoint

```bash
checkpoint restore
```

Reverts the working tree to the state of the previous checkpoint. If your working tree has changes that weren't saved in the last checkpoint, you'll be asked to confirm before they are discarded.

### List all checkpoints

```bash
checkpoint -l / --list
```

Opens an interactive TUI list (built with [Bubbletea](https://github.com/charmbracelet/bubbletea)) showing all checkpoints in the current timeline, newest first. The current checkpoint is marked `CURRENT`.

| Key | Action |
|-----|--------|
| `↑` / `↓` | Navigate |
| `q` / `Esc` | Quit |

---

## Timeline behaviour

Checkpoints form a linear timeline. If you restore to an earlier checkpoint and then create a new one, the "future" checkpoints (the ones you skipped past) are automatically discarded — just like in a version-controlled timeline.

**Example:**

```
checkpoint -m "A"   →  timeline: A
checkpoint -m "B"   →  timeline: A → B
checkpoint restore  →  back to A
checkpoint -m "C"   →  timeline: A → C   (B is gone)
```

You can restore as many times as you like without creating a new checkpoint — the timeline only changes when you *create* a new checkpoint.

---

## Installation

### Build from source

Requires [Go 1.21+](https://go.dev/dl/).

```bash
git clone <repo>
cd git-checkpoint
go build -o checkpoint .
```

Then move the binary somewhere on your `$PATH`:

```bash
mv checkpoint /usr/local/bin/checkpoint
```

### Run directly

```bash
go run . -m "my checkpoint"
go run . restore
go run . list
```

---

## Developer guide

### Project structure

```
git-checkpoint/
├── main.go                    Entry point
├── cmd/
│   ├── root.go                `checkpoint [-m msg]` command
│   ├── restore.go             `checkpoint restore` command
│   └── list.go                `checkpoint list` command
├── internal/
│   ├── checkpoint/
│   │   ├── types.go           Checkpoint and Meta data structures
│   │   ├── store.go           Read/write checkpoints in .git/checkpoints/
│   │   └── git.go             Git operations (diff, apply, reset, untracked)
│   └── ui/
│       └── list.go            Bubbletea TUI for the list command
└── README.md
```

### Storage format

Checkpoints are stored in `.git/checkpoints/`:

- `meta.json` — `{ "head": "<current-checkpoint-id>" }`
- `<id>.json` — one file per checkpoint containing the patch, untracked file contents, message, timestamp, and parent ID

The checkpoints form a singly-linked list via `parent_id`. Only checkpoints reachable from `head` are part of the active timeline.

### Dependencies

| Package | Purpose |
|---------|---------|
| `github.com/spf13/cobra` | CLI command structure |
| `github.com/charmbracelet/bubbletea` | TUI framework |
| `github.com/charmbracelet/bubbles` | List component |
| `github.com/charmbracelet/lipgloss` | Terminal styling |

### Running tests

```bash
go test ./...
```

### Linting

```bash
go vet ./...
```
