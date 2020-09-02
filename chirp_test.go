package chirp

import (
	"bytes"
	"math/rand"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestInsertClient(t *testing.T) {
	testNest := NewNest(nil)
	testCount := 1000

	for i := 0; i < testCount; i++ {
		go func() {
			topic := uuid.New().String()
			numClients := rand.Intn(500) + 1
			for j := 0; j < numClients; j++ {
				c := &Client{}
				testNest.InsertClient(topic, c)
				assert.NotNil(t, c)
			}

			clientCount := len(testNest.topicMap[topic].clients)
			assert.Equal(t, numClients, clientCount)
		}()
	}
}

func TestMsgSubscribers(t *testing.T) {
	testNest := NewNest(nil)
	buffers := []*bytes.Buffer{}
	testTopic := "testing"

	for i := 0; i < 10; i++ {
		buf := bytes.NewBuffer(nil)
		buffers = append(buffers, buf)
		c := &Client{Writer: buf}
		testNest.InsertClient(testTopic, c)
	}

	err := testNest.MsgSubscribers(testTopic, []byte("12345"))
	if err != nil {
		t.Fatal(err)
	}

	for _, buf := range buffers {
		assert.Equal(t, "12345", buf.String())
	}
}

func TestMsgSubscribersIgnore(t *testing.T) {
	testNest := NewNest(nil)
	testTopic := "testing"
	ignored := []*Client{}
	buffers := []*bytes.Buffer{}

	for i := 0; i < 10; i++ {
		buf := bytes.NewBuffer(nil)
		buffers = append(buffers, buf)
		c := &Client{Writer: buf}
		testNest.InsertClient(testTopic, c)
		if i%2 == 0 {
			ignored = append(ignored, c)
		}
	}

	err := testNest.MsgSubscribers(testTopic, []byte("testing"), ignored...)
	if err != nil {
		t.Fatal(err)
	}

	for i, buf := range buffers {
		if i%2 == 0 {
			assert.Zero(t, buf.Len())
		} else {
			assert.NotZero(t, buf.Len())
		}
	}
}

func TestRemoveClient(t *testing.T) {
	testNest := NewNest(nil)
	testTopic := "testing"
	clientCount := 20

	toRemove := []*Client{}
	for i := 0; i < clientCount; i++ {
		c := &Client{}
		testNest.InsertClient(testTopic, c)
		if i%2 == 0 {
			toRemove = append(toRemove, c)
		}
	}

	slice, ok := testNest.topicMap[testTopic]
	assert.True(t, ok)
	for _, r := range toRemove {
		slice.removeClient(r)
	}
	assert.Equal(t, clientCount-len(toRemove), len(slice.clients))
}
