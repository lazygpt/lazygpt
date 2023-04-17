package log

import (
	"bytes"

	"github.com/hashicorp/go-hclog"
)

type BufferSinkAdapter struct {
	buf *bytes.Buffer
}

var _ hclog.SinkAdapter = (*BufferSinkAdapter)(nil)

func NewBufferSinkAdapter(buf *bytes.Buffer) *BufferSinkAdapter {
	return &BufferSinkAdapter{buf}
}

func (b *BufferSinkAdapter) Accept(name string, level hclog.Level, msg string, args ...interface{}) {
	// TODO(seanj): Ensure this buffer does not grow unbounded.
	b.buf.WriteString(msg)
}

func (b *BufferSinkAdapter) Buffer() *bytes.Buffer {
	return b.buf
}
