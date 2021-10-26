package core

import (
	"io"
	"os"
)

type logBuffer struct {
	*os.File
}

// LogBuffer allows to read and write command output to a file.
type LogBuffer interface {
	io.WriteCloser
	NewReader() (io.ReadCloser, error)
}

// NewReader returns a stream of the file contents.
func (lb logBuffer) NewReader() (io.ReadCloser, error) {
	f, err := os.Open(lb.Name())
	if err != nil {
		return nil, err
	}
	return f, nil
}

// NewLogBuffer creates the log file to store output to.
func NewLogBuffer(id string) (LogBuffer, error) {
	// TODO: allow users to configure this folder
	f, err := os.Create("/var/log/linux-process-runner/" + id + ".log")
	if err != nil {
		return nil, err
	}
	return logBuffer{File: f}, nil
}
