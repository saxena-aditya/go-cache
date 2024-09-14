package cache

import (
	"strconv"
	"sync"
	"testing"
	"time"
)

// Test for basic PubSub subscription and publishing
func TestPubSub_SubscribeAndPublish(t *testing.T) {
	ps := NewPubSub()

	// Subscribe to a topic
	topic := "test-topic"
	ch := ps.Subscribe(topic)

	// Publish a message to the topic
	msg := Message{msgType: "test", body: "Hello, World!"}
	go ps.Publish(topic, msg)

	// Receive the message from the subscriber
	select {
	case receivedMsg := <-ch:
		if receivedMsg.msgType != "test" || receivedMsg.body != "Hello, World!" {
			t.Errorf("Expected message {msgType: 'test', body: 'Hello, World!'}, got %v", receivedMsg)
		}
	case <-time.After(1 * time.Second):
		t.Error("Did not receive message in time")
	}
}

// Test for multiple subscribers receiving the same message
func TestPubSub_MultipleSubscribers(t *testing.T) {
	ps := NewPubSub()

	// Subscribe multiple subscribers to the same topic
	topic := "multi-topic"
	ch1 := ps.Subscribe(topic)
	ch2 := ps.Subscribe(topic)

	// Publish a message
	msg := Message{msgType: "test", body: "Multicast"}
	go ps.Publish(topic, msg)

	// Check both subscribers receive the message
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		select {
		case receivedMsg := <-ch1:
			if receivedMsg.body != "Multicast" {
				t.Errorf("Subscriber 1: Expected 'Multicast', got %v", receivedMsg.body)
			}
		case <-time.After(1 * time.Second):
			t.Error("Subscriber 1: Did not receive message in time")
		}
	}()

	go func() {
		defer wg.Done()
		select {
		case receivedMsg := <-ch2:
			if receivedMsg.body != "Multicast" {
				t.Errorf("Subscriber 2: Expected 'Multicast', got %v", receivedMsg.body)
			}
		case <-time.After(1 * time.Second):
			t.Error("Subscriber 2: Did not receive message in time")
		}
	}()

	// Wait for both to finish
	wg.Wait()
}

// Test for no subscribers scenario
func TestPubSub_NoSubscribers(t *testing.T) {
	ps := NewPubSub()

	// Publish a message without any subscribers
	topic := "empty-topic"
	msg := Message{msgType: "test", body: "No subscribers"}
	ps.Publish(topic, msg)

	// No way to verify this, but should run without panic or deadlock
}

// Test for subscribing to multiple topics
func TestPubSub_SubscribeToMultipleTopics(t *testing.T) {
	ps := NewPubSub()

	// Subscribe to two different topics
	ch1 := ps.Subscribe("topic1")
	ch2 := ps.Subscribe("topic2")

	// Publish to both topics
	go ps.Publish("topic1", Message{msgType: "test", body: "Message 1"})
	go ps.Publish("topic2", Message{msgType: "test", body: "Message 2"})

	// Check each channel gets the correct message
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		select {
		case receivedMsg := <-ch1:
			if receivedMsg.body != "Message 1" {
				t.Errorf("Expected 'Message 1', got %v", receivedMsg.body)
			}
		case <-time.After(1 * time.Second):
			t.Error("Did not receive message for topic1 in time")
		}
	}()

	go func() {
		defer wg.Done()
		select {
		case receivedMsg := <-ch2:
			if receivedMsg.body != "Message 2" {
				t.Errorf("Expected 'Message 2', got %v", receivedMsg.body)
			}
		case <-time.After(1 * time.Second):
			t.Error("Did not receive message for topic2 in time")
		}
	}()

	wg.Wait()
}

// Test for concurrent publishing and subscribing
func TestPubSub_ConcurrentPublishSubscribe(t *testing.T) {
	ps := NewPubSub()

	// Create a topic
	topic := "concurrent-topic"

	// Start multiple subscribers
	const numSubscribers = 5
	channels := make([]chan Message, numSubscribers)
	for i := 0; i < numSubscribers; i++ {
		channels[i] = ps.Subscribe(topic)
	}

	// Publish messages concurrently
	const numMessages = 10
	var wg sync.WaitGroup
	wg.Add(numMessages)

	for i := 0; i < numMessages; i++ {
		go func(i int) {
			defer wg.Done()
			ps.Publish(topic, Message{msgType: "test", body: "Message " + strconv.Itoa(i)})
		}(i)
	}

	// Wait for all messages to be published
	wg.Wait()

	// Ensure each subscriber receives all messages
	for i := 0; i < numSubscribers; i++ {
		go func(ch chan Message) {
			for j := 0; j < numMessages; j++ {
				select {
				case <-ch:
				case <-time.After(1 * time.Second):
					t.Errorf("Subscriber did not receive message %d in time", j)
				}
			}
		}(channels[i])
	}
}
