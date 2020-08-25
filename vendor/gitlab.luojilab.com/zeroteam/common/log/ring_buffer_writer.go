package log

import (
	"fmt"
	"io"
	"runtime"
	"sync"
	"time"

	"sync/atomic"

	"gitlab.luojilab.com/zeroteam/common/ringbuffer"
)

// 日志落盘的时间间隔
var FlushInterval = 3 * time.Second

// FileBufferSize sizes the buffer associated with each log file. It's large
// so that log records can accumulate without the logging thread blocking
// on disk I/O. The flushDaemon will block instead.
// 日志文件的buffer size
var FileBufferSize = 1024 * 1024 * 5

// 如果ringbuffer满了，是否drop此条日志,若不drop，则允许ringbuffer 临时扩容，
// 至下次刷盘时，会恢复其内存使用量
// var DropIfBufferFull = true

type MutexRingBuffer struct {
	//RingBuffer 的size
	bufSize int
	rb      *ringbuffer.RingBuffer

	mu sync.Mutex
}

// bufSize 初始buffer大小，
func newMutexRingBuffer(bufSize int) *MutexRingBuffer {
	return &MutexRingBuffer{
		rb:      ringbuffer.New(bufSize),
		bufSize: bufSize,
	}

}

func (rb *MutexRingBuffer) Write(buf []byte, dropIfFull bool) (n, len1, cap1 int, err error) {
	rb.mu.Lock()
	if dropIfFull { // 如果ringbuffer满了，则主动丢弃此条日志
		n, err = rb.rb.WriteNotAutoExtend(buf)
	} else {
		n, err = rb.rb.Write(buf)
	}

	len1 = rb.rb.Length()
	cap1 = rb.rb.Capacity()

	rb.mu.Unlock()
	return n, len1, cap1, err
}

func (rb *MutexRingBuffer) WriteTo(w io.Writer, checkMaxBufSize bool) (n int64, err error) {
	rb.mu.Lock()
	n, err = rb.rb.WriteTo(w)
	if checkMaxBufSize {
		rb.checkMaxSize()
	}
	rb.mu.Unlock()
	return
}

func (rb *MutexRingBuffer) checkMaxSize() {
	if rb.rb.Capacity() > rb.bufSize {
		rb.rb.Reset(rb.bufSize)
	}
}

const (
	ratio = 3
)

// bufSize 初始buffer大小，
func newRingBufferWriter(file syncWriter, bufSize int, dropIfBufferFull bool) *RingBufferWriter {
	w := &RingBufferWriter{
		// 实际bufsize为totalSize:=bufSize*ratio
		// ,当buf使用量达到 totalSize/ratio时，即bufSize时，就会尝试进行刷盘操作
		// 给buf一量的缓冲量，
		rb:               newMutexRingBuffer(bufSize * ratio),
		file:             file,
		flushChan:        make(chan bool),
		bufPool:          newBufferPool(0, 5),
		stop:             make(chan bool),
		dropIfBufferFull: dropIfBufferFull,
	}
	go w.daemon()
	return w
}

type RingBufferWriter struct {
	rb               *MutexRingBuffer
	flushChan        chan bool
	isFlushing       int32
	mu               sync.Mutex
	file             syncWriter
	bufPool          *bufferpool
	stop             chan bool
	isStopping       bool
	dropIfBufferFull bool
}

func (w *RingBufferWriter) isFlashing() (flashing bool) {
	return atomic.LoadInt32(&w.isFlushing) == 1
}
func (w *RingBufferWriter) setFlashing(flashing bool) {
	if flashing {
		w.isFlushing = atomic.SwapInt32(&w.isFlushing, 0)
	} else {
		w.isFlushing = atomic.SwapInt32(&w.isFlushing, 1)
	}
}
func (w *RingBufferWriter) daemon() {
	defer func() {
		err := recover()
		if err != nil {
			fmt.Println(err)
			for i := 1; i < 20; i++ {
				_, file, line, ok := runtime.Caller(i)
				if ok {
					fmt.Errorf("%v:[file:%v,line:%v]\n", i, file, line)
				}
			}
		}
	}()

	interval := time.Second
	ticker := time.NewTicker(interval) // 以1秒初始，当FlushInterval改变时可以及早发现其变更
	for {
		select {
		case <-w.flushChan:
			w.Flush(false)
		case <-ticker.C:
			atomic.AddInt64(&stats.TimerFlushCount, 1)
			w.Flush(false)
			if interval != FlushInterval {
				ticker.Stop()
				ticker = time.NewTicker(FlushInterval)
				interval = FlushInterval
			}

		case <-w.stop:
			w.Flush(true)
			ticker.Stop()
			return
		}

		if w.isStopping {
			ticker.Stop()
			return
		}
	}
}

func (w *RingBufferWriter) Stopping() {
	w.isStopping = true
	select {
	case w.stop <- true:
	case <-time.After(time.Second):
	}
}

func (w *RingBufferWriter) Flush(sync bool) (err error) {
	w.setFlashing(true)

	buf := w.bufPool.getBuffer()
	_, err = w.rb.WriteTo(buf, true)
	if err != nil {
		if err == ringbuffer.ErrIsEmpty {
			err = nil
		}
		goto exit
	}

	w.mu.Lock()
	_, err = buf.WriteTo(w.file)
	if sync {
		err = w.file.Sync()
	}

	w.mu.Unlock()

exit:
	w.bufPool.putBuffer(buf)
	w.setFlashing(false)
	return err
}

// 通知守护进程刷盘,但不等待其是否刷盘完成，也不保证对方一定收到通知
// 只是尝试让其刷一下
func (w *RingBufferWriter) castFlushing() {
	if w.isFlashing() {
		return
	}

	select {
	case w.flushChan <- true:
		atomic.AddInt64(&stats.BufferFullFlushCount, 1)
	default:
		atomic.AddInt64(&stats.FlushBusyCount, 1)

	}
}

// write到ringbuffer
func (w *RingBufferWriter) Write(p []byte) (n int, err error) {
	if w.isStopping {
		return
	}

	n, len1, cap1, err := w.rb.Write(p, w.dropIfBufferFull)
	if len1*ratio > cap1 { // 如果buf已经用了1/2，则通知其刷盘
		w.castFlushing()
	}

	if err == ringbuffer.ErrIsFull {
		atomic.AddInt64(&stats.DropCount, 1)
		err = nil
	}

	return
}
