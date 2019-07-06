// Package safe provides thread-safe objects for internal use.
package safe

import (
	"bytes"
	"sync"
)

// Buffer provides a thread-safe bytes.Buffer.
type Buffer struct {
	buf bytes.Buffer
	mu  sync.Mutex
}

// Read provides a thread-safe bytes.Buffer.Read function.
func (s *Buffer) Read(p []byte) (n int, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.Read(p)
}

// Write provides a thread-safe bytes.Buffer.Write function.
func (s *Buffer) Write(p []byte) (n int, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.Write(p)
}

// Len provides a thread-safe bytes.Buffer.Len function.
func (s *Buffer) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.Len()
}
