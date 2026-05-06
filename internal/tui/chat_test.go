package tui

import (
	"context"
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/iorubs/smithy-cli/internal/agentchat"
)

// fakeChatClient drives chatModel without a live a2a transport.
// Tests fill SendFn / HistoryFn to control replies.
type fakeChatClient struct {
	HistoryFn func() ([]agentchat.Turn, error)
	SendFn    func(text string) (string, error)
}

func (f *fakeChatClient) FetchHistory(ctx context.Context) ([]agentchat.Turn, error) {
	if f.HistoryFn == nil {
		return nil, nil
	}
	return f.HistoryFn()
}

func (f *fakeChatClient) SendText(ctx context.Context, text string) (string, error) {
	if f.SendFn == nil {
		return "", nil
	}
	return f.SendFn(text)
}

// boot puts the model into the ready state with a known viewport size
// so View() returns rendered content rather than the "starting..."
// placeholder.
func boot(t *testing.T, client ChatClient) chatModel {
	t.Helper()
	m := newChatModel(client, "demo-agent")
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	return updated.(chatModel)
}

// stripANSI removes lipgloss colour codes so assertions can match
// plain text without depending on the terminal profile.
func stripANSI(s string) string {
	var b strings.Builder
	skipping := false
	for _, r := range s {
		if r == 0x1b {
			skipping = true
			continue
		}
		if skipping {
			if r == 'm' {
				skipping = false
			}
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

func TestChat(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(m *chatModel)
		msg     tea.Msg
		want    []string
		notWant []string
		check   func(t *testing.T, m chatModel, cmd tea.Cmd)
	}{
		{
			name: "initial view shows header",
			want: []string{"chat: demo-agent"},
		},
		{
			name: "history loads prior turns",
			msg: chatHistoryMsg{turns: []agentchat.Turn{
				{From: "you", Text: "earlier question"},
				{From: "agent", Text: "earlier reply"},
			}},
			want: []string{"earlier question", "earlier reply"},
		},
		{
			name: "history error surfaces system note",
			msg:  chatHistoryMsg{err: errors.New("boom")},
			want: []string{"history unavailable", "boom"},
		},
		{
			name: "reply appends agent turn and clears thinking",
			setup: func(m *chatModel) {
				m.thinking = true
				m.turns = append(m.turns, chatTurn{from: "you", text: "hi"})
			},
			msg:  chatReplyMsg{text: "hello back"},
			want: []string{"hello back"},
			check: func(t *testing.T, m chatModel, _ tea.Cmd) {
				if m.thinking {
					t.Error("thinking flag still set after reply")
				}
			},
		},
		{
			name: "reply error clears thinking and shows error",
			setup: func(m *chatModel) {
				m.thinking = true
			},
			msg:  chatReplyMsg{err: errors.New("transport down")},
			want: []string{"transport down"},
			check: func(t *testing.T, m chatModel, _ tea.Cmd) {
				if m.thinking {
					t.Error("thinking flag still set after error reply")
				}
			},
		},
		{
			name: "empty reply shows placeholder",
			setup: func(m *chatModel) {
				m.thinking = true
			},
			msg:  chatReplyMsg{text: "   "},
			want: []string{"(empty reply)"},
		},
		{
			name: "ctrl+c returns quit command",
			msg:  tea.KeyMsg{Type: tea.KeyCtrlC},
			check: func(t *testing.T, _ chatModel, cmd tea.Cmd) {
				if cmd == nil {
					t.Fatal("expected non-nil cmd on ctrl+c")
				}
				if _, ok := cmd().(tea.QuitMsg); !ok {
					t.Errorf("expected tea.QuitMsg, got %T", cmd())
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := boot(t, &fakeChatClient{})
			if tc.setup != nil {
				tc.setup(&m)
			}
			updated, cmd := m.Update(tc.msg)
			cm := updated.(chatModel)
			view := stripANSI(cm.View())
			for _, want := range tc.want {
				if !strings.Contains(view, want) {
					t.Errorf("view missing %q; got: %q", want, view)
				}
			}
			for _, notWant := range tc.notWant {
				if strings.Contains(view, notWant) {
					t.Errorf("view unexpectedly contains %q; got: %q", notWant, view)
				}
			}
			if tc.check != nil {
				tc.check(t, cm, cmd)
			}
		})
	}
}
