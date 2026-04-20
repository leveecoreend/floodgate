// Package logger provides a logging decorator backend for floodgate.
//
// It wraps any existing backend and emits structured log messages via
// [log/slog] for rejected requests (and optionally allowed requests).
//
// # Usage
//
//	import (
//		"github.com/floodgate/floodgate/backend/logger"
//		"github.com/floodgate/floodgate/backend/memory"
//	)
//
//	mem, _ := memory.New(memory.Options{Limit: 10, Window: time.Minute})
//	logged, _ := logger.New(logger.Options{
//		Inner:      mem,
//		LogAllowed: false,
//	})
//
// # Log Levels
//
// The following log levels are used depending on the outcome of each request:
//
//   - [log/slog.LevelWarn]: request was rejected (rate limit exceeded)
//   - [log/slog.LevelError]: an error occurred while consulting the inner backend
//   - [log/slog.LevelInfo]: request was allowed (only when LogAllowed is true)
//
// # Structured Fields
//
// Each log record includes structured attributes to aid filtering and
// aggregation in log management systems:
//
//   - "key": the rate-limit key for the request (e.g. IP address or user ID)
//   - "backend": the name of the inner backend implementation
package logger
