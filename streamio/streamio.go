package streamio

import (
	"io"
)

type Recorder interface {
	io.ReadCloser
}

type Player interface {
	io.WriteCloser
}
