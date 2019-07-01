package player

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/gordonklaus/portaudio"
	"github.com/pkg/errors"
)

type PortAudio struct {
	initOnce     sync.Once
	stream       *portaudio.Stream
	playBuffer   *[]int16
	byteOrder    binary.ByteOrder
	writerBuffer bytes.Buffer
}

func NewPortAudio(bufferSize int, channels int, bitDepth int, sampleRate int, byteOrder binary.ByteOrder) (p *PortAudio, err error) {
	playBuffer := make([]int16, bufferSize)
	// initialize PortAudio
	err = portaudio.Initialize()
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize PortAudio")
	}
	// open an input stream
	stream, err := portaudio.OpenDefaultStream(0, channels, float64(sampleRate), bufferSize, playBuffer)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open stream")
	}
	// start the stream
	err = stream.Start()
	if err != nil {
		return nil, errors.Wrap(err, "failed to start stream")
	}
	p = &PortAudio{stream: stream, playBuffer: &playBuffer, byteOrder: byteOrder}
	return p, nil
}

func (p *PortAudio) initialize() {
	go func() {
		for {
			gErr := p.play()
			if gErr != nil {
				// broken data detected
				// NOTE: this is not a bug (in most cases), so puts a log instead of an error
				fmt.Fprint(os.Stderr, "x")
			}
		}
	}()
}

func (p *PortAudio) play() error {
	for p.writerBuffer.Len()*2 <= len(*p.playBuffer) {
		time.Sleep(1 * time.Millisecond) // wait for record
	}
	err := binary.Read(&p.writerBuffer, p.byteOrder, p.playBuffer) // convert []int16 -> []byte
	if err != nil {
		return errors.Wrap(err, "failed to read PCM")
	}
	err = p.stream.Write() // write pcm data from the buffer to the play stream
	if err != nil {
		return errors.Wrap(err, "failed to play")
	}
	return nil
}

func (p *PortAudio) Write(b []byte) (n int, err error) {
	p.initOnce.Do(p.initialize)
	return p.writerBuffer.Write(b)
}

func (p *PortAudio) Close() error {
	err := portaudio.Terminate()
	if err != nil {
		return errors.Wrap(err, "failed to terminate PortAudio")
	}
	err = p.stream.Stop()
	if err != nil {
		return errors.Wrap(err, "failed to stop stream")
	}
	err = p.stream.Close()
	if err != nil {
		return errors.Wrap(err, "failed to close stream")
	}
	return nil
}
