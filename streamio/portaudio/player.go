package portaudio

import (
	"encoding/binary"
	"fmt"
	"os"
	"time"

	"github.com/JulianKnodt/portaudio"
	"github.com/ebiiim/eq/internal/safe"
	"github.com/pkg/errors"
)

// Player is a writable PortAudio output device.
type Player struct {
	stream       *portaudio.Stream
	playBuffer   *[]int16
	byteOrder    binary.ByteOrder
	writerBuffer safe.Buffer
}

// NewPlayer initialize a Player object.
//
// This function invokes a goroutine
// that sequentially reads data from the playback buffer
// and writes the data to the audio output device.
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
	p.initialize()
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
	for p.writerBuffer.Len() < len(*p.playBuffer) {
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

// Write writes len(b) bytes from b to the playback buffer.
func (p *Player) Write(b []byte) (n int, err error) {
	return p.writerBuffer.Write(b)
}

// Close terminates PortAudio.
func (p *Player) Close() (err error) {
	err = p.stream.Stop()
	if err != nil {
		return errors.Wrap(err, "failed to stop stream")
	}
	err = p.stream.Close()
	if err != nil {
		return errors.Wrap(err, "failed to close stream")
	}
	err = portaudio.Terminate()
	if err != nil {
		return errors.Wrap(err, "failed to terminate Player")
	}
	return nil
}
