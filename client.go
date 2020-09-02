package chirp

import (
	"io"
	"sync"
)

type ClientState int

const (
	ACTIVE ClientState = iota
	FAILED
	INACTIVE
)

type Client struct {
	io.Writer
	lock   *sync.Mutex // locked when writing to Writer
	id     string
	errors []error
	active bool
}

// Write writes p to a Client's writer
func (c *Client) Write(p []byte) error {
	if c.Writer == nil {
		return ErrNoWriter
	}

	c.lock.Lock()
	defer c.lock.Unlock()

	_, err := c.Writer.Write(p)
	if err != nil {
		c.errors = append(c.errors, err)
	}
	return err
}

// State returns the ClientState of the client
func (c *Client) State() ClientState {
	if !c.active && c.lock == nil {
		return INACTIVE
	} else if !c.active && c.lock != nil {
		return FAILED
	}
	return ACTIVE
}

func (c *Client) SetID(id string) *Client {
	c.id = id
	return c
}
