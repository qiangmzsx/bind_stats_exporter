package model

import (
	"strconv"

	"gitlab.luojilab.com/zeroteam/common/bit"
)

// SpanFlag 每一位具体的含义 见 本项目的 traceable/traceable.go文件开头处
type SpanFlag bit.BitInt // SpanFlag表示的是一个uint64的数字

type SpanKind uint64 //SpanKind 相当于按二进制

// SpanKind占Flag的第8~15位（以0为基）,即占第2字节的8位
const ( // 此处定义的值的含义，其实是bit位的索引，如值为8 ，表示用flag的第8位来表示其值
	span_kind_v1_pos8 SpanKind = 8 // 2#00000001 00000000
	span_kind_v1_pos9 SpanKind = 9 // 2#00000010 00000000
	// 为兼容老数据，第8 和第9位在新版中不再设置，但依然会读其值以使老数据有效

	//  所以以下新的kind从10开始
	SPAN_KIND_API   SpanKind = 10 // 2#00000100 00000000 ,目前在artemis/middleware/trace.go中有设置，应该是表示应该是表示这是一个api服务被请求了
	SPAN_KIND_ERROR SpanKind = 11 // 2#00001000 00000000
	SPAN_KIND_PANIC SpanKind = 12 // 2#00010000 00000000

	span_kind_max SpanKind = 15 // 2#10000000 00000000 SpanKind不能超过这个值，否则就越界了
)
const ( // DEBUG  SAMPLED_SET等3个flag对应的 pos
	DEBUG_KIND       SpanKind = 0 //  2#00000001 flag第0位表示debug
	SAMPLED_SET_KIND SpanKind = 1 //  2#00000010 flag第1位表示 SAMPLED_SET
	SAMPLED_KIND     SpanKind = 2 //  2#00000100 flag第2位表示 SAMPLED
)

// const (
// 	DEBUG       SpanFlag = 1 << 0 // 1
// 	SAMPLED_SET SpanFlag = 1 << 1 // 2
// 	SAMPLED     SpanFlag = 1 << 2 // 4
// )

func (flag SpanFlag) IsKind(spanKind SpanKind) bool {
	return bit.BitInt(flag).IsSet(uint64(spanKind))
}
func (flag *SpanFlag) Mark(spanKind SpanKind) {
	if spanKind == span_kind_v1_pos8 || spanKind == span_kind_v1_pos9 {
		// 新版不支持对老版所用的两个位置设置值
		return
	}
	b := bit.BitInt(*flag)
	b.Set(uint64(spanKind))
	*flag = SpanFlag(b)
}
func (flag *SpanFlag) Unmark(spanKinds ...SpanKind) {
	if len(spanKinds) == 0 {
		spanKinds = allspankinds
	}
	b := bit.BitInt(*flag)
	for _, k := range spanKinds {
		b.Unset(uint64(k))
	}
	*flag = SpanFlag(b)
}

func (flag SpanFlag) IsKindPanic() bool {
	return flag.IsKind(SPAN_KIND_PANIC)
}
func (flag SpanFlag) IsKindAPI() bool {
	if flag.IsKind(SPAN_KIND_API) {
		return true
	}
	if flag.IsKind(span_kind_v1_pos8) && !flag.IsKind(span_kind_v1_pos9) {
		return true
	}
	return false
}
func (flag SpanFlag) IsKindError() bool {
	if flag.IsKind(SPAN_KIND_ERROR) {
		return true
	}
	// 兼容老版flag数据
	if flag.IsKind(span_kind_v1_pos8) && flag.IsKind(span_kind_v1_pos9) {
		return true
	}
	return false
}
func (flag SpanFlag) GetSampleRate() uint16 {
	return bit.BitInt(flag).GetUint16(1)
}
func (flag SpanFlag) GetKindByte() uint8 {
	return bit.BitInt(flag).GetByte(1)
}
func (flag SpanFlag) GetSampleByte() uint8 {
	return bit.BitInt(flag).GetByte(0)
}

var allspankinds []SpanKind //在这个列表里的，当flag.Unmark()参数为0时，表示unmark 列表内指定位置的kind

func init() {
	for i := span_kind_v1_pos8; i <= span_kind_max; i++ {
		allspankinds = append(allspankinds, i)
	}
}
func (flag SpanFlag) GetHex() string {
	return strconv.FormatUint(uint64(flag), 16)
}
