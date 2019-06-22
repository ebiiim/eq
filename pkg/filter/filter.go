package filter

import (
	"encoding/binary"
	"io"
)

type Filter interface {
	io.Reader
}

type VolumeFilter struct {
	volume float64
	Reader io.Reader
}

func NewVolumeFilter(reader io.Reader, volume float64) (*VolumeFilter, error) {
	return &VolumeFilter{volume: volume, Reader: reader}, nil
}

func (f *VolumeFilter) Read(p []byte) (n int, err error) {
	n, err = f.Reader.Read(p)
	for i := 0; i < n-1; i += 2 {
		slice := p[i : i+2]
		sample := binary.LittleEndian.Uint16(slice)
		bs := make([]byte, 2)
		binary.LittleEndian.PutUint16(bs, uint16(float64(int16(sample))*f.volume))
		p[i], p[i+1] = bs[0], bs[1]
	}
	return
}
