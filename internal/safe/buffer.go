package safe

import (
	"bytes"
	"sync"
)

type Buffer struct {
	buf bytes.Buffer
	mu  sync.Mutex
}

func (s *Buffer) Read(p []byte) (n int, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.Read(p)
}

func (s *Buffer) Write(p []byte) (n int, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.Write(p)
}

func (s *Buffer) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.Len()
}
