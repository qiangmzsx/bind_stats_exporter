package log

import (
	"context"
	"fmt"
	"sync"
)

// Init create logger with context.
var _ ILogger = &loggerImpl{}

const defaultDepth = 1

func newLogger(conf *config) *loggerImpl {
	if conf == nil {
		*conf = defaultConfigAfterParseFlag.Clone()
	}

	h := newHandlers(conf)
	tmpLog := getLogger(*conf, h)
	if conf.asDefaultLogger {
		// 之所以要clone一份，是不想作为currentLogger的logger与 此函数的返回值共享同一个depth
		// 因为log.Info("hello") ,与 tmpLog.Info("hello") 的depth是不应该相同的
		currentLogger = tmpLog
		// currentLogger = tmpLog.clone()
		// currentLogger.SetDepth(1)
	}

	return tmpLog
}

type loggerImpl struct {
	// ctx   context.Context
	conf  config
	h     *Handlers
	depth int
}

var loggerPool sync.Pool

func init() {
	loggerPool.New = func() interface{} {
		return &loggerImpl{}
	}

}
func getLogger(conf config, h *Handlers) (l *loggerImpl) {
	obj := loggerPool.Get()
	if obj == nil {
		l = &loggerImpl{}
	} else {
		l = obj.(*loggerImpl)
	}

	l.conf = conf
	l.h = h
	// l.ctx = context.Background()
	return l
}
func freeLogger(l *loggerImpl) {
	// l.ctx = context.Background()
	l.h = nil
	loggerPool.Put(l)
}

// func (l *loggerImpl) GetContextLogger(ctx context.Context) (l2 Logger, callback func()) {
// 	newL := l.context(ctx)
// 	callback = func() {
// 		freeLogger(newL)
// 	}
// 	return newL, callback
// }
func (l *loggerImpl) CloneLogger() ILogger {
	return l.clone()

}
func (l *loggerImpl) clone() (l2 *loggerImpl) {
	l2 = getLogger(l.conf, l.h)
	// l2.ctx = l.ctx
	l2.depth = l.depth
	return l2
}

// func (l *loggerImpl) context(ctx context.Context) *loggerImpl {
// 	newL := getLogger(l.conf, l.h)
// 	newL.ctx = ctx
// 	return newL
// }
func (l *loggerImpl) getCtx() (ctx context.Context) {
	return context.Background()
	// if l.ctx == nil {
	// 	l.ctx = context.Background()
	// }
	// return l.ctx
}
func (l *loggerImpl) SetDepth(depth int) {
	l.depth = depth
}
func (l *loggerImpl) GetDepth() (depth int) {
	return l.depth
}
func (l *loggerImpl) SetLogFixedKeys(keys ...string) {
	l.h.SetLogFixedKeys(keys...)
}
func (l *loggerImpl) AppendLogFixedKeys(keys ...string) {
	old := l.h.fixedKeys
	old = append(old, keys...)
	l.h.SetLogFixedKeys(old...)
}

func (l *loggerImpl) Log(ctx context.Context, depth int, lv Level, src string, d ...Pair) {
	l.h.Log(ctx, depth, lv, src, d...)
}

// Info logs a message at the info log level.
func (l *loggerImpl) InfoDepth(ctx context.Context, depth int, args ...interface{}) {
	if len(args) == 0 {
		return
	}
	l.h.Log(ctx, l.depth+depth, InfoLevel, "", KV(KeyMsg, fmt.Sprint(args...)))
}

// Warning logs a message at the warning log level.
func (l *loggerImpl) WarningDepth(ctx context.Context, depth int, args ...interface{}) {
	if len(args) == 0 {
		return
	}
	l.h.Log(ctx, l.depth+depth, WarnLevel, "", KV(KeyMsg, fmt.Sprint(args...)))
}

// Error logs a message at the error log level.
func (l *loggerImpl) ErrorDepth(ctx context.Context, depth int, args ...interface{}) {
	if len(args) == 0 {
		return
	}
	l.h.Log(ctx, l.depth+depth, ErrorLevel, "", KV(KeyMsg, fmt.Sprint(args...)))
}

// Info logs a message at the info log level.
func (l *loggerImpl) InfolnDepth(ctx context.Context, depth int, args ...interface{}) {
	if len(args) == 0 {
		return
	}
	str := fmt.Sprintln(args...)
	if len(str) > 0 {
		str = str[0 : len(str)-1] // 去除最后的\n,对json格式的内容， 里面包含一个\n有碍观瞻
	}
	l.h.Log(ctx, l.depth+depth, InfoLevel, "", KV(KeyMsg, str))
}

// Warning logs a message at the warning log level.
func (l *loggerImpl) WarninglnDepth(ctx context.Context, depth int, args ...interface{}) {
	if len(args) == 0 {
		return
	}
	str := fmt.Sprintln(args...)
	if len(str) > 0 {
		str = str[0 : len(str)-1] // 去除最后的\n,对json格式的内容， 里面包含一个\n有碍观瞻
	}
	l.h.Log(ctx, l.depth+depth, WarnLevel, "", KV(KeyMsg, str))
}

// Error logs a message at the error log level.
func (l *loggerImpl) ErrorlnDepth(ctx context.Context, depth int, args ...interface{}) {
	if len(args) == 0 {
		return
	}
	str := fmt.Sprintln(args...)
	if len(str) > 0 {
		str = str[0 : len(str)-1] // 去除最后的\n,对json格式的内容， 里面包含一个\n有碍观瞻
	}
	l.h.Log(ctx, l.depth+depth, ErrorLevel, "", KV(KeyMsg, str))
}

func (l *loggerImpl) InfofDepth(ctx context.Context, depth int, format string, args ...interface{}) {
	l.h.Log(ctx, l.depth+depth, InfoLevel, "", KV(KeyMsg, fmt.Sprintf(format, args...)))
}

func (l *loggerImpl) WarningfDepth(ctx context.Context, depth int, format string, args ...interface{}) {
	l.h.Log(ctx, l.depth+depth, WarnLevel, "", KV(KeyMsg, fmt.Sprintf(format, args...)))
}

func (l *loggerImpl) ErrorfDepth(ctx context.Context, depth int, format string, args ...interface{}) {
	l.h.Log(ctx, l.depth+depth, ErrorLevel, "", KV(KeyMsg, fmt.Sprintf(format, args...)))
}

// InfoKV logs a message with some additional context. The variadic key-value pairs are treated as they are in Witl.h.
func (l *loggerImpl) InfoKVDepth(ctx context.Context, depth int, args ...interface{}) {
	if len(args) == 0 {
		return
	}
	l.h.Log(ctx, l.depth+depth, InfoLevel, "", logKV(args)...)
}

// WarningKV logs a message with some additional context. The variadic key-value pairs are treated as they are in Witl.h.
func (l *loggerImpl) WarningKVDepth(ctx context.Context, depth int, args ...interface{}) {
	if len(args) == 0 {
		return
	}
	l.h.Log(ctx, l.depth+depth, WarnLevel, "", logKV(args)...)
}

// ErrorKV logs a message with some additional context. The variadic key-value pairs are treated as they are in Witl.h.
func (l *loggerImpl) ErrorKVDepth(ctx context.Context, depth int, args ...interface{}) {
	if len(args) == 0 {
		return
	}
	l.h.Log(ctx, l.depth+depth, ErrorLevel, "", logKV(args)...)
}

// InfoPairs logs a message at the info log level.
func (l *loggerImpl) InfoPairsDepth(ctx context.Context, depth int, args ...Pair) {
	l.h.Log(ctx, l.depth+depth, InfoLevel, "", args...)
}

// WarningPairs logs a message at the warning log level.
func (l *loggerImpl) WarningPairsDepth(ctx context.Context, depth int, args ...Pair) {
	l.h.Log(ctx, l.depth+depth, WarnLevel, "", args...)
}

// ErrorPairs logs a message at the error log level.
func (l *loggerImpl) ErrorPairsDepth(ctx context.Context, depth int, args ...Pair) {
	l.h.Log(ctx, l.depth+depth, ErrorLevel, "", args...)
}

// Info logs a message at the info log level.
func (l *loggerImpl) Info(ctx context.Context, args ...interface{}) {
	if len(args) == 0 {
		return
	}
	l.h.Log(ctx, l.depth, InfoLevel, "", KV(KeyMsg, fmt.Sprint(args...)))
}

// Warning logs a message at the warning log level.
func (l *loggerImpl) Warning(ctx context.Context, args ...interface{}) {
	if len(args) == 0 {
		return
	}
	l.h.Log(ctx, l.depth, WarnLevel, "", KV(KeyMsg, fmt.Sprint(args...)))
}

// Error logs a message at the error log level.
func (l *loggerImpl) Error(ctx context.Context, args ...interface{}) {
	if len(args) == 0 {
		return
	}
	l.h.Log(ctx, l.depth, ErrorLevel, "", KV(KeyMsg, fmt.Sprint(args...)))
}

// Info logs a message at the info log level.
func (l *loggerImpl) Infoln(ctx context.Context, args ...interface{}) {
	if len(args) == 0 {
		return
	}
	str := fmt.Sprintln(args...)
	if len(str) > 0 {
		str = str[0 : len(str)-1] // 去除最后的\n,对json格式的内容， 里面包含一个\n有碍观瞻
	}
	l.h.Log(ctx, l.depth, InfoLevel, "", KV(KeyMsg, str))
}

// Warning logs a message at the warning log level.
func (l *loggerImpl) Warningln(ctx context.Context, args ...interface{}) {
	if len(args) == 0 {
		return
	}
	str := fmt.Sprintln(args...)
	if len(str) > 0 {
		str = str[0 : len(str)-1] // 去除最后的\n,对json格式的内容， 里面包含一个\n有碍观瞻
	}
	l.h.Log(ctx, l.depth, WarnLevel, "", KV(KeyMsg, str))
}

// Error logs a message at the error log level.
func (l *loggerImpl) Errorln(ctx context.Context, args ...interface{}) {
	if len(args) == 0 {
		return
	}
	str := fmt.Sprintln(args...)
	if len(str) > 0 {
		str = str[0 : len(str)-1] // 去除最后的\n,对json格式的内容， 里面包含一个\n有碍观瞻
	}
	l.h.Log(ctx, l.depth, ErrorLevel, "", KV(KeyMsg, str))
}

func (l *loggerImpl) Infof(ctx context.Context, format string, args ...interface{}) {
	l.h.Log(ctx, l.depth, InfoLevel, "", KV(KeyMsg, fmt.Sprintf(format, args...)))
}

func (l *loggerImpl) Warningf(ctx context.Context, format string, args ...interface{}) {
	l.h.Log(ctx, l.depth, WarnLevel, "", KV(KeyMsg, fmt.Sprintf(format, args...)))
}

func (l *loggerImpl) Errorf(ctx context.Context, format string, args ...interface{}) {
	l.h.Log(ctx, l.depth, ErrorLevel, "", KV(KeyMsg, fmt.Sprintf(format, args...)))
}

// InfoKV logs a message with some additional context. The variadic key-value pairs are treated as they are in Witl.h.
func (l *loggerImpl) InfoKV(ctx context.Context, args ...interface{}) {
	if len(args) == 0 {
		return
	}
	l.h.Log(ctx, l.depth, InfoLevel, "", logKV(args)...)
}

// WarningKV logs a message with some additional context. The variadic key-value pairs are treated as they are in Witl.h.
func (l *loggerImpl) WarningKV(ctx context.Context, args ...interface{}) {
	if len(args) == 0 {
		return
	}
	l.h.Log(ctx, l.depth, WarnLevel, "", logKV(args)...)
}

// ErrorKV logs a message with some additional context. The variadic key-value pairs are treated as they are in Witl.h.
func (l *loggerImpl) ErrorKV(ctx context.Context, args ...interface{}) {
	if len(args) == 0 {
		return
	}
	l.h.Log(ctx, l.depth, ErrorLevel, "", logKV(args)...)
}

// InfoPairs logs a message at the info log level.
func (l *loggerImpl) InfoPairs(ctx context.Context, args ...Pair) {
	l.h.Log(ctx, l.depth, InfoLevel, "", args...)
}

// WarningPairs logs a message at the warning log level.
func (l *loggerImpl) WarningPairs(ctx context.Context, args ...Pair) {
	l.h.Log(ctx, l.depth, WarnLevel, "", args...)
}

// ErrorPairs logs a message at the error log level.
func (l *loggerImpl) ErrorPairs(ctx context.Context, args ...Pair) {
	l.h.Log(ctx, l.depth, ErrorLevel, "", args...)
}
func (l *loggerImpl) Print(args ...interface{}) {
	l.InfoDepth(context.Background(), l.depth+1, args...)
}

// Info logs a message at the info log level.
func (l *loggerImpl) Println(args ...interface{}) {
	l.InfolnDepth(context.Background(), l.depth+1, args...)
}
func (l *loggerImpl) Printf(format string, args ...interface{}) {
	l.InfofDepth(context.Background(), l.depth+1, format, args...)
}
func (l *loggerImpl) CloseLogger() error {
	if l.h == nil {
		return nil
	}

	return l.h.CloseLogger()
}
func (l *loggerImpl) FlushLogger() {
	l.h.FlushLogger()
}
func (l *loggerImpl) SetLogToStderr(b bool) {
	l.h.SetLogToStderr(b)
}
func (l *loggerImpl) SetLogFormatJson(b bool) {
	l.h.SetLogFormatJson(b)
}

func logKV(args []interface{}) []Pair {
	if len(args)%2 != 0 {
		Warning("log: the variadic must be plural, the last one will ignored", args)
	}
	ds := make([]Pair, 0, len(args)/2)
	for i := 0; i < len(args)-1; i = i + 2 {
		if key, ok := args[i].(string); ok {
			ds = append(ds, KV(key, args[i+1]))
		} else {
			Warningf("log: key must be string, get %T, ignored", args[i])
		}
	}
	return ds
}
