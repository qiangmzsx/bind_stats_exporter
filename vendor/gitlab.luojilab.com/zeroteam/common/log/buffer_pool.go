package log

import (
	"bytes"
	"sync"
)

type bufferpool struct {
	// freeList is a list of byte buffers, maintained under freeListMu.
	freeList  *buffer
	bufCnt    int
	maxBufCnt int

	// freeListMu maintains the free list. It is separate from the main mutex
	// so buffers can be grabbed and printed to without holding the main lock,
	// for better parallelization.
	freeListMu sync.Mutex
	bigBufSize int
}

// 如果buffersize大于bigBufSize后， 则不会pool之
// 为0 则依然pool之
// 若maxBufCnt非0，则最多只pool maxBufCnt个buffer,为0则不限制
func newBufferPool(bigBufSize, maxBufCnt int) *bufferpool {
	return &bufferpool{bigBufSize: bigBufSize, maxBufCnt: maxBufCnt}
}

// getBuffer returns a new, ready-to-use buffer.
func (l *bufferpool) getBuffer() *buffer {
	l.freeListMu.Lock()
	b := l.freeList
	if b != nil {
		l.freeList = b.next
		l.bufCnt--
	}
	l.freeListMu.Unlock()
	if b == nil {
		b = new(buffer)
	} else {
		b.next = nil
		b.Reset()
	}
	return b
}

// putBuffer returns a buffer to the free list.
func (l *bufferpool) putBuffer(b *buffer) {
	if l.bigBufSize > 0 && b.Len() >= l.bigBufSize {
		// Let big buffers die a natural death.
		return
	}
	if l.maxBufCnt > 0 && l.bufCnt >= l.maxBufCnt {
		return
	}

	l.freeListMu.Lock()
	b.next = l.freeList
	l.bufCnt++
	l.freeList = b
	l.freeListMu.Unlock()
}

type buffer struct {
	bytes.Buffer
	tmp  [64]byte // temporary byte array for creating headers.
	next *buffer
}

const digits = "0123456789"

// twoDigits formats a zero-prefixed two-digit integer at buf.tmp[i].
func (buf *buffer) twoDigits(i, d int) {
	buf.tmp[i+1] = digits[d%10]
	d /= 10
	buf.tmp[i] = digits[d%10]
}

// nDigits formats an n-digit integer at buf.tmp[i],
// padding with pad on the left.
// It assumes d >= 0.
func (buf *buffer) nDigits(n, i, d int, pad byte) {
	j := n - 1
	for ; j >= 0 && d > 0; j-- {
		buf.tmp[i+j] = digits[d%10]
		d /= 10
	}
	for ; j >= 0; j-- {
		buf.tmp[i+j] = pad
	}
}

// someDigits formats a zero-prefixed variable-width integer at buf.tmp[i].
func (buf *buffer) someDigits(i, d int) int {
	// Print into the top, then copy down. We know there'lv space for at least
	// a 10-digit number.
	j := len(buf.tmp)
	for {
		j--
		buf.tmp[j] = digits[d%10]
		d /= 10
		if d == 0 {
			break
		}
	}
	return copy(buf.tmp[i:], buf.tmp[j:])
}
