package model

type Tag struct {
	Key       string `json:"key"`
	Value     string `json:"value"`
	Timestamp int64  `json:"timestamp"` //measured in microseconds
}

const (
	DD_CLIENT_SEND        = "cs"
	DD_CLIENT_RECV        = "cr"
	DD_SERVER_SEND        = "ss"
	DD_SERVER_RECV        = "sr"
	DD_WIRE_SEND          = "ws"
	DD_WIRE_RECV          = "wr"
	DD_HTTP_HOST          = "http.host"
	DD_HTTP_METHOD        = "http.method"
	DD_HTTP_PATH          = "http.path"
	DD_HTTP_URL           = "http.url"
	DD_HTTP_STATUS_CODE   = "http.status_code"
	DD_HTTP_REQUEST_SIZE  = "http.request.size"
	DD_HTTP_RESPONSE_SIZE = "http.response.size"
	//DD_SPAN_ERROR         = "error"
)

func IsPredefineTagType(name string) bool {
	switch name {
	case DD_CLIENT_SEND, DD_CLIENT_RECV, DD_SERVER_SEND, DD_SERVER_RECV:
		fallthrough
	case DD_WIRE_SEND, DD_WIRE_RECV:
		//	fallthrough
		//case DD_HTTP_HOST, DD_HTTP_METHOD, DD_HTTP_PATH, DD_HTTP_URL, DD_HTTP_STATUS_CODE, DD_HTTP_REQUEST_SIZE, DD_HTTP_RESPONSE_SIZE:
		//	fallthrough
		//case DD_SPAN_ERROR:
		return true
	default:
		return false
	}
}
