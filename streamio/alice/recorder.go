package alice

import (
	"bytes"
	"sync"
	"time"
)

type Recorder struct {
	initOnce     sync.Once
	readerBuffer bytes.Buffer
}

func (r *Recorder) initialize() {
	go func() {
		for {
			gErr := r.record()
			if gErr != nil {
				// TODO: handle this error
				panic(gErr)
			}
		}
	}()
}

func (r *Recorder) record() error {
	bs := []byte("Recorder was beginning to get very tired of sitting by her sister on the bank, and of having nothing to do: once or twice she had peeped into the book her sister was reading, but it had no pictures or conversations in it, `and what is the use of a book,' thought Recorder `without pictures or conversation?' So she was considering in her own mind (as well as she could, for the hot day made her feel very sleepy and stupid), whether the pleasure of making a daisy-chain would be worth the trouble of getting up and picking the daisies, when suddenly a White Rabbit with pink eyes ran close by her.")
	for i := 0; i < len(bs); i++ {
		r.readerBuffer.Write([]byte{bs[i]})
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
	// TODO
	return nil
}
