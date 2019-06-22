package main

import (
	"encoding/binary"

	"github.com/ebiiim/eq/pkg/filter"
	"github.com/ebiiim/eq/pkg/player"
	"github.com/ebiiim/eq/pkg/recorder"
)

func main() {
	var (
		bs   = 8192
		ch   = 2
		bit  = 16
		rate = 48000
		bo   = binary.LittleEndian
	)

	var r recorder.Recorder
	r, err := recorder.NewPortAudioRecorder(bs, ch, bit, rate, bo)
	if err != nil {
		panic(err)
	}

	var p player.Player
	p, err = player.NewPortAudioPlayer(bs, ch, bit, rate, bo)
	if err != nil {
		panic(err)
	}

	defer p.Close()
	defer r.Close()

	for {
		reader, err := r.Record()
		if err != nil {
			panic(err)
		}
		reader, _ = filter.NewVolumeFilter(reader, 0.5)
		err = p.Play(reader)
		if err != nil {
			panic(err)
		}
	}
}
