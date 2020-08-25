package log

import (
	"fmt"
	"time"

	"gitlab.luojilab.com/zeroteam/common/log/core"
)

// Pair represents a map of entry level data used for structured logging.
// type Pair map[string]interface{}
type Pair struct {
	Key   string
	Value interface{}
}

// KV return a log kv for logging field.
func KV(key string, value interface{}) Pair {
	return Pair{
		Key:   key,
		Value: value,
	}
}

// AddTo exports a field through the ObjectEncoder interface. It's primarily
// useful to library authors, and shouldn't be necessary in most applications.
func (d Pair) AddTo(enc core.ObjectEncoder) {
	var err error
	switch val := d.Value.(type) {
	case bool:
		enc.AddBool(d.Key, val)
	case complex128:
		enc.AddComplex128(d.Key, val)
	case complex64:
		enc.AddComplex64(d.Key, val)
	case float64:
		enc.AddFloat64(d.Key, val)
	case float32:
		enc.AddFloat32(d.Key, val)
	case int:
		enc.AddInt(d.Key, val)
	case int64:
		enc.AddInt64(d.Key, val)
	case int32:
		enc.AddInt32(d.Key, val)
	case int16:
		enc.AddInt16(d.Key, val)
	case int8:
		enc.AddInt8(d.Key, val)
	case string:
		enc.AddString(d.Key, val)
	case uint:
		enc.AddUint(d.Key, val)
	case uint64:
		enc.AddUint64(d.Key, val)
	case uint32:
		enc.AddUint32(d.Key, val)
	case uint16:
		enc.AddUint16(d.Key, val)
	case uint8:
		enc.AddUint8(d.Key, val)
	case []byte:
		enc.AddByteString(d.Key, val)
	case uintptr:
		enc.AddUintptr(d.Key, val)
	case time.Time:
		enc.AddTime(d.Key, val)
	case time.Duration:
		enc.AddDuration(d.Key, val)
	case error:
		enc.AddString(d.Key, val.Error())
	case fmt.Stringer:
		enc.AddString(d.Key, val.String())
	default:
		err = enc.AddReflected(d.Key, val)
	}

	if err != nil {
		enc.AddString(fmt.Sprintf("%sError", d.Key), err.Error())
	}
}
