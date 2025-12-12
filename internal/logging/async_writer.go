package logging

import (
	"os"
)

type AsyncWriter struct {
	buffer chan []byte
	file   *os.File
	done   chan struct{}
}

func NewAsyncWriter(path string, bufferSize int) (*AsyncWriter, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	aw := &AsyncWriter{
		buffer: make(chan []byte, bufferSize),
		file:   file,
		done:   make(chan struct{}),
	}

	go aw.writeLoop()

	return aw, nil
}

func (aw *AsyncWriter) Write(data []byte) {
	select {
	case aw.buffer <- data:
	default:
	}
}

func (aw *AsyncWriter) writeLoop() {
	for {
		select {
		case data := <-aw.buffer:
			aw.file.Write(data)
		case <-aw.done:
			for len(aw.buffer) > 0 {
				data := <-aw.buffer
				aw.file.Write(data)
			}
			return
		}
	}
}

func (aw *AsyncWriter) Close() error {
	close(aw.done)
	return aw.file.Close()
}
