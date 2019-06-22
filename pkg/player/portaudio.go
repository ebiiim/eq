package player

import (
	"encoding/binary"
	"io"

	"github.com/gordonklaus/portaudio"
	"github.com/pkg/errors"
)

type PortAudioPlayer struct {
	stream       *portaudio.Stream
	outputBuffer *[]int16
	byteOrder    binary.ByteOrder
}

func NewPortAudioPlayer(bufferSize int, channels int, bitDepth int, sampleRate int, byteOrder binary.ByteOrder) (p *PortAudioPlayer, err error) {
	outputBuffer := make([]int16, bufferSize)

	// initialize PortAudio
	err = portaudio.Initialize()
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize PortAudio")
	}

	// open an input stream
	stream, err := portaudio.OpenDefaultStream(0, channels, float64(sampleRate), bufferSize, outputBuffer)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open stream")
	}

	// start the stream
	err = stream.Start()
	if err != nil {
		return nil, errors.Wrap(err, "failed to start stream")
	}

	p = &PortAudioPlayer{stream: stream, outputBuffer: &outputBuffer, byteOrder: byteOrder}
	return p, nil
}

func (player *PortAudioPlayer) Play(r io.Reader) error {
	// copy pcm data from the reader to the buffer
	err := binary.Read(r, player.byteOrder, player.outputBuffer)
	if err != nil {
		return errors.Wrap(err, "failed to read PCM")
	}
	err = player.stream.Write()
	if err != nil {
		return errors.Wrap(err, "failed to play")
	}
	return nil
}

func (player *PortAudioPlayer) Close() error {
	err := portaudio.Terminate()
	if err != nil {
		return errors.Wrap(err, "failed to terminate PortAudio")
	}
	err = player.stream.Stop()
	if err != nil {
		return errors.Wrap(err, "failed to stop stream")
	}
	err = player.stream.Close()
	if err != nil {
		return errors.Wrap(err, "failed to close stream")
	}
	return nil
}
