// Package filter provides the Filter interface,
// focusing on working with streamio.Player and streamio.Recorder interfaces.
package filter

import (
	"io"
)

type Filter interface {
	io.ReadWriteCloser
}
