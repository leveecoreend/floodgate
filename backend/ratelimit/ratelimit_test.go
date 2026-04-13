package ratelimit_test

import (
	"testing"
	"time"

	"github.com/yourorg/floodgate/backend/ratelimit"
)

func TestOptions_Validate_Valid(t *testing.T) {
	o := ratelimit.Options{Limit: 10, Window: time.Second}
	if err := o.Validate(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestOptions_Validate_ZeroLimit(t *testing.T) {
	o := ratelimit.Options{Limit: 0, Window: time.Second}
	if err := o.Validate(); err != ratelimit.ErrInvalidLimit {
		t.Fatalf("expected ErrInvalidLimit, got %v", err)
	}
}

func TestOptions_Validate_NegativeLimit(t *testing.T) {
	o := ratelimit.Options{Limit: -5, Window: time.Second}
	if err := o.Validate(); err != ratelimit.ErrInvalidLimit {
		t.Fatalf("expected ErrInvalidLimit, got %v", err)
	}
}

func TestOptions_Validate_ZeroWindow(t *testing.T) {
	o := ratelimit.Options{Limit: 10, Window: 0}
	if err := o.Validate(); err != ratelimit.ErrInvalidWindow {
		t.Fatalf("expected ErrInvalidWindow, got %v", err)
	}
}

func TestBucketOptions_Validate_Valid(t *testing.T) {
	o := ratelimit.BucketOptions{Capacity: 100, Rate: 10.0}
	if err := o.Validate(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestBucketOptions_Validate_ZeroCapacity(t *testing.T) {
	o := ratelimit.BucketOptions{Capacity: 0, Rate: 10.0}
	if err := o.Validate(); err != ratelimit.ErrInvalidCapacity {
		t.Fatalf("expected ErrInvalidCapacity, got %v", err)
	}
}

func TestBucketOptions_Validate_NegativeCapacity(t *testing.T) {
	o := ratelimit.BucketOptions{Capacity: -1, Rate: 10.0}
	if err := o.Validate(); err != ratelimit.ErrInvalidCapacity {
		t.Fatalf("expected ErrInvalidCapacity, got %v", err)
	}
}

func TestBucketOptions_Validate_ZeroRate(t *testing.T) {
	o := ratelimit.BucketOptions{Capacity: 100, Rate: 0}
	if err := o.Validate(); err != ratelimit.ErrInvalidRate {
		t.Fatalf("expected ErrInvalidRate, got %v", err)
	}
}

func TestBucketOptions_Validate_NegativeRate(t *testing.T) {
	o := ratelimit.BucketOptions{Capacity: 100, Rate: -3.5}
	if err := o.Validate(); err != ratelimit.ErrInvalidRate {
		t.Fatalf("expected ErrInvalidRate, got %v", err)
	}
}
