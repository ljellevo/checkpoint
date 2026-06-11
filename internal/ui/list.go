package ui

import (
	"fmt"
	"strings"

	"git-checkpoint/internal/checkpoint"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	cursorStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	headStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("0")).Background(lipgloss.Color("205"))
	afterStyle   = lipgloss.NewStyle()
	activeStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true)
	dimTextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	pipeStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("238"))
	warnStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true)
	nodeStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)

type uiState int

const (
	stateList uiState = iota
	stateConfirm
)

type Model struct {
	checkpoints []*checkpoint.Checkpoint
	headID      string
	cursor      int
	state       uiState
	selected    *checkpoint.Checkpoint
	isDirty     bool
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.state {
	case stateList:
		return m.updateList(msg)
	case stateConfirm:
		return m.updateConfirm(msg)
	}
	return m, nil
}

func (m Model) updateList(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.checkpoints)-1 {
				m.cursor++
			}
		case "enter":
			if m.isDirty {
				m.state = stateConfirm
			} else {
				m.selected = m.checkpoints[m.cursor]
				return m, tea.Quit
			}
		}
	}
	return m, nil
}

func (m Model) updateConfirm(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "y", "Y":
			m.selected = m.checkpoints[m.cursor]
			return m, tea.Quit
		case "n", "N", "esc":
			m.state = stateList
		case "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Model) View() string {
	var b strings.Builder

	pipe := pipeStyle.Render("|")
	indent := "  "

	b.WriteString(indent + dimTextStyle.Render("Most recent") + "\n")
	b.WriteString(indent + pipe + "\n")

	for i, cp := range m.checkpoints {
		//isFuture := isFutureCheckpoint(m.checkpoints, m.headID, i)
		isHead := cp.ID == m.headID

		// Node character
		var node string
		if i == m.cursor {
			node = cursorStyle.Render(">")
		} else {
			node = nodeStyle.Render("o")
		}

		// Timestamp and message
		ts := cp.Timestamp.Local().Format("2006-01-02 15:04:05")
		msg := cp.Message
		if msg == "" {
			msg = "(no message)"
		}

		// Badge
		var badge string
		if isHead {
			badge = "  " + headStyle.Render("CURRENT")
		}

		line := fmt.Sprintf("[%s] %s  %s%s", cp.ID, ts, msg, badge)
		if i == m.cursor {
			line = activeStyle.Render(line)
		}

		b.WriteString(node + " " + line + "\n")

		if i < len(m.checkpoints)-1 {
			b.WriteString(indent + pipe + "\n")
		}
	}

	b.WriteString(indent + pipe + "\n")
	b.WriteString(indent + dimTextStyle.Render("Oldest") + "\n")

	if m.state == stateConfirm {
		cp := m.checkpoints[m.cursor]
		b.WriteString("\n" + warnStyle.Render("Changes not saved.") +
			fmt.Sprintf(" Restore to [%s]? [y/N] ", cp.ID))
	} else {
		b.WriteString(dimTextStyle.Render("\n↑/↓ navigate • enter restore • q quit"))
	}

	return b.String()
}

func isFutureCheckpoint(checkpoints []*checkpoint.Checkpoint, headID string, index int) bool {
	for i, cp := range checkpoints {
		if cp.ID == headID {
			return index < i
		}
	}
	return false
}

// RunList shows the interactive checkpoint list and returns the selected
// checkpoint, or nil if the user quit without selecting.
func RunList(checkpoints []*checkpoint.Checkpoint, headID string, isDirty bool) (*checkpoint.Checkpoint, error) {
	if len(checkpoints) == 0 {
		fmt.Println("No checkpoints yet. Run 'checkpoint' to create one.")
		return nil, nil
	}

	startCursor := 0
	for i, cp := range checkpoints {
		if cp.ID == headID {
			startCursor = i
			break
		}
	}

	m := Model{
		checkpoints: checkpoints,
		headID:      headID,
		cursor:      startCursor,
		isDirty:     isDirty,
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	result, err := p.Run()
	if err != nil {
		return nil, err
	}
	if final, ok := result.(Model); ok {
		return final.selected, nil
	}
	return nil, nil
}
