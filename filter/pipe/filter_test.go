package pipe_test

import (
	"testing"

	"github.com/ebiiim/eq/filter/pipe"
	"github.com/ebiiim/eq/filter/pipe/sox"
	"github.com/google/go-cmp/cmp"
)

// TestFilter_Read and TestFilter_Write
func TestFilter(t *testing.T) {
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
			var f pipe.Filter
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

func TestFilter_Close(t *testing.T) {
	// NOTE: not implemented
}
