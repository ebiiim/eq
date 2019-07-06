package function

import (
	"encoding/binary"
	"errors"
	"sync"
	"unicode"

	"github.com/ebiiim/eq/internal/safe"
)

type Filter struct {
	initOnce   sync.Once
	inBuf      safe.Buffer
	outBuf     safe.Buffer
	inCh       chan []byte
	outCh      chan []byte
	bufferSize int
	ChunkSize  int
	FilterFunc func([]byte)
}

func (f *Filter) initialize() {
	if f.ChunkSize == 0 {
		f.ChunkSize = 8
	}
	if f.FilterFunc == nil {
		f.FilterFunc = func([]byte) {}
	}
	f.bufferSize = 65536 / f.ChunkSize // 64KB (max.)
	f.inCh = make(chan []byte, f.bufferSize)
	f.outCh = make(chan []byte, f.bufferSize)

	go func() {
		for {
			b := <-f.inCh
			f.FilterFunc(b)
			f.outCh <- b
		}
	}()
}

func (f *Filter) Read(b []byte) (n int, err error) {
	f.initOnce.Do(f.initialize)

	readLen := len(b)
	for f.outBuf.Len() < readLen {
		_, err := f.outBuf.Write(<-f.outCh)
		if err != nil {
			return 0, err
		}
	}
	return f.outBuf.Read(b)
}

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

func (f *Filter) Close() error {
	close(f.outCh)
	close(f.inCh)
	return nil
}

var Rot13 = func(b []byte) {
	for i, c := range b {
		if c <= 'Z' && c >= 'A' {
			b[i] = (c-'A'+13)%26 + 'A'
		} else if c >= 'a' && c <= 'z' {
			b[i] = (c-'a'+13)%26 + 'a'
		}
	}
}

var ToUpper = func(b []byte) {
	for i := range b {
		b[i] = byte(unicode.ToUpper(rune(b[i])))
	}
}

var ToLower = func(b []byte) {
	for i := range b {
		b[i] = byte(unicode.ToLower(rune(b[i])))
	}
}

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
