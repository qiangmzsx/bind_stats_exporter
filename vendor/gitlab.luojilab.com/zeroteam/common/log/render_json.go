package log

import (
	"context"
	"fmt"
	"sync"

	"gitlab.luojilab.com/zeroteam/common/log/core"
)

func newJsonRender(w ...*writer) (a *jsonRender) {
	a = &jsonRender{
		enc: core.NewJSONEncoder(core.EncoderConfig{
			EncodeTime:     core.EpochTimeEncoder,
			EncodeDuration: core.SecondsDurationEncoder,
		}, core.NewBuffer(0)),
		ws:      w,
		bufPool: core.NewPool(2048),
	}
	a.fieldPool.New = func() interface{} {
		return make([]core.Field, 0, 16)
	}
	return
}

type jsonRender struct {
	fieldPool sync.Pool
	enc       core.Encoder
	ws        []*writer
	bufPool   core.Pool
}

func (r *jsonRender) Render(ctx context.Context, depth int, lv Level, src string, wt writerType, args ...Pair) {
	if args == nil {
		return
	}
	f := r.data()
	for i := range args {
		f = append(f, args[i])
	}

	buf := r.bufPool.Get()
	r.enc.Encode(buf, f...)
	r.free(f)
	var err error
	for _, w := range r.ws {
		_, err = w.Write(lv, wt, buf.Bytes())
		if err != nil {
			fmt.Println("jsonRender.Write error", err)
		}
	}

	r.bufPool.Put(buf)
}

func (h *jsonRender) data() []core.Field {
	return h.fieldPool.Get().([]core.Field)
}

func (h *jsonRender) free(f []core.Field) {
	f = f[0:0]
	h.fieldPool.Put(f)
}
