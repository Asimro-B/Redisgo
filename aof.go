package main

import (
	"bufio"
	"io"
	"os"
	"sync"
	"time"
)

// create a struct which stores file that will be stored on the disk and bufio.reader to read the RESP files from the file
type Aof struct {
	file *os.File
	rd   *bufio.Reader
	mu   sync.Mutex
}

func NewAof(path string) (*Aof, error) {
	// create a file if it does not exist or open if it exists
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 067)
	if err != nil {
		return nil, err
	}

	// create a reader to read from the file
	aof := &Aof{
		file: f,
		rd:   bufio.NewReader(f),
	}

	// start goroutine to sync the aof file to disk every 1second while the server is running
	go func() {
		for {
			aof.mu.Lock()
			aof.file.Sync()
			aof.mu.Unlock()
			time.Sleep(time.Second)
		}
	}()

	return aof, nil
}

// close: ensures the file is properly closed when the server shuts down
func (aof *Aof) Close() error {
	aof.mu.Lock()
	defer aof.mu.Unlock()

	return aof.file.Close()
}

// create a write method that will be used to write the command to the AOF file whenever we recieve a request from the client
func (aof *Aof) write(value Value) error {
	aof.mu.Lock()
	defer aof.mu.Unlock()

	_, err := aof.file.Write(value.Marshal())
	if err != nil {
		return err
	}

	return nil
}

func (aof *Aof) Read(callback func(value Value)) error {
	aof.mu.Lock()
	defer aof.mu.Unlock()

	resp := NewResp(aof.rd)
	for {
		value, err := resp.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		callback(value)
	}

	return nil
}
