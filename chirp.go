package chirp

import (
	"errors"
	"sync"
)

var (
	ErrNoTopic            = errors.New("error: topic nonexistent")
	ErrNoWriter           = errors.New("error: cannot write to nil destination")
	ErrInsertFailedClient = errors.New("error: cannot insert failed client")
)

type Nest struct {
	*Settings
	lock     *sync.Mutex
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

func NewNest(s *Settings) *Nest {
	if s == nil {
		s = DefaultSettings()
	}
	return &Nest{
		Settings: s,
		lock:     &sync.Mutex{},
		topicMap: map[string]*clientSlice{},
	}
}

func DefaultSettings() *Settings {
	return &Settings{
		ErrorTolerance: 5,
	}
}

// InsertClient inserts a new client into the Nest
func (n *Nest) InsertClient(topic string, c *Client) error {
	switch c.State() {
	case FAILED:
		return ErrInsertFailedClient
	case INACTIVE:
		c.lock = &sync.Mutex{}
		c.active = true
		c.errors = []error{}
	}

	n.lock.Lock()
	defer n.lock.Unlock()

	slice, ok := n.topicMap[topic]
	if !ok {
		slice = &clientSlice{
			Mutex:   &sync.Mutex{},
			clients: []*Client{c},
		}
		n.topicMap[topic] = slice
	} else {
		slice.Lock()
		slice.clients = append(slice.clients, c)
		slice.Unlock()
	}
	return nil
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
