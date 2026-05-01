package tui

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
)

// prettifyLogs reformats each slog JSON line as a coloured
// "HH:MM:SS  LEVEL  KIND   SERVICE  msg  key=val …" line.
// Non-JSON lines are passed through unchanged.
//
// svcKind maps service name → kind (mcp/agent/daemon). When a log's
// "kind" field is not a known service kind (e.g. kind=local from an
// indexer), the service's kind is resolved from this map instead.
func prettifyLogs(s string, svcKind map[string]string) string {
	if s == "" {
		return s
	}
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		trim := strings.TrimSpace(line)
		if len(trim) < 2 || trim[0] != '{' {
			continue
		}
		var fields map[string]any
		if err := json.Unmarshal([]byte(trim), &fields); err != nil {
			continue
		}
		lines[i] = formatLogFields(fields, svcKind)
	}
	return strings.Join(lines, "\n")
}

func formatLogFields(fields map[string]any, svcKind map[string]string) string {
	skip := map[string]bool{"time": true, "level": true, "msg": true, "service": true, "kind": true}
	var b strings.Builder

	if t, ok := fields["time"].(string); ok {
		if parsed, err := time.Parse(time.RFC3339Nano, t); err == nil {
			b.WriteString(styleLogTime.Render(parsed.Format("15:04:05")))
		} else {
			b.WriteString(styleLogTime.Render(t))
		}
		b.WriteString("  ")
	}

	if lvl, ok := fields["level"].(string); ok {
		lvlU := strings.ToUpper(lvl)
		switch lvlU {
		case "ERROR":
			b.WriteString(styleLogError.Render(padRight(lvlU, 5)))
		case "WARN":
			b.WriteString(styleLogWarn.Render(padRight(lvlU, 5)))
		case "DEBUG":
			b.WriteString(styleLogDebug.Render(padRight("DEBUG", 5)))
		default:
			b.WriteString(styleLogInfo.Render(padRight(lvlU, 5)))
		}
		b.WriteString("  ")
	}

	kind, _ := fields["kind"].(string)
	svc, _ := fields["service"].(string)
	knownKind := kind == "mcp" || kind == "agent" || kind == "daemon"
	displayKind := kind
	if !knownKind {
		if k, ok := svcKind[svc]; ok {
			displayKind = k
		} else {
			displayKind = ""
		}
		if displayKind != "" {
			delete(skip, "kind")
		}
	}

	if displayKind != "" {
		switch displayKind {
		case "mcp":
			b.WriteString(styleKindMCP.Render(padRight(strings.ToUpper(displayKind), 6)))
		case "agent":
			b.WriteString(styleKindAgent.Render(padRight(strings.ToUpper(displayKind), 6)))
		case "daemon":
			b.WriteString(styleKindDaemon.Render(padRight(strings.ToUpper(displayKind), 6)))
		}
		b.WriteString("  ")
	} else {
		b.WriteString(strings.Repeat(" ", 8))
	}

	if svc != "" {
		b.WriteString(styleMuted.Render(padRight(svc, 18)))
		b.WriteString("  ")
	}

	if msg, ok := fields["msg"].(string); ok {
		b.WriteString(msg)
	}

	keys := make([]string, 0, len(fields))
	for k := range fields {
		if skip[k] {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		b.WriteString("  ")
		b.WriteString(styleLogKey.Render(k + "="))
		b.WriteString(fmt.Sprint(fields[k]))
	}

	return b.String()
}
