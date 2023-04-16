package log

import (
	"io"
	"sync"
)

type ReplaceableOutput struct {
	writer io.Writer
	mu     sync.Mutex
}

var _ io.Writer = (*ReplaceableOutput)(nil)

func NewReplaceableOutput(writer io.Writer) *ReplaceableOutput {
	return &ReplaceableOutput{writer, sync.Mutex{}}
}

func (ro *ReplaceableOutput) Write(p []byte) (n int, err error) {
	ro.mu.Lock()
	defer ro.mu.Unlock()

	return ro.writer.Write(p)
}

func (ro *ReplaceableOutput) ReplaceWriter(writer io.Writer) {
	ro.mu.Lock()
	defer ro.mu.Unlock()

	ro.writer = writer
}
