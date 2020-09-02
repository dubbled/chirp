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
	w      io.Writer
	lock   *sync.Mutex // locked when writing to Writer
	id     string
	errors []error
	active bool
}

// Write writes p to a Client's writer
func (c *Client) Write(p []byte) error {
	if c.w == nil {
		return ErrNoWriter
	}

	c.lock.Lock()
	defer c.lock.Unlock()

	_, err := c.w.Write(p)
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

func (c *Client) SetWriter(w io.Writer) *Client {
	c.w = w
	return c
}
