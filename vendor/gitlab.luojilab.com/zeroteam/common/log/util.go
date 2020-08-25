package log

import (
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

func fileline(skip int) (name string) {
	if _, file, lineNo, ok := runtime.Caller(skip); ok {
		return filepath.Base(file) + ":" + strconv.FormatInt(int64(lineNo), 10)
	}
	return ""
}

// funcName get func name.
var funcNameMap sync.Map

func funcName(skip int) (name string) {
	if pc, _, lineNo, ok := runtime.Caller(skip); ok {
		if v, ok := funcNameMap.Load(pc); ok {
			name = v.(string)
		} else {
			name = runtime.FuncForPC(pc).Name() + ":" + strconv.FormatInt(int64(lineNo), 10)
			funcNameMap.Store(pc, name)
		}
	}
	return
}

var funcFileNameMap sync.Map

// funcname@filename:linenum
func fileNameFuncName(skip int) (name string) {
	if pc, file, lineNo, ok := runtime.Caller(skip); ok {
		if v, ok := funcFileNameMap.Load(pc); ok {
			name = v.(string)
		} else {
			funcName := runtime.FuncForPC(pc).Name()
			idx := strings.LastIndex(funcName, "/")
			if idx > 0 {
				funcName = funcName[idx+1:]
			}
			idx = strings.Index(funcName, ".") // 去除包名,及路径名，
			if idx > 0 {
				funcName = funcName[idx+1:]
			}
			name = funcName + "@" + filepath.Base(file) + ":" + strconv.FormatInt(int64(lineNo), 10)
			funcFileNameMap.Store(pc, name)
		}
	}
	return
}

// func addExtraField(ctx context.Context, lv Level, fields map[string]interface{}) {
// 	traceId := traceable.GetTraceId(ctx)
// 	if traceId != "" {
// 		fields[KeyTraceId] = traceId
// 	}
// 	src := ctx.Value(KeySource)
// 	if src != nil {
// 		fields[KeySource] = src
// 	}

// 	fields[KeyTime] = time.Now()
// 	fields[KeyLevel] = lv.String()
// 	fields[KeyLevelValue] = int(lv)
// 	fields[KeyEnvMode] = env.GetEnv().GetMode().String()
// 	fields[KeyDCID] = env.GetEnv().GetDcid()
// 	fields[KeyAppname] = env.GetEnv().GetAppName()
// 	fields[KeyHostName] = env.GetEnv().GetHostname()
// }
