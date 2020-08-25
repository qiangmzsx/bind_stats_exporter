package log

import (
	"bytes"
	"context"
	stdLog "log"
)

// CopyStandardLogTo arranges for messages written to the Go "log" package's
// default logs to also appear in the Google logs for the named and lower
// severities.  Subsequent changes to the standard log's default output location
// or format may break this behavior.
//
func init() {
	CopyStandardLogTo(InfoLevel)
}
func CopyStandardLogTo(lv Level) {
	// Set a log format that captures the user's file and line:
	//   d.go:23: message
	stdLog.SetFlags(stdLog.Lshortfile)
	stdLog.SetOutput(logBridge(lv))
}

// logBridge provides the Write method that enables CopyStandardLogTo to connect
// Go's standard logs to the logs provided by this package.
type logBridge Level

// Write parses the standard logging line and passes its components to the
// logger for severity(lb).

func (lb logBridge) Write(b []byte) (n int, err error) {
	var (
		src  = ""   // file.go:23
		text string // mmessage
	)
	// Split "d.go:23: message" into "d.go:23",  and "message".
	if parts := bytes.SplitN(b, []byte{':', ' '}, 2); len(parts) != 2 {
		text = string(b)
	} else {
		src = string(parts[0])
		text = string(parts[1])
	}
	if len(text) > 0 && text[len(text)-1] == '\n' {
		text = text[0 : len(text)-1]
	}

	// printWithFileLine with alsoToStderr=true, so standard log messages
	// always appear on standard error.
	currentLogger.Log(context.Background(), 0, Level(lb), src, KV(KeyMsg, text))
	return len(b), nil
}
