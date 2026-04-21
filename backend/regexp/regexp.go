// Package regexp provides a backend decorator that applies rate limiting
// selectively based on whether the request key matches a regular expression.
// Keys that match the pattern are forwarded to the inner backend; keys that
// do not match are allowed through unconditionally.
package regexp

import (
	"context"
	"fmt"
	"regexp"

	"github.com/floodgate/floodgate/backend"
)

// Options configures the regexp backend.
type Options struct {
	// Inner is the backend to delegate to when the key matches Pattern.
	Inner backend.Backend

	// Pattern is the regular expression to match against the key.
	Pattern string
}

func (o Options) validate() error {
	if o.Inner == nil {
		return fmt.Errorf("regexp: Inner backend must not be nil")
	}
	if o.Pattern == "" {
		return fmt.Errorf("regexp: Pattern must not be empty")
	}
	return nil
}

type regexpBackend struct {
	inner   backend.Backend
	pattern *regexp.Regexp
}

// New returns a backend that forwards Allow calls to inner only when the key
// matches the compiled regular expression. Non-matching keys are always allowed.
func New(opts Options) (backend.Backend, error) {
	if err := opts.validate(); err != nil {
		return nil, err
	}
	re, err := regexp.Compile(opts.Pattern)
	if err != nil {
		return nil, fmt.Errorf("regexp: invalid pattern %q: %w", opts.Pattern, err)
	}
	return &regexpBackend{
		inner:   opts.Inner,
		pattern: re,
	}, nil
}

func (r *regexpBackend) Allow(ctx context.Context, key string) (bool, error) {
	if !r.pattern.MatchString(key) {
		return true, nil
	}
	return r.inner.Allow(ctx, key)
}
