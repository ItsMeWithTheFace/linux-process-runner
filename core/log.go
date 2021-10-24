package core

import (
	"io"
	"os"
)

type logBuffer struct {
	*os.File
}

type LogBuffer interface {
	io.WriteCloser
	NewReader() (io.ReadCloser, error)
}

func (lb logBuffer) NewReader() (io.ReadCloser, error) {
	f, err := os.Open(lb.Name())
	if err != nil {
		return nil, err
	}
	return f, nil
}

func NewLogBuffer(id string) (LogBuffer, error) {
	f, err := os.Create("/var/log/linux-process-runner/" + id + ".log")
	if err != nil {
		return nil, err
	}
	return logBuffer{File: f}, nil
}
