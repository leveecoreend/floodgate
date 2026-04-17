// Package logger provides a logging middleware backend for floodgate.
// It wraps an existing backend and logs allow/reject decisions.
package logger

import (
	"context"
	"log/slog"

	"github.com/floodgate/floodgate/backend"
)

// Options configures the logger backend.
type Options struct {
	// Inner is the backend to wrap.
	Inner backend.Backend
	// Logger is the slog logger to use. Defaults to slog.Default().
	Logger *slog.Logger
	// LogAllowed controls whether allowed requests are logged. Default false.
	LogAllowed bool
}

func (o *Options) validate() error {
	if o.Inner == nil {
		return backend.ErrNilInner
	}
	return nil
}

type loggerBackend struct {
	inner      backend.Backend
	log        *slog.Logger
	logAllowed bool
}

// New creates a new logging backend wrapping inner.
func New(opts Options) (backend.Backend, error) {
	if err := opts.validate(); err != nil {
		return nil, err
	}
	l := opts.Logger
	if l == nil {
		l = slog.Default()
	}
	return &loggerBackend{
		inner:      opts.Inner,
		log:        l,
		logAllowed: opts.LogAllowed,
	}, nil
}

func (b *loggerBackend) Allow(ctx context.Context, key string) (bool, error) {
	allowed, err := b.inner.Allow(ctx, key)
	if err != nil {
		b.log.ErrorContext(ctx, "floodgate: backend error", "key", key, "error", err)
		return allowed, err
	}
	if !allowed {
		b.log.WarnContext(ctx, "floodgate: request rejected", "key", key)
	} else if b.logAllowed {
		b.log.InfoContext(ctx, "floodgate: request allowed", "key", key)
	}
	return allowed, nil
}
