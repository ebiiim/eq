package main

import (
	"fmt"
	"os"
	"time"

	"github.com/ebiiim/eq/streamio/alice"
)

func main() {
	const (
		bufferSize = 32
		duration   = 10 * time.Second
	)
	// recorder
	r := alice.Recorder{}
	// player
	f, err := os.Create("out.txt")
	if err != nil {
		panic(err)
	}
	p, err := alice.NewPlayer(f, bufferSize)
	if err != nil {
		panic(err)
	}

	buf := make([]byte, bufferSize)

	timeoutCh := make(chan struct{})
	go func(ch chan struct{}) {
		<-time.After(duration)
		close(ch)
	}(timeoutCh)

loop:
	for {
		select {
		case <-timeoutCh:
			fmt.Printf("\n%.0f seconds passed\n", duration.Seconds())
			break loop
		default:
			_, err := r.Read(buf)
			if err != nil {
				panic(err)
			}
			fmt.Print("R")
			_, err = p.Write(buf)
			if err != nil {
				panic(err)
			}
			fmt.Print("W")
		}
	}
}
