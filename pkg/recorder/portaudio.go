package recorder

import (
	"bytes"
	"encoding/binary"
	"io"

	"github.com/gordonklaus/portaudio"
	"github.com/pkg/errors"
)

type PortAudioRecorder struct {
	stream      *portaudio.Stream
	inputBuffer *[]int16
	byteOrder   binary.ByteOrder
}

func NewPortAudioRecorder(bufferSize int, channels int, bitDepth int, sampleRate int, byteOrder binary.ByteOrder) (r *PortAudioRecorder, err error) {
	inputBuffer := make([]int16, bufferSize)

	// initialize PortAudio
	err = portaudio.Initialize()
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize PortAudio")
	}

	// open an output stream
	stream, err := portaudio.OpenDefaultStream(channels, 0, float64(sampleRate), bufferSize, inputBuffer)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open stream")
	}

	// start the stream
	err = stream.Start()
	if err != nil {
		return nil, errors.Wrap(err, "failed to start stream")
	}

	r = &PortAudioRecorder{stream: stream, inputBuffer: &inputBuffer, byteOrder: byteOrder}
	return r, nil
}

func (recorder *PortAudioRecorder) Record() (io.Reader, error) {
	err := recorder.stream.Read() // read pcm data from the record stream to the buffer
	if err != nil {
		return nil, errors.Wrap(err, "failed to read PCM")
	}

	var buf bytes.Buffer
	err = binary.Write(&buf, recorder.byteOrder, recorder.inputBuffer) // convert []int16 -> []byte
	if err != nil {
		return nil, errors.Wrap(err, "failed to write PCM")
	}
	return &buf, nil
}

func (recorder *PortAudioRecorder) Close() error {
	err := portaudio.Terminate()
	if err != nil {
		return errors.Wrap(err, "failed to terminate PortAudio")
	}
	err = recorder.stream.Stop()
	if err != nil {
		return errors.Wrap(err, "failed to stop stream")
	}
	err = recorder.stream.Close()
	if err != nil {
		return errors.Wrap(err, "failed to close stream")
	}
	return nil
}
