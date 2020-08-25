package log

import (
	"context"
	"sync"
	"time"

	"gitlab.luojilab.com/zeroteam/common/env"
	"gitlab.luojilab.com/zeroteam/common/traceable"
)

// Handler is used to handle log events, outputting them to
// stdio or sending them to remote services. See the "handlers"
// directory for implementations.
//
// It is left up to Handlers to implement thread-safety.
type Render interface {
	// variadic Pair is k-v struct represent log content
	Render(ctx context.Context, depth int, l Level, src string, wt writerType, kv ...Pair)
}

func newHandlers(cfg *config) *Handlers {
	set := make(map[string]struct{})
	for _, k := range cfg.GetFilter() {
		set[k] = struct{}{}
	}
	handler := &Handlers{
		fixedKeys:       cfg.GetFixedKeys(),
		filters:         set,
		isLogFormatJson: cfg.GetLogFormatJson(),
		tostderr:        cfg.GetLogToStderr(),
	}

	handler.w = newWriter(cfg.GetDir(), cfg.GetFilePrefix(), cfg.GetFileName(),
		cfg.GetSymlinks(), cfg.IsWriteToSameFile(), cfg.logStats, cfg.dropIfBufferFull)
	handler.jsonRender = newJsonRender(handler.w)
	handler.consoleRender = newConsoleRender(handler.w)

	return handler
}

// Handlers a bundle for hander with filter function.
type Handlers struct {
	filters         map[string]struct{}
	fixedKeys       []string
	w               *writer
	jsonRender      Render
	consoleRender   Render
	isLogFormatJson bool
	tostderr        bool
	pairPool        sync.Pool
	pairLargePool   sync.Pool
}

// Log handlers logging.
func (hs Handlers) Log(ctx context.Context, depth int, lv Level, src string, d ...Pair) {
	if ctx == nil {
		ctx = context.Background()
	}

	for i := range d {
		if _, ok := hs.filters[d[i].Key]; ok {
			d[i].Value = "***"
		}
	}
	if src == "" {
		src = fileline(3 + depth)

	}
	var pairs []Pair = hs.mallocPairs(len(hs.fixedKeys)+len(d), len(d))
	for _, key := range hs.fixedKeys {
		switch key {
		case KeySource:
			if src != "" {
				pairs = append(pairs, KV(key, src))
			}
		case KeyLevel:
			pairs = append(pairs, KV(key, lv.String()))
		case KeyLevelValue:
			pairs = append(pairs, KV(key, int(lv)))
		case KeyTime:
			pairs = append(pairs, KV(key, time.Now()))
		case KeyEnvMode:
			pairs = append(pairs, KV(key, env.GetEnv().GetEnvMode().String()))
		case KeyDCID:
			dcid := env.GetEnv().GetDcid()
			if dcid != "" {
				pairs = append(pairs, KV(key, dcid))
			}
		case KeyAppname:
			pairs = append(pairs, KV(key, env.GetEnv().GetAppName()))
		case KeyHostName:
			pairs = append(pairs, KV(key, env.GetEnv().GetHostname()))
		case KeyTraceId:
			traceIdValue := traceable.GetTraceId(ctx)
			if traceIdValue != "" {
				pairs = append(pairs, KV(key, traceIdValue))
			}
		default:
			value := ctx.Value(key)
			if value != nil {
				pairs = append(pairs, KV(key, value))
			}
		}
	}
	pairs = append(pairs, d...)
	var isLogFormatJson bool = hs.isLogFormatJson
	ctxIsLogformatJsonValue := ctx.Value(is_log_format_json)
	if ctxIsLogformatJsonValue != nil {
		v, ok := ctxIsLogformatJsonValue.(bool)
		if ok {
			isLogFormatJson = v
		}
	}

	if isLogFormatJson {
		hs.jsonRender.Render(ctx, depth, lv, src, writerTypeFile, pairs...)
	} else {
		hs.consoleRender.Render(ctx, depth, lv, src, writerTypeFile, pairs...)
	}
	if hs.tostderr { // stderr 不支持打印json格式
		hs.consoleRender.Render(ctx, depth, lv, src, writerTypeStderr, pairs...)
	}
	hs.free(len(d), pairs)
}

// Close close resource.
func (hs Handlers) CloseLogger() (err error) {
	hs.w.Close()
	return nil
}

func (hs Handlers) FlushLogger() {
	hs.w.Flush()
	return
}

func (hs *Handlers) SetLogFixedKeys(keys ...string) {
	hs.fixedKeys = keys
}

// log to stderr or not
func (hs *Handlers) SetLogToStderr(b bool) {
	hs.tostderr = b
}

func (hs *Handlers) SetLogFormatJson(b bool) {
	hs.isLogFormatJson = b
}

const large = 6

func (hs *Handlers) mallocPairs(cap int, n int) []Pair {
	var cache interface{}
	if n > large {
		cache = hs.pairLargePool.Get()
	} else {
		cache = hs.pairPool.Get()
	}

	if cache == nil {
		return make([]Pair, 0, cap)
	}

	p := cache.([]Pair)
	return p
}

func (hs *Handlers) free(n int, f []Pair) {
	f = f[0:0]
	if n > large {
		hs.pairLargePool.Put(f)
	} else {
		hs.pairPool.Put(f)
	}
}
