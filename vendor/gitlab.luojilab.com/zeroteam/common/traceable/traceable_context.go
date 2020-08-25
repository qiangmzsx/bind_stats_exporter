package traceable

import (
	"bytes"
	"math/rand"
	"time"
	"unsafe"
)

const (
	ContextTracerKey = "tracer"
	ContextSpanKey   = "span"

	//http 的header都要符合 textproto.CanonicalMIMEHeaderKey(key)
	HeaderTraceId        = "X-Trace-Id"
	HeaderSpanId         = "X-Span-Id"
	HeaderSpanParentId   = "X-Parent-Id"
	HeaderFlag           = "X-Flag"
	HeaderBAGGAGE_PREFIX = "X-Baggage-"
)

// 另外注意 common/context 里提供了 GetTracer() SetTracer()  GetSpan() SetSpan()等方法，更方便使用

// 实际是对common/context中  Value的抽象， 这里避免traceable 依赖context包
// 同时系统包context.Context也可以实现了此接口
type IGetter interface {
	Value(key interface{}) interface{}
}
type ISetter interface {
	Set(key string, value interface{})
}

// return a ITracer or nil
func GetTracer(ctx IGetter) ITracer {
	if v := ctx.Value(ContextTracerKey); v != nil {
		if t, ok := v.(ITracer); ok {
			return t
		}
	}
	return nil
}
func SetTracer(ctx ISetter, t ITracer) {
	ctx.Set(ContextTracerKey, t)
}

// return a ISpan or nil
func GetSpan(ctx IGetter) ISpan {
	if v := ctx.Value(ContextSpanKey); v != nil {
		if span, ok := v.(ISpan); ok {
			return span
		}
	}
	return nil
}
func SetSpan(ctx ISetter, s ISpan) {
	ctx.Set(ContextSpanKey, s)
}
func GetTraceId(ctx IGetter) (value string) {
	if v := ctx.Value(HeaderTraceId); v != nil {
		if id, ok := v.(string); ok {
			return id
		}
	}
	span := GetSpan(ctx)
	if span != nil {
		return span.GetTraceID()
	}

	return ""
}
func GetSpanId(ctx IGetter) (value string) {
	if v := ctx.Value(HeaderSpanId); v != nil {
		if id, ok := v.(string); ok {
			return id
		}
	}
	span := GetSpan(ctx)
	if span != nil {
		return span.GetSpanID()
	}
	return ""
}
func GetSpanParentId(ctx IGetter) (value string) {
	if v := ctx.Value(HeaderSpanParentId); v != nil {
		if id, ok := v.(string); ok {
			return id
		}
	}
	span := GetSpan(ctx)
	if span != nil {
		return span.GetParentID()
	}
	return ""
}
func GetTraceString(ctx IGetter) string {
	var spanID, traceID string
	span := GetSpan(ctx)
	if span != nil {
		spanID = span.GetSpanID()
		traceID = span.GetTraceID()
	}
	buf := &bytes.Buffer{}
	buf.WriteString("spanID=")
	buf.WriteString(spanID)
	buf.Write(TraceKeySeparator)
	buf.WriteString("traceID=")
	buf.WriteString(traceID)
	return makeString(buf.Bytes())
}

func makeString(data []byte) string {
	return *(*string)(unsafe.Pointer(&data))
}

func SetSpanName(ctx IGetter, name string) (err error) {
	span := GetSpan(ctx)
	if span == nil {
		return
	}

	span.SetName(name)
	return
}

func SetSpanTag(s ISpan, k string, v string) {
	tagName := k
	for i := 0; i < 5; i++ {
		if _, err := s.GetTag(tagName); err != nil {
			s.SetTag(tagName, v)
			break
		} else {
			tagName = tagName + RandomSuffixStr(3)
		}
	}
}

func RandomSuffixStr(n int) string {
	var letterRunes = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	var r = rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[r.Intn(len(letterRunes))]
	}
	return "_" + string(b)
}

var (
	TraceKeySeparator = []byte("||")
)
