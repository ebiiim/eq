package pipe

import (
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"

	"github.com/pkg/errors"
)

type Filter struct {
	cmd       *exec.Cmd
	initOnce  sync.Once
	inPipe    io.WriteCloser
	outPipe   io.ReadCloser
	FilterCmd string
}

func (f *Filter) initialize() (err error) {
	ss := strings.Split(f.FilterCmd, " ")
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

func (f *Filter) Read(b []byte) (n int, err error) {
	f.initOnce.Do(func() { err = f.initialize() })
	if err != nil {
		return 0, err
	}
	n, err = f.outPipe.Read(b)
	return
}

func (f *Filter) Write(b []byte) (n int, err error) {
	f.initOnce.Do(func() { err = f.initialize() })
	if err != nil {
		return 0, err
	}
	n, err = f.inPipe.Write(b)
	return
}

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
