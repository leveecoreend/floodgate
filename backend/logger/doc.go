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
// Rejections are logged at WARN level, errors at ERROR level, and
// (when LogAllowed is true) allowed requests at INFO level.
package logger
