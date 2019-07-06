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

type Recorder struct {
	initOnce     sync.Once
	stream       *portaudio.Stream
	recordBuffer *[]int16
	byteOrder    binary.ByteOrder
	readerBuffer safe.Buffer
}

func NewRecorder(bufferSize int, channels int, bitDepth int, sampleRate int, byteOrder binary.ByteOrder) (r *Recorder, err error) {
	recordBuffer := make([]int16, bufferSize)
	// initialize Player
	err = portaudio.Initialize()
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize Player")
	}
	// open an output stream
	stream, err := portaudio.OpenDefaultStream(channels, 0, float64(sampleRate), bufferSize, recordBuffer)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open stream")
	}
	// start the stream
	err = stream.Start()
	if err != nil {
		return nil, errors.Wrap(err, "failed to start stream")
	}
	r = &Recorder{stream: stream, recordBuffer: &recordBuffer, byteOrder: byteOrder}
	return r, nil
}

func (r *Recorder) initialize() {
	go func() {
		for {
			gErr := r.record()
			if gErr != nil {
				// broken data detected
				// NOTE: this is not a bug (in most cases), so puts a log instead of an error
				fmt.Fprint(os.Stderr, "X")
			}
		}
	}()
}

func (r *Recorder) record() error {
	err := r.stream.Read() // read pcm data from the record stream to the buffer
	if err != nil {
		return errors.Wrap(err, "failed to read PCM")
	}
	err = binary.Write(&r.readerBuffer, r.byteOrder, r.recordBuffer) // convert []int16 -> []byte
	if err != nil {
		return errors.Wrap(err, "failed to write PCM")
	}
	return nil
}

func (r *Recorder) Read(b []byte) (n int, err error) {
	r.initOnce.Do(r.initialize)

	readLen := len(b)
	for r.readerBuffer.Len() < readLen {
		time.Sleep(1 * time.Millisecond) // wait for record
	}
	return r.readerBuffer.Read(b)
}

func (r *Recorder) Close() error {
	err := portaudio.Terminate()
	if err != nil {
		return errors.Wrap(err, "failed to terminate Player")
	}
	err = r.stream.Stop()
	if err != nil {
		return errors.Wrap(err, "failed to stop stream")
	}
	err = r.stream.Close()
	if err != nil {
		return errors.Wrap(err, "failed to close stream")
	}
	return nil
}
