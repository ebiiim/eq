package alice

import (
	"io"
	"sync"
	"time"

	"github.com/ebiiim/eq/internal/safe"
	"github.com/pkg/errors"
)

type Player struct {
	initOnce     sync.Once
	writer       io.Writer
	writerBuffer safe.Buffer
	bufLen       int
}

func NewPlayer(writer io.Writer, bufferSize int) (p *Player, err error) {
	p = &Player{writer: writer, bufLen: bufferSize}
	return p, nil
}

func (p *Player) initialize() {
	go func() {
		for {
			gErr := p.play()
			if gErr != nil {
				panic(gErr) // this error should not occur
			}
		}
	}()
}

func (p *Player) play() error {
	for p.writerBuffer.Len() < p.bufLen {
		time.Sleep(100 * time.Millisecond) // wait for record
	}
	buf := make([]byte, p.bufLen)
	_, err := p.writerBuffer.Read(buf)
	if err != nil {
		return errors.Wrap(err, "failed to get data from writerBuffer")
	}
	_, err = p.writer.Write(buf)
	if err != nil {
		return errors.Wrap(err, "failed to write")
	}
	return nil
}

func (p *Player) Write(b []byte) (n int, err error) {
	p.initOnce.Do(p.initialize)
	return p.writerBuffer.Write(b)
}

func (p *Player) Close() error {
	return nil
}
