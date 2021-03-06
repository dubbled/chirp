package chirp

import (
	"bytes"
	"math/rand"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestInsertClient(t *testing.T) {
	testNest := NewNest(nil)
	testCount := 1000
	type testPair struct {
		topic string
		count int
	}
	testChan := make(chan testPair, testCount)
	wg := &sync.WaitGroup{}

	for i := 0; i < testCount; i++ {
		wg.Add(1)
		go func() {
			topic := uuid.New().String()
			numClients := rand.Intn(500) + 1
			testChan <- testPair{topic, numClients}

			for j := 0; j < numClients; j++ {
				c := &Client{}
				err := testNest.InsertClient(topic, c)
				assert.Nil(t, err)
				assert.NotNil(t, c)
			}
			wg.Done()
		}()
	}
	wg.Wait()
	close(testChan)

	for pair := range testChan {
		clientCount := len(testNest.topicMap[pair.topic].clients)
		assert.Equal(t, pair.count, clientCount)
	}
}

func TestMsgSubscribers(t *testing.T) {
	testNest := NewNest(nil)
	buffers := []*bytes.Buffer{}
	testTopic := "testing"

	for i := 0; i < 10; i++ {
		buf := bytes.NewBuffer(nil)
		buffers = append(buffers, buf)
		c := &Client{}
		c.SetWriter(buf)
		err := testNest.InsertClient(testTopic, c)
		assert.Nil(t, err)
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
		c := &Client{}
		c.SetWriter(buf)
		err := testNest.InsertClient(testTopic, c)
		assert.Nil(t, err)
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
		err := testNest.InsertClient(testTopic, c)
		assert.Nil(t, err)
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

func TestClientNoWriter(t *testing.T) {
	c := &Client{}
	err := c.Write([]byte("can't write"))
	assert.NotNil(t, err)
}
