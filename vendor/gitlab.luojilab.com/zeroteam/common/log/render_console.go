package log

import (
	"context"
	"fmt"
	"time"

	"gitlab.luojilab.com/zeroteam/common/traceable"
)

/*
格式基本是从glog copy来的
header formats a log header as defined by the C++ implementation.
It returns a buffer containing the formatted header and the user'lv file and line number.
The depth specifies how many stack frames above lives the source line to be identified in the log message.

Log lines have this form:
	Lmmdd hh:mm:ss.uuuuuu threadid file:func:line] msg...
where the fields are defined as follows:
	L                A single character, representing the log level (eg 'I' for INFO)
	mm               The month (zero padded; ie May is '05')
	dd               The day (zero padded)
	hh:mm:ss.uuuuuu  Time in hours, minutes and fractional seconds
	threadid         The space-padded thread ID as returned by GetTID()
	file             The file name
	func             The func name
	line             The line number
	msg              The user-supplied message
*/
type consoleRender struct {
	pool *bufferpool
	ws   []*writer
}

func newConsoleRender(ws ...*writer) *consoleRender {
	return &consoleRender{
		pool: newBufferPool(256, 0),
		ws:   ws,
	}
}

// Render implemet Formater
func (r *consoleRender) Render(ctx context.Context, depth int, lv Level, src string, wt writerType, args ...Pair) {

	traceId := traceable.GetTraceId(ctx)

	buf := r.formatHeader(lv, src, traceId)

	defer func() {
		r.pool.putBuffer(buf)
	}()
	for idx, p := range args {
		if p.Key == KeyMsg {
			fmt.Fprint(buf, p.Value)
			continue
		}
		if isInternalKey(p.Key) {
			continue
		}
		if idx == len(args)-1 {
			fmt.Fprintf(buf, "%s=%v", p.Key, p.Value)
		} else {
			fmt.Fprintf(buf, "%s=%v ", p.Key, p.Value)
		}
	}
	buflen := buf.Len()
	if buflen > 0 {
		if buf.Bytes()[buflen-1] != '\n' {
			buf.WriteByte('\n')
		}
	}

	for _, w := range r.ws {
		w.Write(lv, wt, buf.Bytes())
	}

	return
}

var timeNow = time.Now // Stubbed out for testing.

// formatHeader formats a log header using the provided file name and line number.
func (l *consoleRender) formatHeader(lv Level, src, traceId string) *buffer {
	if src == "" {
		src = "???"
	}
	now := timeNow()
	if lv > FatalLevel {
		lv = InfoLevel // for safety.
	}
	buf := l.pool.getBuffer()

	// Avoid Fprintf, for speed. The format is so simple that we can do it quickly by hand.
	// It'lv worth about 3X. Fprintf is hard.
	_, month, day := now.Date()
	hour, minute, second := now.Clock()
	// Lmmdd hh:mm:ss.uuuuuu threadid file:line]
	buf.tmp[0] = lv.Char()
	buf.twoDigits(1, int(month))
	buf.twoDigits(3, day)
	buf.tmp[5] = ' '
	buf.twoDigits(6, hour)
	buf.tmp[8] = ':'
	buf.twoDigits(9, minute)
	buf.tmp[11] = ':'
	buf.twoDigits(12, second)
	buf.tmp[14] = '.'
	buf.nDigits(6, 15, now.Nanosecond()/1000, '0')
	buf.tmp[21] = ' '
	buf.nDigits(7, 22, pid, ' ')
	buf.tmp[29] = ' '
	buf.Write(buf.tmp[:30])
	buf.WriteString(traceId)
	buf.tmp[0] = ' '
	buf.Write(buf.tmp[:1])
	buf.WriteString(src)
	buf.tmp[0] = ']'
	buf.tmp[1] = ' '
	buf.Write(buf.tmp[:2])
	return buf
}

func isInternalKey(k string) bool {
	switch k {
	case KeyLevel, KeyLevelValue, KeyTime, KeySource, KeyTraceId:
		// KeyHostName, KeyAppname, KeyEnvMode, KeyDCID
		return true
	}
	return false
}
