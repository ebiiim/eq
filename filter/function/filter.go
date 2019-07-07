// Package function provides an implementation of filter.Filter
// that using functions for filtering.
package function

import (
	"encoding/binary"
	"errors"
	"sync"
	"unicode"

	"github.com/ebiiim/eq/internal/safe"
)

// Filter is a stream data processor using a function.
type Filter struct {
	initOnce   sync.Once
	inBuf      safe.Buffer
	outBuf     safe.Buffer
	inCh       chan []byte
	outCh      chan []byte
	bufferSize int

	// Func is a function that reads len(b) bytes from b
	// and does some processing on them.
	//
	// e.g. function.Rot13
	Func func([]byte)

	// ChunkSize determines the number of bytes to pass to Func once (default: 32).
	//
	// Change this value if the specified Func cares about the data size.
	// Func is called only when the size of data in the input buffer
	// (Write puts data into the input buffer) is ChunkSize or more.
	ChunkSize int
}

func (f *Filter) initialize() {
	if f.ChunkSize == 0 {
		f.ChunkSize = 32
	}
	if f.Func == nil {
		f.Func = func([]byte) {}
	}
	f.bufferSize = 65536 / f.ChunkSize // 64KB (max.)
	f.inCh = make(chan []byte, f.bufferSize)
	f.outCh = make(chan []byte, f.bufferSize)

	go func() {
		for {
			b := <-f.inCh
			f.Func(b)
			f.outCh <- b
		}
	}()
}

// Read reads len(b) bytes from
// the output buffer (contains processed data) into b.
//
// The function blocks until it reads len(b) bytes or more.
// The function does not support ioutil.ReadAll (blocks permanently).
func (f *Filter) Read(b []byte) (n int, err error) {
	readLen := len(b)
	for f.outBuf.Len() < readLen {
		_, err := f.outBuf.Write(<-f.outCh)
		if err != nil {
			return 0, err
		}
	}
	return f.outBuf.Read(b)
}

// Write writes len(b) bytes from b to the input buffer
// that contains pre-processed data.
//
// The first call to this function invokes a goroutine
// that sequentially reads data from the input buffer,
// process data using Func, and writes the processed data into the output buffer.
func (f *Filter) Write(b []byte) (int, error) {
	f.initOnce.Do(f.initialize)

	_, err := f.inBuf.Write(b)
	if err != nil {
		return 0, err
	}
	for f.inBuf.Len() >= f.ChunkSize {
		bb := make([]byte, f.ChunkSize)
		_, err = f.inBuf.Read(bb)
		if err != nil {
			return 0, err
		}
		f.inCh <- bb
	}
	return len(b), nil
}

// Close closes the Filter object.
func (f *Filter) Close() error {
	close(f.outCh)
	close(f.inCh)
	return nil
}

// Rot13 reads len(b) bytes from b and shifts them
// with Caesar Cipher (assuming that all bytes in b are ASCII characters).
var Rot13 = func(b []byte) {
	for i, c := range b {
		if c <= 'Z' && c >= 'A' {
			b[i] = (c-'A'+13)%26 + 'A'
		} else if c >= 'a' && c <= 'z' {
			b[i] = (c-'a'+13)%26 + 'a'
		}
	}
}

// ToUpper reads len(b) bytes from b and converts them
// to upper case (assuming that all bytes in b are ASCII characters).
var ToUpper = func(b []byte) {
	for i := range b {
		b[i] = byte(unicode.ToUpper(rune(b[i])))
	}
}

// ToLower reads len(b) bytes from b and converts them
// to lower case (assuming that all bytes in b are ASCII characters).
var ToLower = func(b []byte) {
	for i := range b {
		b[i] = byte(unicode.ToLower(rune(b[i])))
	}
}

// Volume returns a function that changes the volume of audio data from b.
//
// The returned function reads len(b) bytes from b
// assuming that the data is an Uint16 stream
// and multiply each sample by 'volume'.
func Volume(volume float64) (func(b []byte), error) {
	if volume < 0 {
		return nil, errors.New("volume must be >0")
	}
	fn := func(b []byte) {
		for i := 0; i < len(b)-1; i += 2 {
			sample := binary.LittleEndian.Uint16(b[i : i+2])
			bs := make([]byte, 2)
			binary.LittleEndian.PutUint16(bs, uint16(float64(int16(sample))*volume))
			b[i], b[i+1] = bs[0], bs[1]
		}
	}
	return fn, nil
}
