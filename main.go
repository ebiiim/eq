package main

import (
	"encoding/binary"

	"github.com/ebiiim/eq/pkg/filter"
	"github.com/ebiiim/eq/pkg/player"
	"github.com/ebiiim/eq/pkg/recorder"
)

func main() {
	var (
		bs   = 4096
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

	var f filter.VolumeFilter

	defer p.Close()
	defer r.Close()

	for {
		reader, err := r.Record()
		if err != nil {
			panic(err)
		}

		reader, err = f.Filter(reader)
		if err != nil {
			panic(err)
		}

		err = p.Play(reader)
		if err != nil {
			panic(err)
		}
	}
}
