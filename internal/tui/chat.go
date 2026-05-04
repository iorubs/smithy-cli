package tui

import (
	"context"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/iorubs/smithy-cli/internal/agentchat"
)

// chatTurn is one message in the visible transcript.
type chatTurn struct {
	from string // "you", "agent", "system"
	text string
}

type chatReplyMsg struct {
	text string
	err  error
}

type chatHistoryMsg struct {
	turns []agentchat.Turn
	err   error
}

// ChatClient is the small surface chatModel needs from agentchat.
// Defined here (not in agentchat) so tests can drive the TUI with a
// fake without pulling in a2a transport.
type ChatClient interface {
	SendText(ctx context.Context, text string) (string, error)
	FetchHistory(ctx context.Context) ([]agentchat.Turn, error)
}

type chatModel struct {
	client    ChatClient
	agentName string

	vp       viewport.Model
	input    textarea.Model
	turns    []chatTurn
	thinking bool
	width    int
	height   int
	ready    bool
	quitting bool
}

func newChatModel(client ChatClient, agentName string) chatModel {
	ta := textarea.New()
	ta.Placeholder = "Type a message..."
	ta.Focus()
	ta.SetHeight(3)
	ta.ShowLineNumbers = false
	ta.KeyMap.InsertNewline.SetEnabled(false)
	return chatModel{client: client, agentName: agentName, input: ta}
}

func (m chatModel) Init() tea.Cmd {
	cli := m.client
	fetch := func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		turns, err := cli.FetchHistory(ctx)
		return chatHistoryMsg{turns: turns, err: err}
	}
	return tea.Batch(textarea.Blink, fetch)
}

func (m chatModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		inputH := m.input.Height() + 4
		vpH := m.height - inputH - 4
		if vpH < 4 {
			vpH = 4
		}
		if !m.ready {
			m.vp = viewport.New(m.width-2, vpH)
			m.ready = true
		} else {
			m.vp.Width, m.vp.Height = m.width-2, vpH
		}
		m.input.SetWidth(m.width - 6)
		m.redraw()

	case tea.KeyMsg:
		if m.quitting {
			break
		}
		switch msg.Type {
		case tea.KeyCtrlC:
			m.quitting = true
			return m, tea.Quit
		case tea.KeyEnter:
			text := strings.TrimSpace(m.input.Value())
			if text == "" {
				break
			}
			m.turns = append(m.turns, chatTurn{from: "you", text: text})
			m.thinking = true
			m.input.Reset()
			m.redraw()
			cli := m.client
			cmds = append(cmds, func() tea.Msg {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
				defer cancel()
				reply, err := cli.SendText(ctx, text)
				return chatReplyMsg{text: reply, err: err}
			})
		}

	case chatReplyMsg:
		m.thinking = false
		switch {
		case msg.err != nil:
			m.turns = append(m.turns, chatTurn{from: "system", text: "error: " + msg.err.Error()})
		case strings.TrimSpace(msg.text) == "":
			m.turns = append(m.turns, chatTurn{from: "system", text: "(empty reply)"})
		default:
			m.turns = append(m.turns, chatTurn{from: "agent", text: strings.TrimSpace(msg.text)})
		}
		m.redraw()

	case chatHistoryMsg:
		if msg.err != nil {
			m.turns = append([]chatTurn{{from: "system", text: "history unavailable: " + msg.err.Error()}}, m.turns...)
		} else if len(msg.turns) > 0 {
			loaded := make([]chatTurn, 0, len(msg.turns)+len(m.turns))
			for _, t := range msg.turns {
				loaded = append(loaded, chatTurn{from: t.From, text: t.Text})
			}
			m.turns = append(loaded, m.turns...)
		}
		m.redraw()
	}

	if m.ready {
		var vpCmd tea.Cmd
		m.vp, vpCmd = m.vp.Update(msg)
		cmds = append(cmds, vpCmd)
	}
	var inputCmd tea.Cmd
	m.input, inputCmd = m.input.Update(msg)
	cmds = append(cmds, inputCmd)
	return m, tea.Batch(cmds...)
}

func (m *chatModel) redraw() {
	if !m.ready {
		return
	}
	width := m.vp.Width
	if width < 20 {
		width = 20
	}
	bubbleMax := width - 4
	if bubbleMax < 10 {
		bubbleMax = 10
	}
	youBubble := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#60A5FA")).
		Foreground(lipgloss.Color("#DBEAFE")).
		Padding(0, 1).
		Width(bubbleMax * 65 / 100)
	agentBubble := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#34D399")).
		Foreground(lipgloss.Color("#D1FAE5")).
		Padding(0, 1).
		Width(bubbleMax)

	var b strings.Builder
	for i, t := range m.turns {
		if i > 0 {
			b.WriteString("\n")
		}
		switch t.from {
		case "you":
			block := lipgloss.JoinVertical(lipgloss.Right,
				styleHeader.Render("you"),
				youBubble.Render(t.text),
			)
			b.WriteString(lipgloss.PlaceHorizontal(width, lipgloss.Right, block))
		case "agent":
			block := lipgloss.JoinVertical(lipgloss.Left,
				styleKindAgent.Render(m.agentName),
				agentBubble.Render(t.text),
			)
			b.WriteString(lipgloss.PlaceHorizontal(width, lipgloss.Left, block))
		case "system":
			b.WriteString(lipgloss.PlaceHorizontal(width, lipgloss.Center,
				styleMuted.Render("◆ "+t.text)))
		}
		b.WriteString("\n")
	}
	if m.thinking {
		b.WriteString("\n")
		b.WriteString(styleMuted.Render("◆ " + m.agentName + " is thinking..."))
		b.WriteString("\n")
	}
	m.vp.SetContent(b.String())
	m.vp.GotoBottom()
}

func (m chatModel) View() string {
	if !m.ready {
		return "starting...\n"
	}
	header := styleTitle.Render(" ◆ chat: " + m.agentName + " ")
	help := styleMuted.Render("  enter send   ctrl+c quit")
	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		"",
		m.vp.View(),
		"",
		m.input.View(),
		help,
	)
}

// RunChat opens a bubbletea chat TUI bound to client.
func RunChat(client agentchat.Client, agentName string) error {
	p := tea.NewProgram(newChatModel(client, agentName), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
