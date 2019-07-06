package alice

import (
	"sync"
	"time"

	"github.com/ebiiim/eq/safe"
)

type Recorder struct {
	initOnce     sync.Once
	readerBuffer safe.Buffer
}

func (r *Recorder) initialize() {
	go func() {
		for {
			gErr := r.record()
			if gErr != nil {
				panic(gErr) // this error should not occur
			}
		}
	}()
}

func (r *Recorder) record() error {
	bs := []byte("Down the Rabbit-Hole\nAlice was beginning to get very tired of sitting by her sister on the bank, and of having nothing to do: once or twice she had peeped into the book her sister was reading, but it had no pictures or conversations in it, `and what is the use of a book,' thought Alice `without pictures or conversation?'\nSo she was considering in her own mind (as well as she could, for the hot day made her feel very sleepy and stupid), whether the pleasure of making a daisy-chain would be worth the trouble of getting up and picking the daisies, when suddenly a White Rabbit with pink eyes ran close by her.\n")
	for i := 0; i < len(bs); i++ {
		_, err := r.readerBuffer.Write([]byte{bs[i]})
		if err != nil {
			return err
		}
		time.Sleep(8 * time.Millisecond)
	}
	return nil
}

func (r *Recorder) Read(b []byte) (n int, err error) {
	r.initOnce.Do(r.initialize)

	readLen := len(b)
	for r.readerBuffer.Len() < readLen {
		time.Sleep(500 * time.Millisecond) // wait for record
	}
	return r.readerBuffer.Read(b)
}

func (r *Recorder) Close() error {
	return nil
}
