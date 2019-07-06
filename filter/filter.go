package filter

import (
	"io"
)

type Filter interface {
	io.ReadWriteCloser
}
