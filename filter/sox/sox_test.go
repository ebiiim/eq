package sox_test

import (
	"testing"

	"github.com/ebiiim/eq/filter/sox"
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
			s := sox.NewGain(c.gain)
			if string(s) != c.want {
				t.Errorf("want %v got %v", c.want, s)
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
		{"normal", 1000, 5.0, -10.0, "equalizer 1000 5.000 -10.000"},
		{"int", 5000, 5, -10, "equalizer 5000 5.000 -10.000"},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			s := sox.NewEQ(c.freq, c.q, c.gain)
			if string(s) != c.want {
				t.Errorf("want %v got %v", c.want, s)
			}
		})
	}
}

func TestSoX_Cmd(t *testing.T) {
	cases := []struct {
		name string
		cmd  *sox.Command
		want string
	}{
		{"default", &sox.Command{}, "cmd -traw -b16 -r48000 -c2 -esigned -L - -traw -b16 -r48000 -c2 -esigned -L - --buffer 8192 -V0"},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			s := c.cmd.Get()
			if s != c.want {
				t.Errorf("want %v\ngot  %v", c.want, s)
			}
		})
	}
}
