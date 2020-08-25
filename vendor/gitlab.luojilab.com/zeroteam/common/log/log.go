package log

import (
	"context"
	"flag"
	"fmt"
	"os"
)

var (
	// log level defined in level.go.
	// common log filed.
	KeyMsg        = "MSG"
	KeyLevelValue = "LEVELV"
	//  log level name: INFO, WARN...
	KeyLevel = "LEVEL"
	// log time.
	KeyTime = "TIME"
	// request path.
	// _title = "title"
	// log file.
	KeySource = "FILE"
	// app name.
	KeyAppname = "APP"
	// container ID.
	KeyHostName = "HOST"
	// uniq ID from trace.
	KeyTraceId = "TRACEID"
	// container environment: development,production, simulation,testing
	KeyEnvMode = "ENV"
	// container area.
	KeyDCID = "DCID"
)

var ( // this is the default value ,should not be changed
	defaultLogger    *loggerImpl // 只读变量，初始化后就不变，默认只log到stderr
	currentLogger    *loggerImpl // 当前正在使用的的logger,当调用log.Close()后 currentLogger就会变成defaultLogger
	defaultFixedKeys             = []string{KeyLevel, KeyTime, KeyTraceId, KeySource}
	defaultConfig    config      = config{
		dropIfBufferFull: true,
		logStats:         true,
		asDefaultLogger:  true,
		logToStderr:      true,
		logFormatJson:    true,
		fixedKeys:        defaultFixedKeys,
		writeToSameFile:  true,
	}
	defaultConfigAfterParseFlag config = defaultConfig
	pid                                = os.Getpid()
)

// 用法 调用flag.Parse()之后，再调用log.Init()
// 此时 才能解析到命令行的flag参数
// Init之后 再使用log.Info()等，配置才能生效
//flag.Parse()
// log.Init()
func Init() ILogger {
	fmt.Println("log.conf:", defaultConfigAfterParseFlag)
	return newLogger(&defaultConfigAfterParseFlag)
}
func init() {
	l := newLogger(&defaultConfig)
	defaultLogger = l // defaultLogger这里创建后就不应该被修改
	// defaultLogger.SetDepth(1)

	addFlag(flag.CommandLine, &defaultConfigAfterParseFlag)
}

// Info logs a message at the info log level.
func Info(args ...interface{}) {
	currentLogger.InfoDepth(context.Background(), defaultDepth, args...)
}
func Print(args ...interface{}) {
	currentLogger.InfoDepth(context.Background(), defaultDepth, args...)
}

// Warning logs a message at the warning log level.
func Warning(args ...interface{}) {
	currentLogger.WarningDepth(context.Background(), defaultDepth, args...)
}

// Error logs a message at the error log level.
func Error(args ...interface{}) {
	currentLogger.ErrorDepth(context.Background(), defaultDepth, args...)
}
func Fatal(args ...interface{}) {
	currentLogger.ErrorDepth(context.Background(), defaultDepth, args...)
	currentLogger.FlushLogger()
	os.Exit(1)
}
func Panic(args ...interface{}) {
	currentLogger.ErrorDepth(context.Background(), defaultDepth, args...)
	currentLogger.FlushLogger()
}

// Info logs a message at the info log level.
func Infoln(args ...interface{}) {
	currentLogger.InfolnDepth(context.Background(), defaultDepth, args...)
}
func Println(args ...interface{}) {
	currentLogger.InfolnDepth(context.Background(), defaultDepth, args...)
}

// Warning logs a message at the warning log level.
func Warningln(args ...interface{}) {
	currentLogger.WarninglnDepth(context.Background(), defaultDepth, args...)
}

// Error logs a message at the error log level.
func Errorln(args ...interface{}) {
	currentLogger.ErrorlnDepth(context.Background(), defaultDepth, args...)
}
func Fatalln(args ...interface{}) {
	currentLogger.InfolnDepth(context.Background(), defaultDepth, args...)
	currentLogger.FlushLogger()
	os.Exit(1)
}

func Infof(format string, args ...interface{}) {
	currentLogger.InfofDepth(context.Background(), defaultDepth, format, args...)
}
func Printf(format string, args ...interface{}) {
	currentLogger.InfofDepth(context.Background(), defaultDepth, format, args...)
}

// Warning logs a message at the warning log level.
func Warningf(format string, args ...interface{}) {
	currentLogger.WarningfDepth(context.Background(), defaultDepth, format, args...)
}

// Error logs a message at the error log level.
func Errorf(format string, args ...interface{}) {
	currentLogger.ErrorfDepth(context.Background(), defaultDepth, format, args...)
}
func Fatalf(format string, args ...interface{}) {
	currentLogger.ErrorfDepth(context.Background(), defaultDepth, format, args...)
	currentLogger.FlushLogger()
	os.Exit(1)
}

// InfoPairs logs a message at the info log level.
func InfoPairs(args ...Pair) {
	currentLogger.InfoPairsDepth(context.Background(), defaultDepth, args...)
}

// WarningPairs logs a message at the warning log level.
func WarningPairs(args ...Pair) {
	currentLogger.WarningPairsDepth(context.Background(), defaultDepth, args...)
}

// ErrorPairs logs a message at the error log level.
func ErrorPairs(args ...Pair) {
	currentLogger.ErrorPairsDepth(context.Background(), defaultDepth, args...)
}

// InfoKV logs a message with some additional context. The variadic key-value pairs are treated as they are in Witl.h.
func InfoKV(args ...interface{}) {
	currentLogger.InfoKVDepth(context.Background(), defaultDepth, args...)
}

// WarningKV logs a message with some additional context. The variadic key-value pairs are treated as they are in Witl.h.
func WarningKV(args ...interface{}) {
	currentLogger.WarningKVDepth(context.Background(), defaultDepth, args...)
}

// ErrorKV logs a message with some additional context. The variadic key-value pairs are treated as they are in Witl.h.
func ErrorKV(args ...interface{}) {
	currentLogger.ErrorKVDepth(context.Background(), defaultDepth, args...)
}

// InfoDepth logs a message at the info log level.
func InfoDepth(depth int, args ...interface{}) {
	currentLogger.InfoDepth(context.Background(), defaultDepth+depth, args...)
}

// WarningDepth logs a message at the warning log level.
func WarningDepth(depth int, args ...interface{}) {
	currentLogger.WarningDepth(context.Background(), defaultDepth+depth, args...)
}

// ErrorDepth logs a message at the error log level.
func ErrorDepth(depth int, args ...interface{}) {
	currentLogger.ErrorDepth(context.Background(), defaultDepth+depth, args...)
}

// InfoDepth logs a message at the info log level.
func InfolnDepth(depth int, args ...interface{}) {
	currentLogger.InfolnDepth(context.Background(), defaultDepth+depth, args...)
}

// WarningDepth logs a message at the warning log level.
func WarninglnDepth(depth int, args ...interface{}) {
	currentLogger.WarninglnDepth(context.Background(), defaultDepth+depth, args...)
}

// ErrorDepth logs a message at the error log level.
func ErrorlnDepth(depth int, args ...interface{}) {
	currentLogger.ErrorlnDepth(context.Background(), defaultDepth+depth, args...)
}

func InfofDepth(depth int, format string, args ...interface{}) {
	currentLogger.InfofDepth(context.Background(), defaultDepth+depth, format, args...)
}

// WarningDepth logs a message at the warning log level.
func WarningfDepth(depth int, format string, args ...interface{}) {
	currentLogger.WarningfDepth(context.Background(), defaultDepth+depth, format, args...)
}

// ErrorDepth logs a message at the error log level.
func ErrorfDepth(depth int, format string, args ...interface{}) {
	currentLogger.ErrorfDepth(context.Background(), defaultDepth+depth, format, args...)
}

// InfoPairsDepth logs a message at the info log level.
func InfoPairsDepth(depth int, args ...Pair) {
	currentLogger.InfoPairsDepth(context.Background(), defaultDepth+depth, args...)
}

// WarningPairsDepth logs a message at the warning log level.
func WarningPairsDepth(depth int, args ...Pair) {
	currentLogger.WarningPairsDepth(context.Background(), defaultDepth+depth, args...)
}

// ErrorPairsDepth logs a message at the error log level.
func ErrorPairsDepth(depth int, args ...Pair) {
	currentLogger.ErrorPairsDepth(context.Background(), defaultDepth+depth, args...)
}

// InfoKVDepth logs a message with some additional context. The variadic key-value pairs are treated as they are in Witl.h.
func InfoKVDepth(depth int, args ...interface{}) {
	currentLogger.InfoKVDepth(context.Background(), defaultDepth+depth, args...)
}

// WarningKVDepth logs a message with some additional context. The variadic key-value pairs are treated as they are in Witl.h.
func WarningKVDepth(depth int, args ...interface{}) {
	currentLogger.WarningKVDepth(context.Background(), defaultDepth+depth, args...)
}

// ErrorKVDepth logs a message with some additional context. The variadic key-value pairs are treated as they are in Witl.h.
func ErrorKVDepth(depth int, args ...interface{}) {
	currentLogger.ErrorKVDepth(context.Background(), defaultDepth+depth, args...)
}

func ContextInfo(ctx context.Context, args ...interface{}) {
	currentLogger.InfoDepth(ctx, defaultDepth, args...)
}

// Warning logs a message at the warning log level.
func ContextWarning(ctx context.Context, args ...interface{}) {
	currentLogger.WarningDepth(ctx, defaultDepth, args...)
}

// Error logs a message at the error log level.
func ContextError(ctx context.Context, args ...interface{}) {
	currentLogger.ErrorDepth(ctx, defaultDepth, args...)
}

// Info logs a message at the info log level.
func ContextInfoln(ctx context.Context, args ...interface{}) {
	currentLogger.InfolnDepth(ctx, defaultDepth, args...)
}

// Warning logs a message at the warning log level.
func ContextWarningln(ctx context.Context, args ...interface{}) {
	currentLogger.WarninglnDepth(ctx, defaultDepth, args...)
}

// Error logs a message at the error log level.
func ContextErrorln(ctx context.Context, args ...interface{}) {
	currentLogger.ErrorlnDepth(ctx, defaultDepth, args...)
}

func ContextInfof(ctx context.Context, format string, args ...interface{}) {
	currentLogger.InfofDepth(ctx, defaultDepth, format, args...)
}

// Warning logs a message at the warning log level.
func ContextWarningf(ctx context.Context, format string, args ...interface{}) {
	currentLogger.WarningfDepth(ctx, defaultDepth, format, args...)
}

// Error logs a message at the error log level.
func ContextErrorf(ctx context.Context, format string, args ...interface{}) {
	currentLogger.ErrorfDepth(ctx, defaultDepth, format, args...)
}

// InfoPairs logs a message at the info log level.
func ContextInfoPairs(ctx context.Context, args ...Pair) {
	currentLogger.InfoPairsDepth(ctx, defaultDepth, args...)
}

// WarningPairs logs a message at the warning log level.
func ContextWarningPairs(ctx context.Context, args ...Pair) {
	currentLogger.WarningPairsDepth(ctx, defaultDepth, args...)
}

// ErrorPairs logs a message at the error log level.
func ContextErrorPairs(ctx context.Context, args ...Pair) {
	currentLogger.ErrorPairsDepth(ctx, defaultDepth, args...)
}

// InfoKV logs a message with some additional context. The variadic key-value pairs are treated as they are in Witl.h.
func ContextInfoKV(ctx context.Context, args ...interface{}) {
	currentLogger.InfoKVDepth(ctx, defaultDepth, args...)
}

// WarningKV logs a message with some additional context. The variadic key-value pairs are treated as they are in Witl.h.
func ContextWarningKV(ctx context.Context, args ...interface{}) {
	currentLogger.WarningKVDepth(ctx, defaultDepth, args...)
}

// ErrorKV logs a message with some additional context. The variadic key-value pairs are treated as they are in Witl.h.
func ContextErrorKV(ctx context.Context, args ...interface{}) {
	currentLogger.ErrorKVDepth(ctx, defaultDepth, args...)
}

func ContextInfoDepth(ctx context.Context, depth int, args ...interface{}) {
	currentLogger.InfoDepth(ctx, defaultDepth+depth, args...)
}

// Warning logs a message at the warning log level.
func ContextWarningDepth(ctx context.Context, depth int, args ...interface{}) {
	currentLogger.WarningDepth(ctx, defaultDepth+depth, args...)
}

// Error logs a message at the error log level.
func ContextErrorDepth(ctx context.Context, depth int, args ...interface{}) {
	currentLogger.ErrorDepth(ctx, defaultDepth+depth, args...)
}

// Info logs a message at the info log level.
func ContextInfolnDepth(ctx context.Context, depth int, args ...interface{}) {
	currentLogger.InfolnDepth(ctx, defaultDepth+depth, args...)
}

// Warning logs a message at the warning log level.
func ContextWarninglnDepth(ctx context.Context, depth int, args ...interface{}) {
	currentLogger.WarninglnDepth(ctx, defaultDepth+depth, args...)
}

// Error logs a message at the error log level.
func ContextErrorlnDepth(ctx context.Context, depth int, args ...interface{}) {
	currentLogger.ErrorlnDepth(ctx, defaultDepth+depth, args...)
}

func ContextInfofDepth(ctx context.Context, depth int, format string, args ...interface{}) {
	currentLogger.InfofDepth(ctx, defaultDepth+depth, format, args...)
}

// Warning logs a message at the warning log level.
func ContextWarningfDepth(ctx context.Context, depth int, format string, args ...interface{}) {
	currentLogger.WarningfDepth(ctx, defaultDepth+depth, format, args...)
}

// Error logs a message at the error log level.
func ContextErrorfDepth(ctx context.Context, depth int, format string, args ...interface{}) {
	currentLogger.ErrorfDepth(ctx, defaultDepth+depth, format, args...)
}

// InfoPairs logs a message at the info log level.
func ContextInfoPairsDepth(ctx context.Context, depth int, args ...Pair) {
	currentLogger.InfoPairsDepth(ctx, defaultDepth+depth, args...)
}

// WarningPairs logs a message at the warning log level.
func ContextWarningPairsDepth(ctx context.Context, depth int, args ...Pair) {
	currentLogger.WarningPairsDepth(ctx, defaultDepth+depth, args...)
}

// ErrorPairs logs a message at the error log level.
func ContextErrorPairsDepth(ctx context.Context, depth int, args ...Pair) {
	currentLogger.ErrorPairsDepth(ctx, defaultDepth+depth, args...)
}

// InfoKV logs a message with some additional context. The variadic key-value pairs are treated as they are in Witl.h.
func ContextInfoKVDepth(ctx context.Context, depth int, args ...interface{}) {
	currentLogger.InfoKVDepth(ctx, defaultDepth+depth, args...)
}

// WarningKV logs a message with some additional context. The variadic key-value pairs are treated as they are in Witl.h.
func ContextWarningKVDepth(ctx context.Context, depth int, args ...interface{}) {
	currentLogger.WarningKVDepth(ctx, defaultDepth+depth, args...)
}

// ErrorKV logs a message with some additional context. The variadic key-value pairs are treated as they are in Witl.h.
func ContextErrorKVDepth(ctx context.Context, depth int, args ...interface{}) {
	currentLogger.ErrorKVDepth(ctx, defaultDepth+depth, args...)
}

// 这些key将会固定出现在所有log中，其对应的value，从ctx中取
// 比如ctx中有 X-Uid的key,则SetLogFixedKeys("X-Uid")之后
// 所有的ctx.Info("this is a log") 中会出现X-Uid:uid这样的key
// 即 "X-Uid:uid this is a log"
func SetLogFixedKeys(keys ...string) {
	currentLogger.SetLogFixedKeys(keys...)
}
func AppendLogFixedKeys(keys ...string) {
	currentLogger.AppendLogFixedKeys(keys...)
}

// Close close resource.
func Close() (err error) {
	err = currentLogger.CloseLogger()
	currentLogger = defaultLogger.clone()
	return
}
func Flush() {
	currentLogger.FlushLogger()
}

// log to stderr or not for the default logger
func SetLogToStderr(b bool) {
	currentLogger.SetLogToStderr(b)
}

// the format is json or not for the default logger
// 只对文件日志有效，控制台不支持json格式
func SetLogFormatJson(b bool) {
	currentLogger.SetLogFormatJson(b)
}
func SetV(value int) {
	globalV = value
}
func GetV() int {
	return globalV
}
func SetModule(m map[string]int) {
	globalModule = m
}
func GetModule() (m map[string]int) {
	m = make(map[string]int)
	for key, value := range globalModule {
		m[key] = value
	}
	return m
}
