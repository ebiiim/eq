package player

import (
	"io"
)

type Player interface {
	io.WriteCloser
}
