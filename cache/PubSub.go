package cache

import "sync"

type Message struct {
	msgType string
	body    string
}

type PubSub struct {
	subscribers map[string][]chan Message // map of topic channels
	lock        sync.RWMutex
}

func NewPubSub() *PubSub {
	return &PubSub{
		subscribers: make(map[string][]chan Message),
	}
}

// TODO: `topic` can be a stuct with other meta data
func (ps *PubSub) Subscribe(topic string) chan Message {
	ps.lock.Lock()
	defer ps.lock.Unlock()

	ch := make(chan Message)
	ps.subscribers[topic] = append(ps.subscribers[topic], ch)
	return ch
}

// TODO: `message` can be a struct with details
func (ps *PubSub) Publish(topic string, message Message) {
	ps.lock.RLock()
	defer ps.lock.RUnlock()

	for _, subscriber := range ps.subscribers[topic] {
		go func(ch chan Message) {
			ch <- message
		}(subscriber)
	}
}
