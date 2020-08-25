package log

import "context"

type ILogger interface {
	ILoggerExt
	ILoggerContext
	StdLogger
}
type StdLogger interface {
	Print(v ...interface{})
	Printf(format string, v ...interface{})
	Println(v ...interface{})
}

type ILoggerBase interface {
	Info(args ...interface{}) //
	Warning(args ...interface{})
	Error(args ...interface{})
	Infoln(args ...interface{}) //
	Warningln(args ...interface{})
	Errorln(args ...interface{})
	Infof(format string, args ...interface{})
	Warningf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}
type ILoggerBaseContext interface {
	Info(ctx context.Context, args ...interface{}) //
	Warning(ctx context.Context, args ...interface{})
	Error(ctx context.Context, args ...interface{})

	Infoln(ctx context.Context, args ...interface{}) //
	Warningln(ctx context.Context, args ...interface{})
	Errorln(ctx context.Context, args ...interface{})

	Infof(ctx context.Context, format string, args ...interface{})
	Warningf(ctx context.Context, format string, args ...interface{})
	Errorf(ctx context.Context, format string, args ...interface{})
}

// 这里定义了对外的Logger接口，
type ILoggerExt interface {
	// Close handler
	CloseLogger() error
	FlushLogger()
	// SetDepth(depth int)    // 默认为0
	// GetDepth() (depth int) // 默认为0
	CloneLogger() ILogger

	// 指定固定输出哪些keys
	// 这些keys可以是 _level  _time 这些预先定义的key
	// 也可以是用户自定义的，当为用户自定义的key时，其对应的value值从传过来的context中取
	// 默认固定输出 []string{KeyLevel, KeyTime, KeyTraceId, KeySource}
	SetLogFixedKeys(keys ...string)
	AppendLogFixedKeys(keys ...string)

	SetLogToStderr(b bool)
	// the log format is json or not ,只对文件日志有效，控制台不支持json格式
	SetLogFormatJson(b bool)
}

type ILoggerContext interface {
	ILoggerBaseContext
	// 注意这里的context.Context可以是 "context"包，也可以是gitlab.luojilab.com/zeroteam/common/context包
	// gitlab.luojilab.com/zeroteam/common/context包中的Context继承了context.Context接口，
	// 提供了更方便的使用方式
	// 可以直接ctx.Info(args....),不需要传ctx仍然能取到trace信息
	// 如 ctx.Info("this is a log")
	// 推荐以后使用这种方式打log

	// 另外提一点zerotea/artemis/engine.Context直接继承自gitlab.luojilab.com/zeroteam/common/context.IContext
	// 故 以下使用方式也是可以的
	// func (this *HelloEndpoint) SayHello(c engine.Context) {
	//  c.Info("this is a log")
	// }

	// 另外一种使用方式是引入 gitlab.luojilab.com/zeroteam/common/log包
	// 然后使用 log.Info("hello" "world")

	// 以key:value对的形式打印日志，用法举例
	// log.InfoKV(ctx context.Context,"key1","value1","key2":"value2")
	// log.InfoKV(ctx context.Context,ctx,"key1","value1","key2":"value2")
	InfoKV(ctx context.Context, kvPairs ...interface{})
	WarningKV(ctx context.Context, kvPairs ...interface{})
	ErrorKV(ctx context.Context, kvPairs ...interface{})

	InfoPairs(ctx context.Context, args ...Pair)
	WarningPairs(ctx context.Context, args ...Pair)
	ErrorPairs(ctx context.Context, args ...Pair)

	InfoDepth(ctx context.Context, depth int, args ...interface{}) //
	WarningDepth(ctx context.Context, depth int, args ...interface{})
	ErrorDepth(ctx context.Context, depth int, args ...interface{})

	InfolnDepth(ctx context.Context, depth int, args ...interface{}) //
	WarninglnDepth(ctx context.Context, depth int, args ...interface{})
	ErrorlnDepth(ctx context.Context, depth int, args ...interface{})

	InfofDepth(ctx context.Context, depth int, format string, args ...interface{})
	WarningfDepth(ctx context.Context, depth int, format string, args ...interface{})
	ErrorfDepth(ctx context.Context, depth int, format string, args ...interface{})

	// 以key:value对的形式打印日志，用法举例
	// log.InfoKVDepth(ctx context.Context,depth int,"key1","value1","key2":"value2")
	// log.InfoKVDepth(ctx context.Context,depth int,ctx,"key1","value1","key2":"value2")
	InfoKVDepth(ctx context.Context, depth int, kvPairs ...interface{})
	WarningKVDepth(ctx context.Context, depth int, kvPairs ...interface{})
	ErrorKVDepth(ctx context.Context, depth int, kvPairs ...interface{})

	InfoPairsDepth(ctx context.Context, depth int, args ...Pair)
	WarningPairsDepth(ctx context.Context, depth int, args ...Pair)
	ErrorPairsDepth(ctx context.Context, depth int, args ...Pair)
}

// 如果你想使用Logger接口形式，请持有此返回值，因为每次都会CloneLogger()
// 若不想持有，请直接使用log.Info()而不要用GetLogger().Info()
func GetLogger() ILogger {
	return currentLogger
	// l2 := currentLogger.CloneLogger()
	// l2.SetDepth(l2.GetDepth() - 1)
	// return l2
}
func NewLogger(conf *config) ILogger {
	return newLogger(conf)
}

func Redirect() {
	// 这两个方法 重定向系统"log" 与logrus到本log日志
	// 目前已经在 redirect_log.go redirect_logrus.go 的init()
	// 里做了处理，
	// 但有可能第3方库会覆盖这里的配置， 所以提供Init()方法 以便使用者可
	// 比如gin/debug.go内有 log.SetFlags(0) 会覆盖"log"的flag
	// 再次覆盖回来
	CopyStandardLogTo(InfoLevel)
	redirectLogrus()
}

const (
	is_log_format_json = "zeroteam_log_is_log_format_json"
)

// 临时改变日志格式
func WithLogFormatJson(ctx context.Context, isFormatJson bool) context.Context {
	return context.WithValue(ctx, is_log_format_json, isFormatJson)
}
