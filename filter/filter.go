package filter

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
	"unicode"

	"github.com/pkg/errors"
)

type Filter interface {
	io.ReadWriteCloser
}

type CmdFilter struct {
	cmd       *exec.Cmd
	initOnce  sync.Once
	inPipe    io.WriteCloser
	outPipe   io.ReadCloser
	FilterCmd string
}

func (f *CmdFilter) initialize() (err error) {
	ss := strings.Split(f.FilterCmd, " ")
	f.cmd = exec.Command(ss[0], ss[1:]...)

	f.inPipe, err = f.cmd.StdinPipe()
	if err != nil {
		return errors.Wrap(err, "could not open stdin pipe")
	}
	f.outPipe, err = f.cmd.StdoutPipe()
	if err != nil {
		return errors.Wrap(err, "could not open stdout pipe")
	}
	err = f.cmd.Start()
	if err != nil {
		return errors.Wrap(err, "could not start exec")
	}
	return nil
}

func (f *CmdFilter) Read(b []byte) (n int, err error) {
	f.initOnce.Do(func() { err = f.initialize() })
	if err != nil {
		return 0, err
	}
	n, err = f.outPipe.Read(b)
	return
}

func (f *CmdFilter) Write(b []byte) (n int, err error) {
	f.initOnce.Do(func() { err = f.initialize() })
	if err != nil {
		return 0, err
	}
	n, err = f.inPipe.Write(b)
	return
}

func (f *CmdFilter) Close() (err error) {
	err = f.inPipe.Close()
	if err != nil {
		return errors.Wrap(err, "could not close stdin pipe")
	}
	// TODO: wait until outPipe is empty
	err = f.outPipe.Close()
	if err != nil {
		return errors.Wrap(err, "could not close stdout pipe")
	}
	err = f.cmd.Wait()
	if err != nil {
		return errors.Wrap(err, "could not wait exec")
	}
	if f.cmd.ProcessState.ExitCode() != 0 {
		return fmt.Errorf("abnormal exit code %d", f.cmd.ProcessState.ExitCode())
	}
	return nil
}

type FuncFilter struct {
	initOnce   sync.Once
	inBuf      bytes.Buffer
	outBuf     bytes.Buffer
	inCh       chan []byte
	outCh      chan []byte
	bufferSize int
	ChunkSize  int
	FilterFunc func([]byte)
}

func (f *FuncFilter) initialize() {
	f.ChunkSize = 8
	f.bufferSize = 65536 / f.ChunkSize // 64kB (max.)
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

func (f *FuncFilter) Read(b []byte) (n int, err error) {
	f.initOnce.Do(f.initialize)

	readLen := len(b)
	for f.outBuf.Len() < readLen {
		f.outBuf.Write(<-f.outCh)
	}
	return f.outBuf.Read(b)
}

func (f *FuncFilter) Write(b []byte) (n int, err error) {
	f.initOnce.Do(f.initialize)

	f.inBuf.Write(b)
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

func (f *FuncFilter) Close() error {
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

var Vol50 = func(b []byte) {
	newVolume := 0.5
	for i := 0; i < len(b)-1; i += 2 {
		sample := binary.LittleEndian.Uint16(b[i : i+2])
		bs := make([]byte, 2)
		binary.LittleEndian.PutUint16(bs, uint16(float64(int16(sample))*newVolume))
		b[i], b[i+1] = bs[0], bs[1]
	}
}
