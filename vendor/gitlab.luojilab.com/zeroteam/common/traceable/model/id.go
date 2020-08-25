package model

import (
	"strconv"
	"strings"

	"gitlab.luojilab.com/zeroteam/common/idgen"
)

// traceid 与spanid 相同不会有问题，故用两个id 生成器，避免用一个 1s 内产生的id 数量过多导致id 生成器超载
// 1年=365*24*3600s=2#1 11100001 00110011 10000000 25 位
//10.6.29.57 对 a 类内网ip timestamp_bit_count=25,ip_bit_count=24,seq_bit_count=15,max_id_each_second=32768
var traceIDGen = idgen.NewIDGen(idgen.WithTimestampBitCount(24)) // 时间戳占用25 位，1 年内不重复即可，1 年后，循环往复，对trace 这种场景可以接受
var spanIDGen = idgen.NewIDGen(idgen.WithTimestampBitCount(24))

func GetNewSpanID() uint64 {
	return spanIDGen.Generate()
}
func GetNewTraceID() uint64 {
	return traceIDGen.Generate()
}

// 可以从json中解析负数的uint64

// A IDV1 represents a JSON number literal.
type IDV1 uint64

func NewIDV1(n uint64) IDV1 {
	return IDV1(n)
}

// String returns the literal text of the number.
func (n IDV1) String() string {
	return strconv.FormatUint(uint64(n), 10)
}

// Hex returns the literal text of the number.
func (n IDV1) Hex() string {
	return strconv.FormatUint(uint64(n), 16)
}

// Int64 returns the number as an int64.
func (n IDV1) Int64() int64 {
	return int64(n)
}
func (n IDV1) Uint64() uint64 {
	return uint64(n)
}

func (jt *IDV1) UnmarshalJSON(b []byte) error {
	str := string(b)
	if str == `""` {
		return nil
	}
	if str == `"0"` {
		return nil
	}

	trimStr := strings.Trim(str, "\"")
	if trimStr == "" {
		return nil
	}
	if trimStr == "0" {
		return nil
	}

	if trimStr[0] == '-' {
		i, err := strconv.ParseInt(trimStr, 0, 64)
		if err == nil {
			*jt = IDV1(uint64(i))
			return nil
		}
		return err
	} else {
		i, err := strconv.ParseUint(trimStr, 0, 64)
		if err == nil {
			*jt = IDV1(i)
			return nil
		}
		return err

	}

}
