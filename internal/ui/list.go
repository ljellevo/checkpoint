package ui

import (
	"fmt"
	"io"

	"git-checkpoint/internal/checkpoint"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2).Bold(true).Foreground(lipgloss.Color("205"))
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170")).Bold(true)
	dimStyle          = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	headBadge         = lipgloss.NewStyle().
				Background(lipgloss.Color("205")).
				Foreground(lipgloss.Color("0")).
				Padding(0, 1).
				Render("HEAD")
)

type item struct {
	cp      *checkpoint.Checkpoint
	isHead  bool
}

func (i item) FilterValue() string { return i.cp.Message + i.cp.ID }

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	ts := dimStyle.Render(i.cp.Timestamp.Local().Format("2006-01-02 15:04:05"))
	msg := i.cp.Message
	if msg == "" {
		msg = dimStyle.Render("(no message)")
	}
	idPart := fmt.Sprintf("[%s]", i.cp.ID)

	var badge string
	if i.isHead {
		badge = " " + headBadge
	}

	line := fmt.Sprintf("%s %s  %s%s", idPart, ts, msg, badge)

	if index == m.Index() {
		fmt.Fprint(w, selectedItemStyle.Render("> "+line))
	} else {
		fmt.Fprint(w, itemStyle.Render(line))
	}
}

type model struct {
	list list.Model
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		m.list.SetHeight(msg.Height - 2)
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return "\n" + m.list.View()
}

func RunList(checkpoints []*checkpoint.Checkpoint) error {
	if len(checkpoints) == 0 {
		fmt.Println("No checkpoints yet. Run 'checkpoint' to create one.")
		return nil
	}

	items := make([]list.Item, len(checkpoints))
	for i, cp := range checkpoints {
		items[i] = item{cp: cp, isHead: i == 0}
	}

	const defaultWidth = 80
	const defaultHeight = 20

	l := list.New(items, itemDelegate{}, defaultWidth, defaultHeight)
	l.Title = "Checkpoints"
	l.Styles.Title = titleStyle
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)

	l.SetShowHelp(false)

	p := tea.NewProgram(model{list: l})
	_, err := p.Run()
	return err
}
