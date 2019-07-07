package main

import (
	"bufio"
	"context"
	"encoding/binary"
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/ebiiim/eq/filter/function"
	"github.com/ebiiim/eq/filter/pipe"
	"github.com/ebiiim/eq/filter/pipe/sox"
	"github.com/ebiiim/eq/streamio"
	"github.com/ebiiim/eq/streamio/portaudio"
	term "github.com/nsf/termbox-go"
)

type TUI struct {
	r      streamio.Recorder
	p      streamio.Player
	vf     *function.Filter
	sf     *pipe.Filter
	volume float64
	isMute bool
	mu     sync.Mutex
}

var tui TUI

func initialize() error {
	var scanIDLoop = func(s string) int {
		fmt.Print(s)
		sc := bufio.NewScanner(os.Stdin)
		for sc.Scan() {
			s := sc.Text()
			if s == "" {
				return -1
			}
			id, err := strconv.Atoi(s)
			if err != nil {
				fmt.Println(err)
				fmt.Print(s)
				continue
			}
			return id
		}
		return 9999 // unreachable
	}

	const (
		bs   = 4096  // buffer size
		ch   = 2     // channels
		bit  = 16    // bit rate
		rate = 48000 // sampling rate
	)

	ds, err := portaudio.ListDevices()
	if err != nil {
		return err
	}
	for _, v := range ds {
		fmt.Println(v)
	}
	inID := scanIDLoop("Select an input device > ")
	outID := scanIDLoop("Select an output device > ")

	r, err := portaudio.NewRecorder(inID, bs, ch, bit, rate, binary.LittleEndian)
	if err != nil {
		return err
	}

	p, err := portaudio.NewPlayer(outID, bs, ch, bit, rate, binary.LittleEndian)
	if err != nil {
		return err
	}

	var volumeFilter function.Filter
	var soxFilter pipe.Filter

	tui.r = r
	tui.p = p
	tui.vf = &volumeFilter
	tui.sf = &soxFilter
	tui.volume = 1

	var soxCommand sox.Command
	soxCommand.BufferSize = bs
	soxCommand.Effects = []sox.Effect{sox.NewGain(-3.0), sox.NewEQ(80, 5.0, +3)}

	tui.vf.Func, err = function.Volume(tui.volume)
	if err != nil {
		return err
	}
	tui.sf.Cmd = soxCommand.String()
	return nil
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
	b := make([]byte, 4096*2)
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

func startTUI() error {
	var clearTerm = func() {
		err := term.Sync()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		fmt.Printf(
			"[↑]     Volume Up\n" +
				"[↓]     Volume Down\n" +
				"[Space] Mute/Unmute" +
				"\n[Esc]   Exit\n\n")
	}

	var setVolume = func(vol float64) error {
		fn, err := function.Volume(vol)
		if err != nil {
			return err
		}
		tui.mu.Lock()
		tui.vf.Func = fn
		tui.mu.Unlock()
		return nil
	}

	var mute = func() {
		if !tui.isMute {
			fmt.Println("Mute")
			tui.isMute = true
			setVolume(0) // safe
		}
	}

	var unmute = func() {
		if tui.isMute {
			fmt.Println("Unmute")
			tui.isMute = false
			setVolume(tui.volume) // safe
		}
	}

	var volumeUp = func() {
		tui.volume += 0.1
		setVolume(tui.volume) // safe
		fmt.Printf("Volume Up %.2f\n", tui.volume)
	}

	var volumeDown = func() {
		vol := tui.volume - 0.1
		err := setVolume(vol)
		if err != nil {
			fmt.Fprintf(os.Stderr, "err %v", err)
			return
		}
		tui.volume = vol
		fmt.Printf("Volume Down %.2f\n", tui.volume)
	}

	err := term.Init()
	if err != nil {
		return err
	}
	defer term.Close()

	// start the TUI app
	clearTerm()
keyboardListenerLoop:
	for {
		switch ev := term.PollEvent(); ev.Type {
		case term.EventKey:
			switch ev.Key {
			case term.KeyEsc, term.KeyEnter, term.KeyCtrlC:
				break keyboardListenerLoop
			case term.KeyArrowUp, term.KeyArrowLeft:
				clearTerm()
				unmute()
				volumeUp()
			case term.KeyArrowDown, term.KeyArrowRight:
				clearTerm()
				unmute()
				volumeDown()
			case term.KeySpace:
				clearTerm()
				if tui.isMute {
					unmute()
				} else {
					mute()
				}
			default:
				if ev.Ch == 0x71 { // 'q'
					break keyboardListenerLoop
				}
			}
		case term.EventError:
			return ev.Err
		}
	}
	return nil
}

func main() {
	err := initialize()
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}
	bc := context.Background()
	ctx, cancel := context.WithCancel(bc)
	defer cancel()
	go play(ctx)
	err = startTUI()
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}
}
