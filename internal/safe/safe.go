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

// Func holds a func([]byte) and makes it thread-safe re-assignable.
type Func struct {
	mu sync.Mutex
	fn func([]byte)
}

// IsNil returns true when f.fn is nil.
func (f *Func) IsNil() (isNil bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.fn == nil {
		isNil = true
	}
	return
}

// Set assigns a func to f.fn (thread-safe).
func (f *Func) Set(fn func([]byte)) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.fn = fn
}

// Do calls f.fn (thread-safe).
func (f *Func) Do(b []byte) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.fn(b)
}
