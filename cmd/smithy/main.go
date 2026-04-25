package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/alecthomas/kong"
	"github.com/iorubs/smithy-cli/internal/commands"
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

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: commands.ParseLogLevel(c.LogLevel),
	})))

	kctx.FatalIfErrorf(kctx.Run(&c))
}
