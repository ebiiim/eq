package portaudio

import (
	"encoding/binary"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/ebiiim/eq/internal/safe"
	"github.com/gordonklaus/portaudio"
	"github.com/pkg/errors"
)

type Player struct {
	initOnce     sync.Once
	stream       *portaudio.Stream
	playBuffer   *[]int16
	byteOrder    binary.ByteOrder
	writerBuffer safe.Buffer
}

func NewPlayer(outputDeviceID int, bufferSize int, channels int, bitDepth int, sampleRate int, byteOrder binary.ByteOrder) (p *Player, err error) {
	playBuffer := make([]int16, bufferSize)
	// initialize Player
	err = portaudio.Initialize()
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize Player")
	}
	// open an input stream
	stream, err := OpenStream(-1, outputDeviceID, 0, channels, float64(sampleRate), bufferSize, playBuffer)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open stream")
	}
	// start the stream
	err = stream.Start()
	if err != nil {
		return nil, errors.Wrap(err, "failed to start stream")
	}
	p = &Player{stream: stream, playBuffer: &playBuffer, byteOrder: byteOrder}
	return p, nil
}

func (p *Player) initialize() {
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

func (p *Player) play() error {
	for p.writerBuffer.Len()*2 < len(*p.playBuffer) {
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

func (p *Player) Write(b []byte) (n int, err error) {
	p.initOnce.Do(p.initialize)
	return p.writerBuffer.Write(b)
}

func (p *Player) Close() error {
	err := portaudio.Terminate()
	if err != nil {
		return errors.Wrap(err, "failed to terminate Player")
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
