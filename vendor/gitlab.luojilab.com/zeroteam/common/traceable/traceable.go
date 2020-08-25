package traceable

import (
	"gitlab.luojilab.com/zeroteam/common/traceable/model"
)

//
// model.Span.Flag 是一个int64,按二进制位的含义如下：model/span_flag.go
//
//								 FlAG DESCRIPTION
//|---------|---------|---------| --------|----sample_rate----|span_kind|--sample-|
//|0000 0000|0000 0000|0000 0000|0000 0000|0000 0000|0000 0000|0000 0000|0000 0000|
//                                                                               ↑
//（下面提及bit的pos时，说的是从0开始计数的，比如上一行箭头所指位置为第0位）               pos第0位
// 64个比特位的具体含义
// 第一个字节的8位目前只有低3位用到了，如下：
// 第0位: debug标志 （不知具体什么用）
// 第1位: 如果此位置为0 则会根据采样率判断对此span是否采样，若采样则会将 第2位 置为1,若否则将 第2位 置为0,
//       若此值已经为1,说明已经根据采样率计算过此span是否命中采样,直接根据第2位的值判断此span是否命中采样
// 第2位：是否命中采样，此值根据第1位 与采样率来决定。

// 第2字节的8位(即第8~15位)主要记录span的kind信息，目前有以下几种SpanKind
// SPAN_KIND_API,SPAN_KIND_ERROR,SPAN_KIND_PANIC
// SPAN_KIND_API表示这是一条api，此种类型的span在我的服务被其他服务访问时生成， 可以用于统计我这条api的qps
// SPAN_KIND_ERROR,SPAN_KIND_PANIC则分别表示这是一条异常的span,可能是超时,panic等原因，可以用于报警。
// 最新设计的时候，span的kind只能有一种，即如果设置成SPAN_KIND_API，就不能同时把它标记为SPAN_KIND_ERROR
// 这样设计显然是不合理的。
// 老版本的4个kind定义如下
// span_kind_normal  spankind = iota << 8 // =0 此值0值无意义
// span_kind_api                          // =256	=2#00000001 00000000 等同于 span_kind_v1_pos8=1 span_kind_v1_pos9=0
// span_kind_metrics                      // =512	=2#00000010 00000000 等同于 span_kind_v1_pos8=0 span_kind_v1_pos9=1
// span_kind_error                        // =768	=2#00000011 00000000 等同于 span_kind_v1_pos8=1 span_kind_v1_pos9=1
//
//新版kind 第8位：与 第9位： 已废弃，若一旦发现某个span第8 或第9位有值， 则说明是老版的span，依然会按老版的规则解析。
//
// 第10位： SPAN_KIND_API 此种类型的span在我的服务被其他服务访问时生成 ， 可以用于统计我这条api的qps
// 第11位：  SPAN_KIND_ERROR 此span为异常span
// 第12位：  SPAN_KIND_PANIC 此span为异常panic类型异常
// 第13~15位 留空

// 第16~31位存的是一个int16的数值，表示采样率

type SpanFlag = model.SpanFlag
type SpanKind = model.SpanKind // SpanKind占Flag的第8~15位（以0为基）,即占第2字节的8位

// const (
// 	DEBUG       SpanFlag = model.DEBUG
// 	SAMPLED_SET SpanFlag = model.SAMPLED_SET
// 	SAMPLED     SpanFlag = model.SAMPLED
// )

const (
	DEBUG_KIND       SpanKind = model.DEBUG_KIND       // flag第0位表示debug
	SAMPLED_SET_KIND SpanKind = model.SAMPLED_SET_KIND // flag第1位表示 SAMPLED_SET
	SAMPLED_KIND     SpanKind = model.SAMPLED_KIND     //flag第2位表示 SAMPLED

	SPAN_KIND_API   SpanKind = model.SPAN_KIND_API
	SPAN_KIND_ERROR SpanKind = model.SPAN_KIND_ERROR
	SPAN_KIND_PANIC SpanKind = model.SPAN_KIND_PANIC

	// SPAN_KIND_MAX SpanKind = model.SPAN_KIND_MAX // 2#10000000 00000000
)

type ISpanContext interface {
	GetTraceIDUint64() uint64
	GetTraceID() string
	SetTraceID(i uint64)

	GetSpanID() string
	GetSpanIDUint64() uint64
	SetSpanID(uint64)

	GetParentID() string // get span parentid
	GetParentIDUint64() uint64
	SetParentID(i uint64)

	GetFlag() string
	SetSpanFlag(SpanFlag)
	GetSpanFlag() SpanFlag
	AddBaggageItem(key, value string)
	GetBaggageItem(key string) string
	GetFullBaggage() map[string]string
	Clone() ISpanContext
}

// 对ddtrace的抽象， 以避免context直接对ddtrace依赖
type ICarrier interface {
	// 从ICarrier中把trace信息取出放到ISpanContext中
	Extract(ISpanContext)

	// 把trace信息从ISpanContext中取出注入到ICarrier中
	Inject(ISpanContext)
}

// ddtrace/tracer.DDTrace实现了此接口
type ITracer interface {
	// 把trace信息从ISpanContext中取出注入到ICarrier中
	Inject(ctx ISpanContext, carrier ICarrier) error

	// 从ICarrier中把trace信息取出放到ISpanContext中
	Extract(carrier ICarrier) (ISpanContext, error)

	GetEndPoint() (e *EndPoint)
	GetSampleRate() uint16
	ChildOf(name string, sc ISpanContext, opts ...SpanOpt) ISpan
	StartSpan(name string, opts ...SpanOpt) ISpan
	Close()
}

// ddtrace/tracer.DDSpan实现了此接口
type ISpan interface {
	Finish()
	Mark(spanKind SpanKind)
	Unmark(spanKind ...SpanKind) // 空表示unmark all  known SpanKind
	IsKind(spanKind SpanKind) bool
	SetName(name string)
	Context() ISpanContext
	SetContext(ISpanContext)
	SetTag(key, value string)
	GetTag(key string) (string, error)
	GetTraceID() string
	GetSpanID() string
	GetParentID() string
	GetFlag() string
	IsSampled() bool
	Tracer() ITracer
	ForceSample()
	SetEndPoint(e *EndPoint)
	GetEndPoint() (e *EndPoint)
}
type SpanOpt func(ISpan)

func WithEndPoint(ep *EndPoint) SpanOpt {
	return func(s ISpan) {
		s.SetEndPoint(ep)
	}
}
func WithSpanKind(spanKinds ...SpanKind) SpanOpt {
	return func(s ISpan) {
		for _, k := range spanKinds {
			s.Mark(k)
		}
	}
}
func WithSpanContextOpt(context ISpanContext) SpanOpt {
	return func(s ISpan) {
		s.SetContext(context)
	}
}

type EndPoint = model.EndPoint
