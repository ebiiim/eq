package portaudio_test

import (
	"encoding/binary"
	"testing"

	"github.com/ebiiim/eq/streamio/portaudio"
)

func TestNewPlayer(t *testing.T) {
	// NOTE: currently, we do not test this function because the result depends on sound devices
}

func initPlayers(t *testing.T) []*portaudio.Player {
	t.Helper()
	var ret []*portaudio.Player
	ps := []struct {
		bs, ch, b, r int
		bo           binary.ByteOrder
	}{
		{8192, 2, 16, 48000, binary.LittleEndian},
		{8192, 2, 16, 48000, binary.LittleEndian},
		{8192, 2, 16, 48000, binary.LittleEndian},
		{8192, 2, 16, 48000, binary.LittleEndian},
		{8192, 2, 16, 48000, binary.LittleEndian},
	}
	for i, v := range ps {
		p, err := portaudio.NewPlayer(-1, v.bs, v.ch, v.b, v.r, v.bo)
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
		p      *portaudio.Player
		bufLen int
	}{
		{"1B", ps[0], 1},
		{"1sample", ps[1], 2 * 2},
		{"8192B", ps[2], 8192},
		{"1s", ps[3], 1 * 48000 * 2 * 2},
		{"3s", ps[4], 3 * 48000 * 2 * 2},
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
		p     *portaudio.Player
		fn    func(r *portaudio.Player)
		isErr bool
	}{
		{"init", ps[0], func(r *portaudio.Player) {}, false},
		{"after_write", ps[1], func(r *portaudio.Player) { r.Write(make([]byte, 8192)) }, false},
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
