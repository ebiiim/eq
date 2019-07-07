// Package pipe provides an implementation of filter.Filter
// that using os.Pipe and external applications for filtering.
package pipe

import (
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"

	"github.com/pkg/errors"
)

// Filter is a stream data processor using an external program.
type Filter struct {
	cmd      *exec.Cmd
	initOnce sync.Once
	inPipe   io.WriteCloser
	outPipe  io.ReadCloser

	// Cmd is a command that contains an execPath and args.
	//
	// The command should sequentially read stdin,
	// do some processing, and write into stdout.
	//
	// e.g. "tee -i /dev/null"
	Cmd string
}

func (f *Filter) initialize() (err error) {
	ss := strings.Split(f.Cmd, " ")
	f.cmd = exec.Command(ss[0], ss[1:]...)

	f.inPipe, err = f.cmd.StdinPipe()
	if err != nil {
		return errors.Wrap(err, "could not open stdin pipe")
	}
	f.outPipe, err = f.cmd.StdoutPipe()
	if err != nil {
		return errors.Wrap(err, "could not open stdout pipe")
	}
	err = f.cmd.Start()
	if err != nil {
		return errors.Wrap(err, "could not start exec")
	}
	return nil
}

// Read reads len(b) bytes from
// the output pipe (pipes data from stdout of the external application)
// into b.
//
// The function blocks until it reads len(b) bytes or more.
// The function does not support ioutil.ReadAll (blocks permanently).
func (f *Filter) Read(b []byte) (n int, err error) {
	if err != nil {
		return 0, err
	}
	n, err = f.outPipe.Read(b)
	return
}

// Write writes len(b) bytes from b to the input pipe
// that pipes data to stdin of the external application.
//
// The first call to this function invokes an exec.Command.Start()
// that using an external application (specified in f.Cmd)
// to sequentially reads data from the input pipe,
// process them, and writes the processed data into the output pipe.
func (f *Filter) Write(b []byte) (n int, err error) {
	f.initOnce.Do(func() { err = f.initialize() })

	if err != nil {
		return 0, err
	}
	n, err = f.inPipe.Write(b)
	return
}

// Close closes the Filter object by closing pipes and the external application.
func (f *Filter) Close() (err error) {
	err = f.inPipe.Close()
	if err != nil {
		return errors.Wrap(err, "could not close stdin pipe")
	}
	// FIXME: wait until outPipe is empty
	err = f.outPipe.Close()
	if err != nil {
		return errors.Wrap(err, "could not close stdout pipe")
	}
	err = f.cmd.Wait()
	if err != nil {
		return errors.Wrap(err, "could not wait exec")
	}
	if f.cmd.ProcessState.ExitCode() != 0 {
		return fmt.Errorf("abnormal exit code %d", f.cmd.ProcessState.ExitCode())
	}
	return nil
}
