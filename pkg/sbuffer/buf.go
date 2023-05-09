package bytes

import (
	"io"
	"sync"
)

var buffer65536Pool = sync.Pool{
	New: func() any {
		return &[65536]byte{}
	},
}

func AcquireBuffer65536() *Buffer65536 {
	return &Buffer65536{
		p: buffer65536Pool.Get().(*[65536]byte),
	}
}

type Buffer65536 struct {
	p *[65536]byte
	r int
	w int
}

func (b *Buffer65536) Write(p []byte) (n int, _ error) {
	n = copy(b.p[b.w:], p)

	if b.w += n; b.w == len(b.p) {
		return n, io.EOF
	}
	return
}

func (b *Buffer65536) Read(p []byte) (n int, _ error) {
	n = copy(p, b.p[b.r:b.w])

	if b.r += n; b.r >= b.w {
		buffer65536Pool.Put(b.p)
		return n, io.EOF
	}
	return
}
