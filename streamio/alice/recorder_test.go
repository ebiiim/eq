package alice_test

import (
	"testing"

	"github.com/ebiiim/eq/streamio/alice"
	"github.com/google/go-cmp/cmp"
)

func TestRecorder_Read(t *testing.T) {
	cases := []struct {
		name   string
		r      *alice.Recorder
		bufLen int
		want   [][]byte
	}{
		{"1B", &alice.Recorder{}, 1, [][]byte{[]byte("R"), []byte("e")}},
		{"8B", &alice.Recorder{}, 8, [][]byte{[]byte("Recorder"), []byte(" was beg")}},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			got := make([]byte, c.bufLen)
			for i, v := range c.want {
				_, err := c.r.Read(got)
				if err != nil {
					t.Errorf("unexpected error %v", err)
				}
				if !cmp.Equal(got, v) {
					t.Errorf("idx %v got %v want %v", i, got, v)
				}
			}
		})
	}
}

func TestRecorder_Close(t *testing.T) {
	cases := []struct {
		name  string
		r     *alice.Recorder
		fn    func(r *alice.Recorder)
		isErr bool
	}{
		{"init", &alice.Recorder{}, func(r *alice.Recorder) {}, false},
		{"after_read", &alice.Recorder{}, func(r *alice.Recorder) { r.Read(make([]byte, 4)) }, false},
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
