package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"os"
	"sync"

	"github.com/ebiiim/eq/filter"
	"github.com/ebiiim/eq/filter/sox"
	"github.com/ebiiim/eq/streamio"
	"github.com/ebiiim/eq/streamio/portaudio"
	term "github.com/nsf/termbox-go"
)

type TUI struct {
	r      streamio.Recorder
	p      streamio.Player
	vf     *filter.Func
	sf     *filter.Pipe
	volume float64
	isMute bool
	mu     sync.Mutex
}

var tui TUI

func initialize() {
	var (
		bs   = 4096
		ch   = 2
		bit  = 16
		rate = 48000
		bo   = binary.LittleEndian
	)

	r, err := portaudio.NewRecorder(bs, ch, bit, rate, bo)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	p, err := portaudio.NewPlayer(bs, ch, bit, rate, bo)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	var volumeFilter filter.Func
	var soxFilter filter.Pipe

	tui.r = r
	tui.p = p
	tui.vf = &volumeFilter
	tui.sf = &soxFilter
	tui.volume = 1

	var soxCommand sox.Command
	soxCommand.Effects = []sox.Effect{sox.NewGain(-3.0), sox.NewEQ(80, 5.0, +3)}

	tui.mu.Lock()
	tui.vf.FilterFunc = filter.Volume(tui.volume)
	tui.sf.FilterCmd = soxCommand.Get()
	tui.mu.Unlock()
}

func play(ctx context.Context) {
	defer func() {
		err := tui.p.Close()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}()
	defer func() {
		err := tui.r.Close()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}()

	b := make([]byte, 1024*8)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			_, err := tui.r.Read(b)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
			_, err = tui.sf.Write(b)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
			_, err = tui.sf.Read(b)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
			_, err = tui.vf.Write(b)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
			_, err = tui.vf.Read(b)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
			_, err = tui.p.Write(b)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
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
				break keyboardListenerLoop
			case term.KeyArrowUp, term.KeyArrowLeft:
				clearTerm()
				tui.volume += 0.1
				fmt.Printf("Volume Up %.2f\n", tui.volume)
				tui.mu.Lock()
				tui.vf.FilterFunc = filter.Volume(tui.volume)
				tui.mu.Unlock()
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
				}
			case term.KeyArrowDown, term.KeyArrowRight:
				clearTerm()
				tui.volume -= 0.1
				fmt.Printf("Volume Down %.2f\n", tui.volume)
				tui.mu.Lock()
				tui.vf.FilterFunc = filter.Volume(tui.volume)
				tui.mu.Unlock()
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
				}
			case term.KeySpace:
				clearTerm()
				if tui.isMute {
					fmt.Println("Unmute")
					tui.isMute = false
					tui.mu.Lock()
					tui.vf.FilterFunc = filter.Volume(tui.volume)
					tui.mu.Unlock()
					if err != nil {
						fmt.Fprintln(os.Stderr, err)
					}
				} else {
					fmt.Println("Mute")
					tui.isMute = true
					tui.mu.Lock()
					tui.vf.FilterFunc = filter.Volume(0)
					tui.mu.Unlock()
					if err != nil {
						fmt.Fprintln(os.Stderr, err)
					}
				}
			default: // do nothing
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
