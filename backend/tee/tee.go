// Package tee provides a backend that forwards every Allow call to two
// backends simultaneously and returns the more restrictive decision.
//
// This is useful when you want to enforce two independent rate-limiting
// policies at the same time — for example, a per-user limit and a global
// service limit — and only allow a request when both backends agree.
package tee

import (
	"context"
	"fmt"

	"github.com/your-org/floodgate/backend"
)

// Options configures the Tee backend.
type Options struct {
	// Primary is the first backend to consult. Its decision is always
	// returned to the caller, regardless of what Secondary decides.
	Primary backend.Backend

	// Secondary is the second backend. It is always called (side-effects
	// such as counter increments are applied), but its rejection only
	// overrides Primary when StrictMode is true.
	Secondary backend.Backend

	// StrictMode controls how the two decisions are combined.
	// When true  – both backends must allow the request (logical AND).
	// When false – only Primary must allow (Secondary is called for its
	//              side-effects but cannot block the request on its own).
	StrictMode bool
}

func (o Options) validate() error {
	if o.Primary == nil {
		return fmt.Errorf("tee: Primary backend must not be nil")
	}
	if o.Secondary == nil {
		return fmt.Errorf("tee: Secondary backend must not be nil")
	}
	return nil
}

type tee struct {
	opts Options
}

// New returns a Backend that forwards every Allow call to both Primary and
// Secondary. See Options.StrictMode for how the two decisions are combined.
func New(opts Options) (backend.Backend, error) {
	if err := opts.validate(); err != nil {
		return nil, err
	}
	return &tee{opts: opts}, nil
}

// Allow calls both backends and combines their decisions according to
// StrictMode. Both backends are always called so that counters and other
// stateful side-effects are applied consistently.
func (t *tee) Allow(ctx context.Context, key string) (bool, error) {
	primaryOK, primaryErr := t.opts.Primary.Allow(ctx, key)
	secondaryOK, secondaryErr := t.opts.Secondary.Allow(ctx, key)

	// Surface the first non-nil error. We still applied both calls above so
	// neither backend is starved of traffic information.
	if primaryErr != nil {
		return false, fmt.Errorf("tee: primary: %w", primaryErr)
	}
	if secondaryErr != nil {
		return false, fmt.Errorf("tee: secondary: %w", secondaryErr)
	}

	if t.opts.StrictMode {
		// Both must agree to allow.
		return primaryOK && secondaryOK, nil
	}

	// Lenient mode: Primary's decision wins.
	return primaryOK, nil
}
