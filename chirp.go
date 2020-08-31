package chirp

import (
	"errors"
	"io"
	"sync"
)

var (
	ErrNoTopic = errors.New("error: topic nonexistent")
)

type Nest struct {
	*sync.Mutex
	*Settings
	topicMap map[string]*clientSlice
}

type Settings struct {
	// Amount of errors before Client is dropped from Nest
	ErrorTolerance int
}

type clientSlice struct {
	*sync.Mutex
	clients []*Client
}

type Client struct {
	*sync.Mutex // locked when writing to Writer
	io.Writer
	id     string
	errors []error
	active bool
}

func NewNest(s *Settings) *Nest {
	if s == nil {
		s = DefaultSettings()
	}
	return &Nest{
		Settings: s,
		Mutex:    &sync.Mutex{},
		topicMap: map[string]*clientSlice{},
	}
}

func DefaultSettings() *Settings {
	return &Settings{
		ErrorTolerance: 5,
	}
}

func (n *Nest) NewClient(topic, id string, dst io.Writer) *Client {
	n.Lock()
	defer n.Unlock()

	client := &Client{
		Mutex:  &sync.Mutex{},
		Writer: dst,
		id:     id,
		errors: []error{},
		active: true,
	}

	slice, ok := n.topicMap[topic]
	if !ok {
		slice = &clientSlice{
			Mutex:   &sync.Mutex{},
			clients: []*Client{client},
		}
		n.topicMap[topic] = slice
	} else {
		slice.Lock()
		slice.clients = append(slice.clients, client)
		slice.Unlock()
	}

	return client
}

func (n *Nest) MsgSubscribers(topic string, msg []byte, ignore ...*Client) error {
	slice, ok := n.topicMap[topic]
	if !ok {
		return ErrNoTopic
	}

	toRemove := []*Client{}
ClientSend:
	for _, c := range slice.clients {
		for _, ign := range ignore {
			if c == ign {
				continue ClientSend
			}
		}

		_ = c.Write(msg)
		if len(c.errors) > n.ErrorTolerance {
			toRemove = append(toRemove, c)
		}
	}

	for _, r := range toRemove {
		slice.removeClient(r)
	}
	return nil
}

func (s *clientSlice) removeClient(c *Client) {
	s.Lock()
	defer s.Unlock()

	toRemove := -1
	for i, client := range s.clients {
		if client == c {
			toRemove = i
		}
	}

	if toRemove != -1 {
		last := len(s.clients) - 1
		s.clients[toRemove] = s.clients[last]
		s.clients[last] = nil
		s.clients = s.clients[:last]
	}
	c.active = false
}

func (c *Client) Write(p []byte) error {
	c.Lock()
	defer c.Unlock()

	_, err := c.Writer.Write(p)
	if err != nil {
		c.errors = append(c.errors, err)
	}
	return err
}
