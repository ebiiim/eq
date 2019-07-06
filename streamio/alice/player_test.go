package alice_test

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/ebiiim/eq/streamio/alice"
)

func TestNewPlayer(t *testing.T) {
	cases := []struct {
		name    string
		writer  io.Writer
		bufSize int
		isErr   bool
	}{
		{"8B", os.Stdout, 8, false},
		{"8KB", os.Stdout, 8192, false},
		{"stderr", os.Stderr, 8, false},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			got, err := alice.NewPlayer(c.writer, c.bufSize)
			if !((err != nil) == c.isErr) {
				t.Errorf("got %v, want %v(isErr) ", err, c.isErr)
			}
			_, err = got.Write(make([]byte, 2))
			if !((err != nil) == c.isErr) {
				t.Errorf("got %v, want %v(isErr) ", err, c.isErr)
			}
		})
	}

}

func initPlayers(t *testing.T) []*alice.Player {
	t.Helper()
	var ret []*alice.Player
	ps := []struct {
		w  io.Writer
		bs int
	}{
		{bytes.NewBuffer([]byte{}), 8},
		{bytes.NewBuffer([]byte{}), 8},
		{bytes.NewBuffer([]byte{}), 8},
		{bytes.NewBuffer([]byte{}), 8},
		{bytes.NewBuffer([]byte{}), 8},
	}
	for i, v := range ps {
		p, err := alice.NewPlayer(v.w, v.bs)
		if err != nil {
			t.Fatalf("could not init player #%d", i)
		}
		ret = append(ret, p)
	}
	return ret
}

func TestPlayer_Write(t *testing.T) {
	ps := initPlayers(t)
	cases := []struct {
		name   string
		p      *alice.Player
		bufLen int
	}{
		{"1B", ps[0], 1},
		{"8B", ps[1], 8},
		{"512B", ps[2], 512},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			got := make([]byte, c.bufLen)
			for i := 0; i < 5; i++ { // test 5 times
				n, err := c.p.Write(got)
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

func TestPlayer_Close(t *testing.T) {
	ps := initPlayers(t)
	cases := []struct {
		name  string
		p     *alice.Player
		fn    func(r *alice.Player)
		isErr bool
	}{
		{"init", ps[0], func(r *alice.Player) {}, false},
		{"after_write", ps[1], func(r *alice.Player) { r.Write(make([]byte, 8)) }, false},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			c.fn(c.p)
			err := c.p.Close()
			if !((err != nil) == c.isErr) {
				t.Errorf("got %v, want %v(isErr) ", err, c.isErr)
			}
		})
	}
}
