package log

// import "io"

// func GetLoggerWriter(l Logger) io.Writer {
// 	return &writerWrapper{l}
// }

// type writerWrapper struct {
// 	l Logger
// }

// func (w *writerWrapper) Write(data []byte) (n int, err error) {
// 	w.l.InfoKV(KeyMsg, data)
// 	return len(data), nil

// }
