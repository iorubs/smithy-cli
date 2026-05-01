// Command smithy is the unified CLI for running and managing smithy
// MCP servers and agents.
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/alecthomas/kong"
	"github.com/iorubs/smithy-cli/internal/commands"
	"github.com/iorubs/smithy-cli/internal/runtime"
)

func main() {
	if len(os.Args) == 1 {
		os.Args = append(os.Args, "--help")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var c commands.CLI
	kctx := kong.Parse(&c,
		kong.Name("smithy"),
		kong.Description("Run and manage AI Agent and MCP server stacks"),
		kong.UsageOnError(),
		kong.HelpOptions{Compact: true, NoExpandSubcommands: true},
		kong.BindTo(ctx, (*context.Context)(nil)),
	)

	var level slog.Level
	_ = level.UnmarshalText([]byte(c.LogLevel))
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})))
	ctx = runtime.WithLogLevel(ctx, c.LogLevel)
	kctx.BindTo(ctx, (*context.Context)(nil))

	kctx.FatalIfErrorf(kctx.Run(&c))
}
