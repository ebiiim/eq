package sox_test

import (
	"runtime"
	"testing"

	"github.com/ebiiim/eq/filter/pipe/sox"
)

func TestNewSoXGain(t *testing.T) {
	cases := []struct {
		name string
		gain float64
		want string
	}{
		{"-float", -10.0, "gain -10.000"},
		{"+float", 10.0, "gain 10.000"},
		{"0float", 0.0, "gain 0.000"},
		{"-int", -10, "gain -10.000"},
		{"+int", 10, "gain 10.000"},
		{"0int", 0, "gain 0.000"},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			got := sox.NewGain(c.gain)
			if string(got) != c.want {
				t.Errorf("got %v want %v", got, c.want)
			}
		})
	}
}

func TestNewSoXEQ(t *testing.T) {
	cases := []struct {
		name string
		freq uint
		q    float64
		gain float64
		want string
	}{
		{"normal", 1000, 5.0, -10.0, "equalizer 1000 5.000q -10.000"},
		{"int", 5000, 5, -10, "equalizer 5000 5.000q -10.000"},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			got := sox.NewEQ(c.freq, c.q, c.gain)
			if string(got) != c.want {
				t.Errorf("got %v want %v", got, c.want)
			}
		})
	}
}

func TestSoX_Cmd(t *testing.T) {
	cases := []struct {
		name string
		os   string
		cmd  *sox.Command
		want string
	}{
		{"darwin", "darwin", &sox.Command{}, "sox -traw -b16 -r48000 -c2 -esigned -L - -traw -b16 -r48000 -c2 -esigned -L - --buffer 8192 -V0"},
		{"linux", "linux", &sox.Command{}, "stdbuf -o16384 sox -traw -b16 -r48000 -c2 -esigned -L - -traw -b16 -r48000 -c2 -esigned -L - --buffer 8192 -V0"},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			got := c.cmd.String()
			if runtime.GOOS != c.os {
				return
			}
			if got != c.want {
				t.Errorf("got %v\nwant  %v", got, c.want)
			}
		})
	}
}
