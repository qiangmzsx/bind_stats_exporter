package log

import (
	"fmt"
	"io"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

// flushSyncWriter is the interface satisfied by logging destinations.
type flushSyncWriter interface {
	Flush(sync bool) error
	io.Writer
}
type writerType int

const (
	writerTypeFile   writerType = 1
	writerTypeStderr writerType = 2
)

type Stats struct {
	WriteCount           int64 // 调用Write次数
	BufferFullFlushCount int64 //当ringbuffer的使用量达到BufferSize时，强制flush次数
	TimerFlushCount      int64 // 定时flsuh调用次数
	DropCount            int64 // buffer满，而丢弃的log数
	FlushBusyCount       int64 //尝试flush时，worker进程busy
	BufferSize           int   // BufferSize通常是BufferSizeThreshold的1/3
	// BufferSizeThreshold通常是BufferSize的三倍，ringbuffer的真正容量是BufferSizeThreshold
	// 当ringbuffer满时，再往ringbuffer写日志时，会直接丢弃之
	// 但通常ringbuffer不会满，当ringbuffer使用量达到1/3时，就会通知守护进程读ringbuffer并刷盘
	//  故除非刷盘卡顿极大，且日志量极大，很少会出现ringbuffer写满的情形
	BufferSizeThreshold int
	FlushInterval       time.Duration
	WaitingWriteMS      int64
}

var stats Stats

func (s Stats) String() string {
	return fmt.Sprintf(`{"Desc":"Log Stats Info","WriteCount":%d,"DropCount":%d,"FlushBusyCount":%d,"BufferFullFlushCount":%d,"TimerFlushCount":%d,"BufferSize":"%dk","BuferfSizeThreshold":%dk,"FlushInterval":"%ds","WaitingWriteMS":%d}`,
		s.WriteCount, s.DropCount, s.FlushBusyCount, s.BufferFullFlushCount, s.TimerFlushCount, s.BufferSize/1024, s.BufferSizeThreshold/1024, int(s.FlushInterval/time.Second), s.WaitingWriteMS/1000)
}

func GetStats() Stats {
	stats.BufferSize = FileBufferSize
	stats.FlushInterval = FlushInterval
	stats.BufferSizeThreshold = FileBufferSize * ratio
	return stats
}

type noopflushSyncWriter struct {
}

func (sb *noopflushSyncWriter) Flush(sync bool) error {
	return nil
}

func (sb *noopflushSyncWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

type writer struct {
	logdir         string
	symlinks       []string
	filenamePrefix string
	filename       string
	// mu protects the remaining elements of this structure and is
	// used to synchronize logging.
	mu sync.Mutex
	// file holds writer for each of the log types.
	files            [NumLevel]flushSyncWriter
	isclosed         int32
	writeToSameFile  bool //是否根据日志等级，写到不同的日志文件中
	logStats         bool
	dropIfBufferFull bool
}

func newWriter(logdir string, filenamePrefix string, filename string, symlinks []string,
	writeToSameFile, logStats, dropIfBufferFull bool) *writer {
	fw := &writer{
		filename:         filename,
		logdir:           logdir,
		symlinks:         symlinks,
		filenamePrefix:   filenamePrefix,
		writeToSameFile:  writeToSameFile,
		logStats:         logStats,
		dropIfBufferFull: dropIfBufferFull,
	}
	if filename != "" { //当指定了特定的log文件名后，必然只会写入到这一个日志文件
		fw.writeToSameFile = true
	}

	return fw
}

func (l *writer) writeStderr(data []byte) (n int, err error) {
	l.mu.Lock()
	n, err = os.Stderr.Write(data)
	l.mu.Unlock()
	return
}
func (l *writer) writeLevel(lv Level, data []byte) (n int, err error) {
	w := l.files[lv]
	if w == nil {
		l.mu.Lock()
		w = l.files[lv]
		if w == nil { // double check after get lock
			if err = l.createFiles(lv); err != nil {
				os.Stderr.Write(data) // Make sure the message appears somewhere.
				l.mu.Unlock()
				return 0, err
			}
			w = l.files[lv]
		}
		l.mu.Unlock()
	}
	n, err = w.Write(data)
	if err != nil {
		l.mu.Lock()
		os.Stderr.Write(data) // Make sure the message appears somewhere.
		l.mu.Unlock()
		return n, err
	}
	return
}
func (l *writer) Write(lv Level, wt writerType, data []byte) (n int, err error) {
	if wt == writerTypeFile && l.logdir != "" {
		// 如果已经关闭，则将日志打印到stderr
		if atomic.LoadInt32(&l.isclosed) == 1 {
			return l.writeStderr(data)
		}
		switch lv {
		case FatalLevel:
			n, err = l.writeLevel(FatalLevel, data)
			fallthrough
		case ErrorLevel:
			n, err = l.writeLevel(ErrorLevel, data)
			fallthrough
		case WarnLevel:
			n, err = l.writeLevel(WarnLevel, data)
			fallthrough
		case InfoLevel:
			n, err = l.writeLevel(InfoLevel, data)
		}
		atomic.AddInt64(&stats.WriteCount, 1)
	}
	if wt == writerTypeStderr {
		return l.writeStderr(data)
	}
	return
}

// createFiles creates all the log files for Level from lv down to _infoLevel.
// l.mu is held.
func (l *writer) createFiles(lv Level) error {
	// Files are created in decreasing Level order, so as soon as we find one
	// has already been created, we can stop.
	if l.writeToSameFile && lv != InfoLevel {
		l.files[lv] = &noopflushSyncWriter{}
	} else {
		fr := &filerotator{
			logdir:          l.logdir,
			symlinks:        l.symlinks,
			filenamePrefix:  l.filenamePrefix,
			filename:        l.filename,
			lv:              lv,
			writeToSameFile: l.writeToSameFile,
			logStats:        l.logStats,
		}
		fr.rotateFile(true, time.Now())
		sb := newRingBufferWriter(fr, FileBufferSize, l.dropIfBufferFull)
		l.files[lv] = sb
	}
	return nil
}

// Flush is like flushAll but locks l.mu first.
func (l *writer) Flush() {
	l.mu.Lock()
	l.flushAll(false)
	l.mu.Unlock()
}

// flushAll flushes all the logs and attempts to "sync" their data to disk.
// l.mu is held.
func (l *writer) flushAll(flushStats bool) {
	// Flush from fatal down, in case there's trouble flushing.
	for s := FatalLevel; s >= InfoLevel; s-- {
		file := l.files[s]
		if file != nil {
			if flushStats && l.logStats {
				file.Write([]byte(GetStats().String() + "\n"))
			}

			file.Flush(true) // ignore error
		}
	}
}

func (l *writer) Close() {
	l.mu.Lock()
	atomic.SwapInt32(&l.isclosed, 1)
	l.flushAll(true)
	l.mu.Unlock()
}
