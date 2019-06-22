package recorder

import (
	"io"
	"os"
)

type Recorder interface {
	Record() (r io.Reader, err error)
	io.Closer
}

type FileRecorder struct {
	file *os.File
}

func NewFileRecorder(filePath string) (r *FileRecorder, err error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	r = &FileRecorder{file: f}
	return r, nil
}

func (recorder *FileRecorder) Record() (r io.Reader, err error) {
	return recorder.file, nil
}

func (recorder *FileRecorder) Close() error {
	err := recorder.file.Close()
	if err != nil {
		return err
	}
	return nil
}
