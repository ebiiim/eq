// Package portaudio provides an implementation of streamio.Player and streamio.Recorder using PortAudio.
package portaudio

import (
	"fmt"

	"github.com/gordonklaus/portaudio"
	"github.com/pkg/errors"
)

// ListDevices returns a slice of string containing device info on each line.
func ListDevices() ([]string, error) {
	err := portaudio.Initialize()
	if err != nil {
		return nil, err
	}
	defer func() {
		tErr := portaudio.Terminate()
		if tErr != nil {
			err = errors.Wrapf(err, "(%v)", tErr)
		}
	}()

	ds, err := portaudio.Devices()
	if err != nil {
		return nil, err
	}
	var ss []string
	for i, v := range ds {
		s := fmt.Sprintf("ID: %d, Type: %s, Name: %s, InputCh: %d, OutputCh: %d", i, v.HostApi.Name, v.Name, v.MaxInputChannels, v.MaxOutputChannels)
		ss = append(ss, s)
	}
	return ss, nil
}

// OpenStream opens a stream with device IDs that portaudio.OpenDefaultStream does not support.
//
// If the device ID is -1, the function uses the default input/output device.
func OpenStream(inputDeviceID, outputDeviceID int, numInputChannels, numOutputChannels int, sampleRate float64, framesPerBuffer int, args ...interface{}) (*portaudio.Stream, error) {
	var in, out *portaudio.DeviceInfo
	var err error

	if numInputChannels > 0 {
		if inputDeviceID < 0 {
			in, err = portaudio.DefaultInputDevice()
			if err != nil {
				return nil, err
			}
		} else {
			ds, err := portaudio.Devices()
			if err != nil {
				return nil, err
			}
			if inputDeviceID >= len(ds) {
				return nil, errors.New("invalid inputDeviceID")
			}
			in = ds[inputDeviceID]
		}
	}

	if numOutputChannels > 0 {
		if outputDeviceID < 0 {
			out, err = portaudio.DefaultOutputDevice()
			if err != nil {
				return nil, err
			}
		} else {
			ds, err := portaudio.Devices()
			if err != nil {
				return nil, err
			}
			if outputDeviceID >= len(ds) {
				return nil, errors.New("invalid outputDeviceID")
			}
			out = ds[outputDeviceID]
		}
	}
	p := portaudio.HighLatencyParameters(in, out)
	p.Input.Channels = numInputChannels
	p.Output.Channels = numOutputChannels
	p.SampleRate = sampleRate
	p.FramesPerBuffer = framesPerBuffer
	return portaudio.OpenStream(p, args...)
}
