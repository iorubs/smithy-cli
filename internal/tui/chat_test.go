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

// TestChat_InitialView shows the header and agent name once the
// window-size handshake completes.
func TestChat_InitialView(t *testing.T) {
	m := boot(t, &fakeChatClient{})
	view := stripANSI(m.View())
	if !strings.Contains(view, "chat: demo-agent") {
		t.Errorf("view missing header; got: %q", view)
	}
}

// TestChat_HistoryLoads prepends server-side transcript turns when
// FetchHistory returns content.
func TestChat_HistoryLoads(t *testing.T) {
	client := &fakeChatClient{
		HistoryFn: func() ([]agentchat.Turn, error) {
			return []agentchat.Turn{
				{From: "you", Text: "earlier question"},
				{From: "agent", Text: "earlier reply"},
			}, nil
		},
	}
	m := boot(t, client)
	updated, _ := m.Update(chatHistoryMsg{turns: []agentchat.Turn{
		{From: "you", Text: "earlier question"},
		{From: "agent", Text: "earlier reply"},
	}})
	view := stripANSI(updated.(chatModel).View())
	if !strings.Contains(view, "earlier question") {
		t.Errorf("view missing prior user turn; got: %q", view)
	}
	if !strings.Contains(view, "earlier reply") {
		t.Errorf("view missing prior agent turn; got: %q", view)
	}
}

// TestChat_HistoryError surfaces the error as a system message but
// doesn't abort the model.
func TestChat_HistoryError(t *testing.T) {
	m := boot(t, &fakeChatClient{})
	updated, _ := m.Update(chatHistoryMsg{err: errors.New("boom")})
	view := stripANSI(updated.(chatModel).View())
	if !strings.Contains(view, "history unavailable") {
		t.Errorf("view missing history-error system note; got: %q", view)
	}
	if !strings.Contains(view, "boom") {
		t.Errorf("view missing original error text; got: %q", view)
	}
}

// TestChat_ReplyAppendsAgentTurn renders the agent reply once the
// command's chatReplyMsg arrives.
func TestChat_ReplyAppendsAgentTurn(t *testing.T) {
	m := boot(t, &fakeChatClient{})
	m.thinking = true
	m.turns = append(m.turns, chatTurn{from: "you", text: "hi"})

	updated, _ := m.Update(chatReplyMsg{text: "hello back"})
	cm := updated.(chatModel)
	if cm.thinking {
		t.Error("thinking flag still set after reply")
	}
	view := stripANSI(cm.View())
	if !strings.Contains(view, "hello back") {
		t.Errorf("view missing agent reply; got: %q", view)
	}
}

// TestChat_ReplyError renders an error system bubble and clears the
// thinking flag.
func TestChat_ReplyError(t *testing.T) {
	m := boot(t, &fakeChatClient{})
	m.thinking = true
	updated, _ := m.Update(chatReplyMsg{err: errors.New("transport down")})
	cm := updated.(chatModel)
	if cm.thinking {
		t.Error("thinking flag still set after error reply")
	}
	view := stripANSI(cm.View())
	if !strings.Contains(view, "transport down") {
		t.Errorf("view missing error text; got: %q", view)
	}
}

// TestChat_EmptyReply substitutes a placeholder when the model
// returns blank text so the user sees that the round-trip completed.
func TestChat_EmptyReply(t *testing.T) {
	m := boot(t, &fakeChatClient{})
	m.thinking = true
	updated, _ := m.Update(chatReplyMsg{text: "   "})
	view := stripANSI(updated.(chatModel).View())
	if !strings.Contains(view, "(empty reply)") {
		t.Errorf("view missing empty-reply placeholder; got: %q", view)
	}
}

// TestChat_CtrlCQuits returns tea.Quit on ctrl+c.
func TestChat_CtrlCQuits(t *testing.T) {
	m := boot(t, &fakeChatClient{})
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Fatal("expected non-nil cmd on ctrl+c")
	}
	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Errorf("expected tea.QuitMsg, got %T", msg)
	}
}
