package player

import (
	"io"
)

type Player interface {
	Play(r io.Reader) error
	io.Closer
}
