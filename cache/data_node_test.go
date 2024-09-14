package cache

import (
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"
)

// Test for setting and getting cache values
func TestDataNode_SetAndGet(t *testing.T) {
	pubSub := NewPubSub()
	dataNode := NewDataNode("node1", 5, pubSub)

	// Set a value in the cache
	dataNode.Set("key1", "value1")

	// Try to get the value
	value := dataNode.Get("key1")
	if value == nil || *value != "value1" {
		t.Errorf("Expected 'value1', got %v", value)
	}

	// Try to get a non-existent key
	value = dataNode.Get("nonexistent")
	if value != nil {
		t.Errorf("Expected nil for non-existent key, got %v", *value)
	}
}

// Test for LRU eviction policy
func TestDataNode_Eviction(t *testing.T) {
	pubSub := NewPubSub()
	dataNode := NewDataNode("node1", 2, pubSub)

	// Set two values
	dataNode.Set("key1", "value1")
	dataNode.Set("key2", "value2")

	// Set another value to trigger eviction (cache size is 2)
	dataNode.Set("key3", "value3")

	// The least recently used (key1) should be evicted
	if value := dataNode.Get("key1"); value != nil {
		t.Errorf("Expected key1 to be evicted, but got %v", *value)
	}

	// key2 and key3 should still exist
	if value := dataNode.Get("key2"); value == nil || *value != "value2" {
		t.Errorf("Expected 'value2', but got %v", value)
	}

	if value := dataNode.Get("key3"); value == nil || *value != "value3" {
		t.Errorf("Expected 'value3', but got %v", value)
	}
}

// Test for handling pub/sub eviction message
func TestDataNode_PubSubEvict(t *testing.T) {
	pubSub := NewPubSub()
	dataNode := NewDataNode("node1", 5, pubSub)

	// Set a value in the cache
	dataNode.Set("key1", "value1")
	dataNode.Set("key2", "value2")

	// Publish an eviction message
	pubSub.Publish(fmt.Sprintf("evict-%s", dataNode.id), Message{msgType: "invalidate", body: "value1"})

	// Give time for the message to be processed
	time.Sleep(100 * time.Millisecond)

	// Check if "key1" is evicted
	if value := dataNode.Get("key1"); value != nil {
		t.Errorf("Expected key1 to be invalidated, but got %v", *value)
	}

	// key2 should still exist
	if value := dataNode.Get("key2"); value == nil || *value != "value2" {
		t.Errorf("Expected 'value2', but got %v", value)
	}
}

// Test for auto eviction of stale entries after a timeout
func TestDataNode_AutoEvictStaleEntries(t *testing.T) {
	pubSub := NewPubSub()
	dataNode := NewDataNode("node1", 5, pubSub)

	// Set a key and sleep to simulate time passage
	dataNode.Set("key1", "value1")
	time.Sleep(11 * time.Second)

	// Publish an eviction message to invalidate all old entries
	pubSub.Publish(fmt.Sprintf("evict-%s", dataNode.id), Message{msgType: "invalidate", body: "value1"})

	// Give time for the message to be processed
	time.Sleep(100 * time.Millisecond)

	// Check if the stale key1 is evicted
	if value := dataNode.Get("key1"); value != nil {
		t.Errorf("Expected key1 to be evicted due to staleness, but got %v", *value)
	}
}

// Test for concurrent Set and Get operations
func TestDataNode_ConcurrentSetGet(t *testing.T) {
	pubSub := NewPubSub()
	dataNode := NewDataNode("node1", 50, pubSub)

	var wg sync.WaitGroup

	// Concurrently set values
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			dataNode.Set("key"+strconv.Itoa(i), "value"+strconv.Itoa(i))
		}(i)
	}

	// Wait for all sets to complete
	wg.Wait()

	// Concurrently get values
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			value := dataNode.Get("key" + strconv.Itoa(i))
			expected := "value" + strconv.Itoa(i)
			if value == nil || *value != expected {
				t.Errorf("Expected %s, got %v", expected, value)
			}
		}(i)
	}

	// Wait for all gets to complete
	wg.Wait()
}
