package runtime

import (
	"context"
	"log/slog"
)

type serviceKey struct{}
type logLevelKey struct{}

type serviceInfo struct {
	name string
	kind string
}

// WithLogLevel stores the log level string in ctx for daemon re-exec children.
func WithLogLevel(ctx context.Context, level string) context.Context {
	return context.WithValue(ctx, logLevelKey{}, level)
}

// LogLevelFromCtx returns the log level stored by WithLogLevel, or "info".
func LogLevelFromCtx(ctx context.Context) string {
	if v, ok := ctx.Value(logLevelKey{}).(string); ok && v != "" {
		return v
	}
	return "info"
}

// WithServiceKind returns a ctx that carries name and kind so that
// ctxHandler prepends service= and kind= to every slog record emitted with that ctx.
func WithServiceKind(ctx context.Context, name, kind string) context.Context {
	return context.WithValue(ctx, serviceKey{}, serviceInfo{name: name, kind: kind})
}

// ctxHandler wraps an slog.Handler and injects "service" and "kind"
// attributes on every record whose context carries them via WithServiceKind.
type ctxHandler struct {
	inner slog.Handler
}

func (h ctxHandler) Enabled(ctx context.Context, l slog.Level) bool {
	return h.inner.Enabled(ctx, l)
}

func (h ctxHandler) Handle(ctx context.Context, r slog.Record) error {
	if info, ok := ctx.Value(serviceKey{}).(serviceInfo); ok && info.name != "" {
		nr := slog.NewRecord(r.Time, r.Level, r.Message, r.PC)
		nr.AddAttrs(slog.String("service", info.name))
		if info.kind != "" {
			nr.AddAttrs(slog.String("kind", info.kind))
		}
		r.Attrs(func(a slog.Attr) bool {
			nr.AddAttrs(a)
			return true
		})
		return h.inner.Handle(ctx, nr)
	}
	return h.inner.Handle(ctx, r)
}

func (h ctxHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return ctxHandler{inner: h.inner.WithAttrs(attrs)}
}

func (h ctxHandler) WithGroup(name string) slog.Handler {
	return ctxHandler{inner: h.inner.WithGroup(name)}
}

// WrapDefaultWithCtx replaces slog's default handler with ctxHandler
// wrapping it so that service= and kind= are injected from context automatically.
func WrapDefaultWithCtx() {
	slog.SetDefault(slog.New(ctxHandler{inner: slog.Default().Handler()}))
}
