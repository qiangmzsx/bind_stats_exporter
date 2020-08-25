package log

import (
	"context"
	"io/ioutil"
	"os"

	"github.com/sirupsen/logrus"
)

func init() {
	redirectLogrus()
}

func redirectLogrus() {
	// FIXME: because of different stack depth call runtime.Caller will get error function name.
	logrus.AddHook(redirectHook{})
	if os.Getenv("LOGRUS_STDOUT") == "" {
		logrus.SetOutput(ioutil.Discard)
	}
}

type redirectHook struct{}

func (redirectHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (redirectHook) Fire(entry *logrus.Entry) error {
	lv := InfoLevel
	var logrusLv string
	var verbose int
	switch entry.Level {
	case logrus.FatalLevel, logrus.PanicLevel:
		logrusLv = entry.Level.String()
		fallthrough
	case logrus.ErrorLevel:
		lv = ErrorLevel
	case logrus.WarnLevel:
		lv = WarnLevel
	case logrus.InfoLevel:
		lv = InfoLevel
	case logrus.DebugLevel:
		// use verbose log replace of debuglevel
		verbose = 10
	}
	args := make([]Pair, 0, len(entry.Data)+1)
	args = append(args, Pair{Key: KeyMsg, Value: entry.Message})
	for k, v := range entry.Data {
		args = append(args, Pair{Key: k, Value: v})
	}
	if logrusLv != "" {
		args = append(args, Pair{Key: "logrus_lv", Value: logrusLv})
	}
	if verbose != 0 {
		V(verbose).InfoPairs(args...)
	} else {
		currentLogger.h.Log(context.Background(), 0, lv, "", args...)
	}
	return nil
}
