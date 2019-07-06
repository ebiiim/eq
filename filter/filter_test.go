package filter_test

import (
	"testing"

	"github.com/ebiiim/eq/filter"
	"github.com/ebiiim/eq/filter/sox"
	"github.com/google/go-cmp/cmp"
)

// TestFunc_Read and TestFunc_Write
func TestFunc(t *testing.T) {
	cases := []struct {
		name      string
		fn        func([]byte)
		chunkSize int
		in        []byte
		want      []byte
		isErr     bool
	}{
		{"no_change_bs1", func(b []byte) {}, 1, []byte{11, 22, 33, 44}, []byte{11, 22, 33, 44}, false},
		{"no_change_bs4", func(b []byte) {}, 4, []byte{11, 22, 33, 44}, []byte{11, 22, 33, 44}, false},
		{"rot13", filter.Rot13, 1, []byte("hello"), []byte("uryyb"), false},
		{"upper", filter.ToUpper, 1, []byte("hello"), []byte("HELLO"), false},
		{"lower", filter.ToLower, 1, []byte("HELLO"), []byte("hello"), false},
		{"vol0.5", func() func([]byte) {
			fn, _ := filter.Volume(0.5)
			return fn
		}(), 4, []byte{10, 00, 20, 00, 30, 00, 40, 00}, []byte{05, 00, 10, 00, 15, 00, 20, 00}, false},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			var f filter.Func
			f.ChunkSize = c.chunkSize
			f.FilterFunc = c.fn
			_, err := f.Write(c.in)
			if !((err != nil) == c.isErr) {
				t.Errorf("got %v, want %v(isErr) ", err, c.isErr)
			}
			_, err = f.Read(c.in)
			if !((err != nil) == c.isErr) {
				t.Errorf("got %v, want %v(isErr) ", err, c.isErr)
			}
			if c.isErr {
				return
			}
			if !cmp.Equal(c.in, c.want) {
				t.Errorf("got %v want %v", c.in, c.want)
			}
		})
	}
}

func TestFunc_Close(t *testing.T) {
	// NOTE: not implemented
}

// TestPipe_Read and TestPipe_Write
func TestPipe(t *testing.T) {
	cases := []struct {
		name  string
		cmd   string
		in    []byte
		want  []byte
		isErr bool
	}{
		{"no_change", "tee -i /dev/null", []byte{11, 22, 33, 44}, []byte{11, 22, 33, 44}, false},
		{"sox", (&sox.Command{}).String(), make([]byte, 8192*2), make([]byte, 8192*2), false},
		{"sox_max_bs", (&sox.Command{}).String(), make([]byte, 8192*4), make([]byte, 8192*4), false},
		{"sox_3to2", (&sox.Command{}).String(), make([]byte, 8192*3), make([]byte, 8192*3), false},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			var f filter.Pipe
			f.FilterCmd = c.cmd
			_, err := f.Write(c.in)
			if !((err != nil) == c.isErr) {
				t.Errorf("got %v, want %v(isErr) ", err, c.isErr)
			}
			_, err = f.Read(c.in)
			if !((err != nil) == c.isErr) {
				t.Errorf("got %v, want %v(isErr) ", err, c.isErr)
			}
			if c.isErr {
				return
			}
			if !cmp.Equal(c.in, c.want) {
				t.Errorf("got %v want %v", c.in, c.want)
			}
		})
	}
}

func TestPipe_Close(t *testing.T) {
	// NOTE: not implemented
}

func TestVolume(t *testing.T) {
	cases := []struct {
		name  string
		vol   float64
		in    []byte // uint16 little endian
		want  []byte // uint16 little endian
		isErr bool
	}{
		{"no_change_2B", 1.0, []byte{10, 00}, []byte{10, 00}, false},
		{"no_change_8B", 1.0, []byte{10, 00, 20, 00, 30, 00, 40, 00}, []byte{10, 00, 20, 00, 30, 00, 40, 00}, false},
		{"0.5", 0.5, []byte{10, 00, 20, 00, 30, 00, 40, 00}, []byte{05, 00, 10, 00, 15, 00, 20, 00}, false},
		{"mute", 0, []byte{10, 00, 20, 00, 30, 00, 40, 00}, []byte{00, 00, 00, 00, 00, 00, 00, 00}, false},
		{"F_negative", -0.1, []byte{10, 00, 20, 00, 30, 00, 40, 00}, nil, true},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			fn, err := filter.Volume(c.vol)
			if !((err != nil) == c.isErr) {
				t.Errorf("got %v, want %v(isErr) ", err, c.isErr)
			}
			if c.isErr {
				return
			}
			fn(c.in)
			if !cmp.Equal(c.in, c.want) {
				t.Errorf("got %v want %v", c.in, c.want)
			}
		})
	}
}
