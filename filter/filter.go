package filter

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"sync"
)

type Filter interface {
	Filter(reader io.Reader) (io.Reader, error)
}

type VolumeFilter struct {
	initOnce sync.Once
	mu       sync.Mutex
	volume   float64
}

func (f *VolumeFilter) SetVolume(volume float64) error {
	if volume < 0.0 {
		return fmt.Errorf("invalid volume %v", volume)
	}
	f.mu.Lock()
	f.volume = volume
	f.mu.Unlock()
	return nil
}

func (f *VolumeFilter) Filter(reader io.Reader) (io.Reader, error) {
	f.initOnce.Do(func() { _ = f.SetVolume(1.0) }) // safe

	f.mu.Lock()
	newVolume := f.volume
	f.mu.Unlock()

	buf, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(buf)-1; i += 2 {
		sample := binary.LittleEndian.Uint16(buf[i : i+2])
		bs := make([]byte, 2)
		binary.LittleEndian.PutUint16(bs, uint16(float64(int16(sample))*newVolume))
		buf[i], buf[i+1] = bs[0], bs[1]
	}
	return bytes.NewReader(buf), nil
}
