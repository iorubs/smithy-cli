package tui

import "github.com/charmbracelet/lipgloss"

var (
	styleTitle      = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7C3AED"))
	styleRunning    = lipgloss.NewStyle().Foreground(lipgloss.Color("#10B981"))
	styleStopped    = lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280"))
	styleFatal      = lipgloss.NewStyle().Foreground(lipgloss.Color("#EF4444"))
	styleMuted      = lipgloss.NewStyle().Foreground(lipgloss.Color("#9CA3AF")).Italic(true)
	styleHeader     = lipgloss.NewStyle().Bold(true).Underline(true)
	styleKindMCP    = lipgloss.NewStyle().Foreground(lipgloss.Color("#F472B6")).Bold(true)
	styleKindAgent  = lipgloss.NewStyle().Foreground(lipgloss.Color("#34D399")).Bold(true)
	styleKindDaemon = lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280")).Bold(true)

	styleLogTime  = lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280"))
	styleLogInfo  = lipgloss.NewStyle().Foreground(lipgloss.Color("#60A5FA"))
	styleLogWarn  = lipgloss.NewStyle().Foreground(lipgloss.Color("#F59E0B"))
	styleLogError = lipgloss.NewStyle().Foreground(lipgloss.Color("#EF4444"))
	styleLogDebug = lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280"))
	styleLogKey   = lipgloss.NewStyle().Foreground(lipgloss.Color("#9CA3AF"))
)
