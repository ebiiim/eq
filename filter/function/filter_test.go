package function_test

import (
	"testing"

	"github.com/ebiiim/eq/filter/function"
	"github.com/google/go-cmp/cmp"
)

// TestFilter_Read and TestFilter_Write
func TestFilter(t *testing.T) {
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
		{"rot13", function.Rot13, 1, []byte("hello"), []byte("uryyb"), false},
		{"upper", function.ToUpper, 1, []byte("hello"), []byte("HELLO"), false},
		{"lower", function.ToLower, 1, []byte("HELLO"), []byte("hello"), false},
		{"vol0.5", func() func([]byte) {
			fn, _ := function.Volume(0.5)
			return fn
		}(), 4, []byte{10, 00, 20, 00, 30, 00, 40, 00}, []byte{05, 00, 10, 00, 15, 00, 20, 00}, false},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			var f function.Filter
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

func TestFilter_Close(t *testing.T) {
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
			fn, err := function.Volume(c.vol)
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
