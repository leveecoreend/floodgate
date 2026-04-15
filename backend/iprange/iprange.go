// Package iprange provides a rate-limiting backend that applies limits
// based on CIDR-grouped IP addresses, treating all IPs within a given
// subnet as a single rate-limited entity.
package iprange

import (
	"fmt"
	"net"
	"net/http"

	"github.com/yourusername/floodgate/backend"
	"github.com/yourusername/floodgate/backend/ratelimit"
)

// Options configures the IPRange backend.
type Options struct {
	// Inner is the underlying Backend used to enforce the rate limit.
	Inner backend.Backend

	// MaskBits is the number of bits in the subnet mask for IPv4 addresses.
	// For example, 24 groups IPs into /24 subnets (e.g. 192.168.1.0/24).
	MaskBits int

	// MaskBits6 is the number of bits in the subnet mask for IPv6 addresses.
	// Defaults to 64 if zero.
	MaskBits6 int
}

func (o *Options) validate() error {
	if o.Inner == nil {
		return fmt.Errorf("iprange: Inner backend must not be nil")
	}
	if o.MaskBits < 0 || o.MaskBits > 32 {
		return fmt.Errorf("iprange: MaskBits must be between 0 and 32, got %d", o.MaskBits)
	}
	if o.MaskBits6 < 0 || o.MaskBits6 > 128 {
		return fmt.Errorf("iprange: MaskBits6 must be between 0 and 128, got %d", o.MaskBits6)
	}
	return nil
}

type ipRangeBackend struct {
	inner     backend.Backend
	maskBits  int
	maskBits6 int
}

// New creates an IPRange backend that groups requests by subnet before
// delegating to the provided inner backend.
func New(opts Options) (backend.Backend, error) {
	if opts.MaskBits6 == 0 {
		opts.MaskBits6 = 64
	}
	if err := opts.validate(); err != nil {
		return nil, err
	}
	return &ipRangeBackend{
		inner:     opts.Inner,
		maskBits:  opts.MaskBits,
		maskBits6: opts.MaskBits6,
	}, nil
}

// Allow implements backend.Backend. It normalises the key to the network
// address of the subnet containing the IP, then delegates to the inner backend.
func (b *ipRangeBackend) Allow(r *http.Request, key string) (ratelimit.Result, error) {
	subnetKey := b.subnetKey(key)
	return b.inner.Allow(r, subnetKey)
}

// subnetKey masks the IP address to its network address string.
// If key is not a valid IP it is returned unchanged.
func (b *ipRangeBackend) subnetKey(key string) string {
	ip := net.ParseIP(key)
	if ip == nil {
		return key
	}
	var mask net.IPMask
	if ip.To4() != nil {
		mask = net.CIDRMask(b.maskBits, 32)
	} else {
		mask = net.CIDRMask(b.maskBits6, 128)
	}
	network := ip.Mask(mask)
	return network.String()
}
