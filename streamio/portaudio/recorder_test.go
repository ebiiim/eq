package portaudio_test

import (
	"encoding/binary"
	"testing"

	"github.com/ebiiim/eq/streamio/portaudio"
)

func TestNewRecorder(t *testing.T) {
	// NOTE: currently, we do not test this function because its results depends on sound devices
}

func initRecorders(t *testing.T) []*portaudio.Recorder {
	t.Helper()
	var ret []*portaudio.Recorder
	rs := []struct {
		bs, ch, b, r int
		bo           binary.ByteOrder
	}{
		{8192, 2, 16, 48000, binary.LittleEndian},
		{8192, 2, 16, 48000, binary.LittleEndian},
		{8192, 2, 16, 48000, binary.LittleEndian},
		{8192, 2, 16, 48000, binary.LittleEndian},
		{8192, 2, 16, 48000, binary.LittleEndian},
	}
	for i, v := range rs {
		r, err := portaudio.NewRecorder(-1, v.bs, v.ch, v.b, v.r, v.bo)
		if err != nil {
			t.Fatalf("could not init recorder #%d", i)
		}
		ret = append(ret, r)
	}
	return ret
}

func TestRecorder_Read(t *testing.T) {
	rs := initRecorders(t)
	cases := []struct {
		name   string
		r      *portaudio.Recorder
		bufLen int
	}{
		{"1B", rs[0], 1},
		{"1sample", rs[1], 2 * 2},
		{"8192B", rs[2], 8192},
		{"1s", rs[3], 1 * 48000 * 2 * 2},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			got := make([]byte, c.bufLen)
			for i := 0; i < 5; i++ { // test 5 times
				n, err := c.r.Read(got)
				if err != nil {
					t.Errorf("unexpected error %v", err)
				}
				if n != c.bufLen {
					t.Errorf("invalid read length idx %d got %d want %d data %v", i, n, c.bufLen, got)
				}
			}
		})
	}
}

func TestRecorder_Close(t *testing.T) {
	rs := initRecorders(t)
	cases := []struct {
		name  string
		r     *portaudio.Recorder
		fn    func(r *portaudio.Recorder)
		isErr bool
	}{
		{"init", rs[0], func(r *portaudio.Recorder) {}, false},
		{"after_read", rs[1], func(r *portaudio.Recorder) { r.Read(make([]byte, 8192)) }, false},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			c.fn(c.r)
			err := c.r.Close()
			if !((err != nil) == c.isErr) {
				t.Errorf("got %v, want %v(isErr) ", err, c.isErr)
			}
		})
	}
}
