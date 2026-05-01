package tui

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/iorubs/smithy-cli/internal/runtime/ipc"
)

type tickMsg struct{}
type statusMsg []ipc.StatusLine
type logsMsg string
type errMsg struct{ err error }

type Model struct {
	client  *ipc.Client
	logPath string
	detach  bool

	rows      []ipc.StatusLine
	lastLogs  string
	logVP     viewport.Model
	logReady  bool
	width     int
	height    int
	err       error
	quitting  bool
	connected bool
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.cmdStatus(), m.cmdReadLogs(), cmdTick())
}

func cmdTick() tea.Cmd {
	return tea.Tick(time.Second, func(_ time.Time) tea.Msg { return tickMsg{} })
}

func (m Model) cmdStatus() tea.Cmd {
	c := m.client
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		resp, err := c.Status(ctx)
		if err != nil {
			return errMsg{err}
		}
		return statusMsg(resp.Services)
	}
}

func (m Model) cmdReadLogs() tea.Cmd {
	p := m.logPath
	return func() tea.Msg {
		b, err := os.ReadFile(p)
		if err != nil {
			return logsMsg("")
		}
		lines := strings.Split(strings.TrimRight(string(b), "\n"), "\n")
		const maxLines = 2000
		if len(lines) > maxLines {
			lines = lines[len(lines)-maxLines:]
		}
		return logsMsg(strings.Join(lines, "\n"))
	}
}

func (m *Model) resizeVP() {
	if !m.logReady || m.width == 0 || m.height == 0 {
		return
	}

	used := len(m.rows) + 6
	avail := m.height - used
	if avail < 3 {
		avail = 3
	}
	m.logVP.Width = m.width - 2
	m.logVP.Height = avail
}

func (m *Model) setLogs(raw string) {
	svcKind := make(map[string]string, len(m.rows))
	for _, r := range m.rows {
		svcKind[r.Name] = string(r.Kind)
	}
	atBottom := m.logVP.AtBottom()
	prevOffset := m.logVP.YOffset
	m.logVP.SetContent(prettifyLogs(raw, svcKind))
	if atBottom {
		m.logVP.GotoBottom()
	} else {
		m.logVP.SetYOffset(prevOffset)
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		if !m.logReady {
			m.logVP = viewport.New(m.width-2, 5)
			m.logReady = true
		}
		m.resizeVP()

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "Q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "pgup", "b":
			if m.logReady {
				m.logVP.HalfPageUp()
			}
		case "pgdown", " ":
			if m.logReady {
				m.logVP.HalfPageDown()
			}
		case "G", "end":
			if m.logReady {
				m.logVP.GotoBottom()
			}
		case "g", "home":
			if m.logReady {
				m.logVP.GotoTop()
			}
		default:
			if m.logReady {
				var vpCmd tea.Cmd
				m.logVP, vpCmd = m.logVP.Update(msg)
				cmds = append(cmds, vpCmd)
			}
		}
		return m, tea.Batch(cmds...)

	case tickMsg:
		cmds = append(cmds, m.cmdStatus(), m.cmdReadLogs(), cmdTick())

	case statusMsg:
		m.rows = []ipc.StatusLine(msg)
		m.connected = true
		m.err = nil
		m.resizeVP()
		if m.logReady && m.lastLogs != "" {
			m.setLogs(m.lastLogs)
		}

	case logsMsg:
		m.lastLogs = string(msg)
		if m.logReady && m.connected {
			m.setLogs(m.lastLogs)
		}

	case errMsg:
		m.err = msg.err
		if m.connected {
			m.quitting = true
			return m, tea.Quit
		}
	}

	if m.logReady {
		var vpCmd tea.Cmd
		m.logVP, vpCmd = m.logVP.Update(msg)
		cmds = append(cmds, vpCmd)
	}
	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if m.quitting {
		if m.detach {
			return "smithy: detached\n"
		}
		return ""
	}
	if m.width == 0 {
		return "starting…\n"
	}

	var b strings.Builder
	b.WriteString(styleTitle.Render(" ◆ smithy "))
	b.WriteString("\n\n")

	if m.err != nil && !m.connected {
		b.WriteString(styleFatal.Render("  ✗ daemon unreachable: "+m.err.Error()) + "\n")
	} else if len(m.rows) == 0 {
		b.WriteString(styleMuted.Render("  (no services)") + "\n")
	} else {
		const (
			wName  = 22
			wKind  = 7
			wState = 10
		)
		hdr := "  " + padRight("NAME", wName) + " " +
			padRight("KIND", wKind) + " " +
			padRight("STATE", wState)
		b.WriteString(styleHeader.Render(hdr) + "\n")
		for _, row := range m.rows {
			line := "  " + padRight(string(row.Name), wName) + " " +
				colorKind(string(row.Kind), wKind) + " " +
				colorState(string(row.State), wState)
			b.WriteString(line + "\n")
		}
	}

	if m.logReady {
		b.WriteString("\n")
		b.WriteString(m.logVP.View())
		b.WriteString("\n")
	}

	b.WriteString("\n")
	if m.detach {
		b.WriteString(styleMuted.Render("  b/space scroll   g/G top/bottom   Q detach"))
	} else {
		b.WriteString(styleMuted.Render("  b/space scroll   g/G top/bottom   Q quit"))
	}
	return b.String()
}

func Run(socketPath, logPath string, detach bool) error {
	m := Model{
		client:  ipc.NewClient(socketPath),
		logPath: logPath,
		detach:  detach,
	}
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
