// Package streamio provides two interfaces, Player and Recorder,
// focusing on working with filter.Filter interface.
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
