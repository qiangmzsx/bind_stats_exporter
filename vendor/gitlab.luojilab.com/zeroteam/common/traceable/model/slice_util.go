package model

import (
	"sync"
)

type SpanBuffer struct {
	mux sync.RWMutex
	s   []Span
	idx int
}

func NewSpanBuffer(size int) *SpanBuffer {
	c := &SpanBuffer{}
	c.s = make([]Span, size, size)
	return c
}
func (c *SpanBuffer) AppendNoAlloc(f func(s *Span) bool) int {
	c.mux.Lock()
	defer c.mux.Unlock()
	if c.idx == len(c.s) {
		c.s = append(c.s, Span{})
	}

	succ := f(&c.s[c.idx])
	if succ {
		c.idx++
	}
	return c.idx
}
func (c *SpanBuffer) Append(item Span) int {
	c.mux.Lock()
	defer c.mux.Unlock()
	if c.idx == len(c.s) {
		c.s = append(c.s, item)
		c.idx++
		return len(c.s)
	}

	c.s[c.idx] = item
	c.idx++
	return c.idx
}

func (c *SpanBuffer) Clear() {
	c.mux.RLock()
	c.idx = 0
	c.mux.RUnlock()
}
func (c *SpanBuffer) Range(f func(s *Span) bool) {
	c.mux.RLock()
	defer c.mux.RUnlock()
	for i := 0; i < c.idx; i++ {
		if f(&c.s[c.idx]) {
			break
		}
	}
}
func (c *SpanBuffer) Callback(f func(list []Span)) {
	c.mux.RLock()
	defer c.mux.RUnlock()
	f(c.s[0:c.idx])
}
func (c *SpanBuffer) Get() []Span {
	c.mux.RLock()
	defer c.mux.RUnlock()
	r := make([]Span, c.idx)
	copy(r, c.s[0:c.idx])
	return r
}
func (c *SpanBuffer) GetByIdx(idx int) (s Span, ok bool) {
	c.mux.RLock()
	defer c.mux.RUnlock()
	if idx < c.idx {
		return c.s[idx], true
	}
	return s, false

}

func (c *SpanBuffer) Cut() []Span {
	c.mux.Lock()
	defer c.mux.Unlock()
	r := make([]Span, c.idx)
	copy(r, c.s[0:c.idx])
	c.idx = 0
	return r
}

//cut [start,start+len)
func (c *SpanBuffer) CutBuffer(start, length int) ([]Span, int) {
	c.mux.Lock()
	defer c.mux.Unlock()
	l := c.idx
	if length == 0 || start > l-1 || start < 0 {
		return []Span{}, 0
	}
	if start+length > l-1 {
		length = l - start
	}
	r := make([]Span, length)
	copy(r, c.s[start:start+length])
	// c.s = append(c.s[0:start], c.s[start+length:]...)
	for i := start + length; i < c.idx; i++ {
		c.s[i-length] = c.s[i]
	}
	c.idx = c.idx - length

	return r, length
}
func (c *SpanBuffer) Size() int {
	c.mux.RLock()
	defer c.mux.RUnlock()
	return c.idx
}
