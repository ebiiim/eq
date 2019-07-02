package portaudio

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

type recorder struct {
	initOnce     sync.Once
	stream       *portaudio.Stream
	recordBuffer *[]int16
	byteOrder    binary.ByteOrder
	readerBuffer bytes.Buffer
}

func NewRecorder(bufferSize int, channels int, bitDepth int, sampleRate int, byteOrder binary.ByteOrder) (r *recorder, err error) {
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
	r = &recorder{stream: stream, recordBuffer: &recordBuffer, byteOrder: byteOrder}
	return r, nil
}

func (r *recorder) initialize() {
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

func (r *recorder) record() error {
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

func (r *recorder) Read(b []byte) (n int, err error) {
	r.initOnce.Do(r.initialize)

	readLen := len(b)
	for r.readerBuffer.Len() < readLen {
		//fmt.Print("r")
		time.Sleep(1 * time.Millisecond) // wait for record
	}
	//time.Sleep(100 * time.Millisecond)
	return r.readerBuffer.Read(b)
}

func (r *recorder) Close() error {
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