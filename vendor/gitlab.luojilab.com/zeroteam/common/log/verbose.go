package log

import (
	"context"
	"path/filepath"
	"runtime"
	"strings"
)

// V reports whether verbosity at the call site is at least the requested level.
// The returned value is a boolean of type Verbose, which implements Info, InfoPairs etc.
// These methods will write to the Info log if called.
// Thus, one may write either
//	if log.V(2) { log.Info("log this") }
// or
//	log.V(2).Info("log this")
// The second form is shorter but the first is cheaper if logging is off because it does
// not evaluate its arguments.
//
// Whether an individual call to V generates a log record depends on the setting of
// the Config.VLevel and Config.Module flags; both are off by default. If the level in the call to
// V is at least the value of Config.VLevel, or of Config.Module for the source file containing the
// call, the V call will log.
// v must be more than 0.
func V(v int) Verbose {
	return Verbose(checkV(v, 1+verboseDepth))
}

const verboseDepth = 1

// Info logs a message at the info log level.
func (v Verbose) Info(args ...interface{}) {
	if v {
		currentLogger.InfoDepth(context.Background(), verboseDepth, args...)
	}
}

// Infoln logs a message at the info log level.
func (v Verbose) Infoln(args ...interface{}) {
	if v {
		currentLogger.InfolnDepth(context.Background(), verboseDepth, args...)
	}
}

// Infof logs a message at the info log level.
func (v Verbose) Infof(format string, args ...interface{}) {
	if v {
		currentLogger.InfofDepth(context.Background(), verboseDepth, format, args...)
	}
}

// InfoPairs logs a message at the info log level.
func (v Verbose) InfoPairs(args ...Pair) {
	if v {
		currentLogger.InfoPairsDepth(context.Background(), verboseDepth, args...)
	}
}

// InfoKV logs a message with some additional context. The variadic key-value pairs are treated as they are in With.
func (v Verbose) InfoKV(args ...interface{}) {
	if v {
		currentLogger.InfoKVDepth(context.Background(), verboseDepth, args...)
	}
}

// Info logs a message at the info log level.
func (v Verbose) ContextInfo(ctx context.Context, args ...interface{}) {
	if v {
		currentLogger.InfoDepth(ctx, verboseDepth, args...)
	}
}

// Info logs a message at the info log level.
func (v Verbose) ContextInfof(ctx context.Context, format string, args ...interface{}) {
	if v {
		currentLogger.InfofDepth(ctx, verboseDepth, format, args...)
	}
}

// InfoPairs logs a message at the info log level.
func (v Verbose) ContextInfoPairs(ctx context.Context, args ...Pair) {
	if v {
		currentLogger.InfoPairsDepth(ctx, verboseDepth, args...)
	}
}

// InfoKV logs a message with some additional context. The variadic key-value pairs are treated as they are in With.
func (v Verbose) ContexInfoKV(ctx context.Context, args ...interface{}) {
	if v {
		currentLogger.InfoKVDepth(ctx, verboseDepth, args...)
	}
}
func (v Verbose) V() bool {
	return bool(v)
}

func VContext(ctx context.Context, logger ILoggerContext, v int, depth int) (vb VerboseContext) {
	if logger == nil {
		logger = currentLogger
	}
	vb.logger = logger
	vb.ctx = ctx
	vb.depth = depth
	vb.b = checkV(v, 1+verboseDepth+depth)

	return
}
func (v VerboseContext) V() bool {
	return v.b
}

// Info logs a message at the info log level.
func (v VerboseContext) Info(args ...interface{}) {
	if v.b {
		v.logger.InfoDepth(v.ctx, v.depth, args...)
	}
}

// Info logs a message at the info log level.
func (v VerboseContext) Infof(format string, args ...interface{}) {
	if v.b {
		v.logger.InfofDepth(v.ctx, v.depth, format, args...)
	}
}

// InfoPairs logs a message at the info log level.
func (v VerboseContext) InfoPairs(args ...Pair) {
	if v.b {
		v.logger.InfoPairsDepth(v.ctx, v.depth, args...)
	}
}

// InfoKV logs a message with some additional context. The variadic key-value pairs are treated as they are in With.
func (v VerboseContext) InfoKV(args ...interface{}) {
	if v.b {
		v.logger.InfoKVDepth(v.ctx, v.depth, args...)
	}
}

func checkV(v int, depth int) bool {
	var (
		file string
	)
	if v < 0 {
		return false
	} else if globalV >= v {
		return true
	}
	if _, f, _, ok := runtime.Caller(depth); ok {
		file = f
	}
	// if strings.HasSuffix(file, ".go") {
	// 	file = file[:len(file)-3]
	// }
	if slash := strings.LastIndex(file, "/"); slash >= 0 {
		file = file[slash+1:]
	}
	for filter, lvl := range globalModule {
		var match bool
		if match = filter == file; !match {
			match, _ = filepath.Match(filter, file)
		}
		if match {
			return lvl >= v
		}
	}

	return false

}
