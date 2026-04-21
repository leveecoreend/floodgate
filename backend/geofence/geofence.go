// Package geofence provides a rate-limiting backend that applies different
// limits based on the geographic region derived from the request key (e.g. an
// IP address mapped to a country code).
package geofence

import (
	"context"
	"errors"

	"github.com/floodgate/floodgate/backend"
)

// LookupFunc maps a key (typically an IP address) to a region identifier such
// as a country code. It returns an empty string when the region is unknown.
type LookupFunc func(key string) string

// Options configures the geofence backend.
type Options struct {
	// Lookup maps a request key to a region identifier.
	Lookup LookupFunc

	// Regions maps region identifiers to dedicated backends. When a key maps
	// to a region that has no dedicated backend the Default backend is used.
	Regions map[string]backend.Backend

	// Default is the backend used when a key's region is unknown or has no
	// dedicated entry in Regions.
	Default backend.Backend
}

func (o Options) validate() error {
	if o.Lookup == nil {
		return errors.New("geofence: Lookup must not be nil")
	}
	if o.Default == nil {
		return errors.New("geofence: Default backend must not be nil")
	}
	for region, b := range o.Regions {
		if b == nil {
			return errors.New("geofence: backend for region " + region + " must not be nil")
		}
	}
	return nil
}

type geofence struct {
	opts Options
}

// New returns a backend.Backend that routes Allow calls to the backend
// registered for the key's region, falling back to opts.Default when no
// region-specific backend is configured.
func New(opts Options) (backend.Backend, error) {
	if err := opts.validate(); err != nil {
		return nil, err
	}
	return &geofence{opts: opts}, nil
}

func (g *geofence) Allow(ctx context.Context, key string) (bool, error) {
	region := g.opts.Lookup(key)
	if b, ok := g.opts.Regions[region]; ok {
		return b.Allow(ctx, key)
	}
	return g.opts.Default.Allow(ctx, key)
}
