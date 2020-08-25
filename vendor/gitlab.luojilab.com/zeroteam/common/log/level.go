package log

import "context"

// Level of severity.
type Level int

type Verbose bool

// Verbose is a boolean type that implements Info, InfoPairs (like Printf) etc.
type VerboseContext struct {
	b      bool
	ctx    context.Context
	logger ILoggerContext
	depth  int
}

// common log level.
const (
	DebugLevel Level = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
	NumLevel = 5
)

var (
	LevelThreshold = InfoLevel
)

var levelNames = [...]string{
	DebugLevel: "DEBUG",
	InfoLevel:  "INFO",
	WarnLevel:  "WARN",
	ErrorLevel: "ERROR",
	FatalLevel: "FATAL",
}

// String implementation.
func (l Level) String() string {
	return levelNames[l]
}
func (l Level) Char() byte {
	return levelNames[l][0]
}
