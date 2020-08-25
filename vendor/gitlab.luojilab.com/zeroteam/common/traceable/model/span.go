package model

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

type Span struct {
	// E0705 16:34:25.047549   24426  kafka_collector.go:149] decode span error cause :json: cannot unmarshal number 471933144554924152069 into Go struct field Span.traceid of type int64{"name":"GET /share/course/article/article_id/81746","traceid":471933144554924152069,"id":3251646576875649693,"parentid":0,"tags":[{"key":"http.client.ip","value":"::ffff:172.19.28.5","timestamp":1562315662576000},{"key":"http.response.code","value":"200","timestamp":1562315662595000}],"timestamp":1562315662576000,"duration":19000,"endpoint":{"servicename":"iget_share_v3","port":7775,"ipv4":-1408019567,"ipv6":""},"flag":65798}
	TraceID   IDV1     `json:"traceid"` // 本来是int64,但是线上会出好多cause :json: cannot unmarshal number 471933144554924152069 into Go struct field Span.traceid
	Name      string   `json:"name"`
	ID        IDV1     `json:"id"`
	ParentID  IDV1     `json:"parentid"`
	Tags      []Tag    `json:"tags"`
	Timestamp int64    `json:"timestamp"` //measured in microseconds
	Duration  int64    `json:"duration"`  //measured in microseconds,at least 1 μs
	EndPt     EndPoint `json:"endpoint"`
	FLAG      SpanFlag `json:"flag"` // 具体含义见span_flag.go文件
}

func (s *Span) GetLogKV() []interface{} {
	return []interface{}{
		"traceid", s.TraceID.Uint64(),
		"name", s.Name,
		"id", s.ID.Uint64(),
		"parentid", s.ParentID.Uint64(),
		"tags", s.Tags,
		"timestamp", s.Timestamp,
		"duration", s.Duration,
		"endpoint", s.EndPt,
		"flag", uint64(s.FLAG),
	}
}
func (s *Span) SetID(id uint64) {
	s.ID = NewIDV1(id)
}
func (s Span) GetTraceIDUint64() (traceID uint64) {
	return s.TraceID.Uint64()
}
func (s *Span) SetTraceID(traceID uint64) {
	s.TraceID = NewIDV1(traceID)
}
func (s Span) GetParentIDUint64() (parentID uint64) {
	return s.ParentID.Uint64()
}
func (s Span) GetSpanIDUint64() (id uint64) {
	return s.ID.Uint64()
}
func (s *Span) SetParentID(parentID uint64) {
	s.ParentID = NewIDV1(parentID)
}

func (s Span) GetTraceID() string {
	return strconv.FormatUint(s.TraceID.Uint64(), 16)
}
func (s Span) GetSpanID() string {
	return strconv.FormatUint(s.ID.Uint64(), 16)
}
func (s Span) GetParentID() string {
	return strconv.FormatUint(s.ParentID.Uint64(), 16)
}

func (span Span) IsKindError() bool {
	if span.FLAG.IsKindError() || span.FLAG.IsKindPanic() {
		return true
	}

	if span.Tags == nil || len(span.Tags) == 0 {
		return false
	}

	for _, tag := range span.Tags {
		if tag.Key == "error" {
			return true
		}

		if tag.Key != "http.response.code" {
			continue
		}
		code, err := strconv.Atoi(tag.Value)
		if err != nil {
			err = fmt.Errorf("cover string to int with strconv.Atoi:%s", err.Error())
			return false
		}
		if code >= http.StatusInternalServerError {
			return true
		}
	}
	return false
}
func (span Span) String() string {
	return span.ToString(false)
}
func (span Span) ToString(indent bool) string {
	if indent {
		data, _ := json.MarshalIndent(span, "", "  ")
		return string(data)
	}
	data, _ := json.Marshal(span)
	return string(data)
}
func (span Span) ToESSpanErrorType() int { //EsSpan.ErrorType
	if span.FLAG.IsKindPanic() {
		return 2
	}
	if span.IsKindError() {
		return 1
	}
	return 0

}
func (s *Span) Clear() {
	s.TraceID = 0
	s.Name = ""
	s.ID = 0
	s.ParentID = 0
	s.Tags = s.Tags[:0]
	s.Timestamp = 0
	s.Duration = 0
	s.FLAG = 0
}

type EsSpan struct {
	TraceID   string   `json:"traceid"`
	Name      string   `json:"name"`
	ID        string   `json:"id"`
	ParentID  string   `json:"parentid"`
	Tags      []Tag    `json:"tags"`
	Timestamp int64    `json:"timestamp"` //measured in microseconds
	Duration  int64    `json:"duration"`  //measured in microseconds,at least 1 μs
	EndPt     EndPoint `json:"endpoint"`
	FLAG      string   `json:"flag"`
	Kind      int      `json:"kind"`
	Sr        int      `json:"sr"`
	ErrorType int      `json:"err"` // 0非error类型,1 普通error ,2 panic
}

func (s *EsSpan) Clear() {
	s.TraceID = ""
	s.Name = ""
	s.ID = ""
	s.ParentID = ""
	s.Tags = s.Tags[:0]
	s.Timestamp = 0
	s.Duration = 0
	s.FLAG = ""
	s.Kind = 0
	s.Sr = 0
	s.ErrorType = 0
}
