package filter_test

import (
	"testing"

	"github.com/ebiiim/eq/pkg/filter"
)

func TestNewSoXGain(t *testing.T) {
	const (
		effect = "gain"
	)
	cases := []struct {
		name string
		gain float64
		want []string
	}{
		{"-float", -10.0, filter.SoXEffect{effect, "-10.000"}},
		{"+float", 10.0, filter.SoXEffect{effect, "10.000"}},
		{"0float", 0.0, filter.SoXEffect{effect, "0.000"}},
		{"-int", -10, filter.SoXEffect{effect, "-10.000"}},
		{"+int", 10, filter.SoXEffect{effect, "10.000"}},
		{"0int", 0, filter.SoXEffect{effect, "0.000"}},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			se := filter.NewSoXGain(c.gain)
			for i, v := range se {
				if v != c.want[i] {
					t.Errorf("idx %v want %v got %v", i, c.want[i], v)
				}
			}
		})
	}
}

func TestNewSoXEqualizer(t *testing.T) {
	const (
		effect = "equalizer"
	)
	cases := []struct {
		name string
		q    float64
		gain float64
		want []string
	}{
		{"normal", 5.0, -10.0, filter.SoXEffect{effect, "5.000", "-10.000"}},
		{"int", 5, -10, filter.SoXEffect{effect, "5.000", "-10.000"}},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			se := filter.NewSoXEqualizer(c.q, c.gain)
			for i, v := range se {
				if v != c.want[i] {
					t.Errorf("idx %v want %v got %v", i, c.want[i], v)
				}
			}
		})
	}
}
