package tui

import "strings"

func padRight(s string, n int) string {
	if len(s) >= n {
		return s[:n]
	}
	return s + strings.Repeat(" ", n-len(s))
}

func colorKind(k string, width int) string {
	padded := padRight(k, width)
	switch k {
	case "mcp":
		return styleKindMCP.Render(padded)
	case "agent":
		return styleKindAgent.Render(padded)
	default:
		return padded
	}
}

func colorState(s string, width int) string {
	padded := padRight(s, width)
	switch s {
	case "running":
		return styleRunning.Render(padded)
	default:
		return styleStopped.Render(padded)
	}
}
