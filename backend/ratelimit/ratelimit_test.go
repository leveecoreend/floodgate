package ratelimit_test

import (
	"testing"
	"time"

	"github.com/example/floodgate/backend/ratelimit"
)

func TestOptions_Validate_Valid(t *testing.T) {
	o := ratelimit.Options{Limit: 10, Window: time.Second}
	if err := o.Validate(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestOptions_Validate_ZeroLimit(t *testing.T) {
	o := ratelimit.Options{Limit: 0, Window: time.Second}
	if err := o.Validate(); err == nil {
		t.Fatal("expected error for zero limit")
	}
}

func TestOptions_Validate_NegativeLimit(t *testing.T) {
	o := ratelimit.Options{Limit: -5, Window: time.Second}
	if err := o.Validate(); err == nil {
		t.Fatal("expected error for negative limit")
	}
}

func TestOptions_Validate_ZeroWindow(t *testing.T) {
	o := ratelimit.Options{Limit: 10, Window: 0}
	if err := o.Validate(); err == nil {
		t.Fatal("expected error for zero window")
	}
}

func TestBucketOptions_Validate_Valid(t *testing.T) {
	o := ratelimit.BucketOptions{Capacity: 5, Rate: time.Second}
	if err := o.Validate(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestBucketOptions_Validate_ZeroCapacity(t *testing.T) {
	o := ratelimit.BucketOptions{Capacity: 0, Rate: time.Second}
	if err := o.Validate(); err == nil {
		t.Fatal("expected error for zero capacity")
	}
}

func TestBucketOptions_Validate_ZeroRate(t *testing.T) {
	o := ratelimit.BucketOptions{Capacity: 5, Rate: 0}
	if err := o.Validate(); err == nil {
		t.Fatal("expected error for zero rate")
	}
}

func TestBucketOptions_Validate_NegativeRate(t *testing.T) {
	o := ratelimit.BucketOptions{Capacity: 5, Rate: -time.Second}
	if err := o.Validate(); err == nil {
		t.Fatal("expected error for negative rate")
	}
}
