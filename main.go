package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"os"

	"github.com/ebiiim/eq/pkg/filter"
	"github.com/ebiiim/eq/pkg/player"
	"github.com/ebiiim/eq/pkg/recorder"
	term "github.com/nsf/termbox-go"
)

type CLIPlayer struct {
	r      recorder.Recorder
	p      player.Player
	vf     *filter.VolumeFilter
	sf     *filter.SoXFilter
	volume float64
	isMute bool
}

var cliPlayer CLIPlayer

func initialize() {
	var (
		bs   = 4096
		ch   = 2
		bit  = 16
		rate = 48000
		bo   = binary.LittleEndian
	)

	r, err := recorder.NewPortAudioRecorder(bs, ch, bit, rate, bo)
	if err != nil {
		panic(err)
	}

	p, err := player.NewPortAudioPlayer(bs, ch, bit, rate, bo)
	if err != nil {
		panic(err)
	}

	var volumeFilter filter.VolumeFilter
	var soxFilter filter.SoXFilter

	cliPlayer.r = r
	cliPlayer.p = p
	cliPlayer.vf = &volumeFilter
	cliPlayer.sf = &soxFilter
	cliPlayer.volume = 1
}

func play(ctx context.Context) {
	defer cliPlayer.p.Close()
	defer cliPlayer.r.Close()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			reader, err := cliPlayer.r.Record()
			if err != nil {
				panic(err)
			}

			reader, err = cliPlayer.vf.Filter(reader)
			if err != nil {
				panic(err)
			}

			reader, err = cliPlayer.sf.Filter(reader)
			if err != nil {
				panic(err)
			}

			err = cliPlayer.p.Play(reader)
			if err != nil {
				panic(err)
			}
		}
	}
}

func clearTerm() {
	err := term.Sync()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	fmt.Println(
		"[↑]     Volume Up\n" +
			"[↓]     Volume Down\n" +
			"[Space] Mute/Unmute" +
			"\n[Esc]   Exit")
	fmt.Println()
}

func listenKeyEvents() {
	err := term.Init()
	if err != nil {
		panic(err)
	}
	defer term.Close()

	clearTerm()
keyboardListenerLoop:
	for {
		switch ev := term.PollEvent(); ev.Type {
		case term.EventKey:
			switch ev.Key {

			case term.KeyEsc, term.KeyEnter, term.KeyCtrlC:
				clearTerm()
				fmt.Println("Exit")
				break keyboardListenerLoop

			case term.KeyArrowUp, term.KeyArrowLeft:
				clearTerm()
				cliPlayer.volume += 0.1
				fmt.Printf("Volume Up %.2f\n", cliPlayer.volume)
				err = cliPlayer.vf.SetVolume(cliPlayer.volume)
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
				}

			case term.KeyArrowDown, term.KeyArrowRight:
				clearTerm()
				cliPlayer.volume -= 0.1
				fmt.Printf("Volume Down %.2f\n", cliPlayer.volume)
				err = cliPlayer.vf.SetVolume(cliPlayer.volume)
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
				}

			case term.KeySpace:
				clearTerm()
				if cliPlayer.isMute {
					fmt.Println("Unmute")
					cliPlayer.isMute = false
					err = cliPlayer.vf.SetVolume(cliPlayer.volume)
					if err != nil {
						fmt.Fprintln(os.Stderr, err)
					}
				} else {
					fmt.Println("Mute")
					cliPlayer.isMute = true
					err = cliPlayer.vf.SetVolume(0)
					if err != nil {
						fmt.Fprintln(os.Stderr, err)
					}
				}

			default:
			}
		case term.EventError:
			panic(ev.Err)
		}
	}
}

func main() {
	initialize()
	bc := context.Background()
	ctx, cancel := context.WithCancel(bc)
	defer cancel()
	go play(ctx)
	listenKeyEvents()
}
