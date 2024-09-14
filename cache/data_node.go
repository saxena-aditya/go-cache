package cache

import (
	"container/list"
	"fmt"
	"sync"
	"time"
)

type CacheNode struct {
	key       string
	value     string
	timestamp time.Time
}
type DataNode struct {
	id              string                   // unique id to identify node
	store           map[string]*list.Element // cache store and key => elem pair
	storeSize       int                      // maximum store size
	usageList       *list.List               // tracks accessed cache to maintain least used and most used data
	storeLocks      map[string]*sync.RWMutex // makes sure mulitple process do no try to operate on same cache key
	storeLocksMutex sync.Mutex               // makes sure than store locks are not accessed by more than one process at a time
	pubSub          *PubSub                  // pubSub protocal to listen and send messages
	evictChannel    chan Message             // channel for listening evit messages
}

func NewDataNode(id string, cacheSize int, pubSub *PubSub) *DataNode {
	node := &DataNode{
		id:           id,
		store:        make(map[string]*list.Element),
		storeLocks:   make(map[string]*sync.RWMutex),
		usageList:    list.New(),
		storeSize:    cacheSize,
		pubSub:       pubSub,
		evictChannel: pubSub.Subscribe(fmt.Sprintf("evict-%s", id)),
	}

	go node.listenForEviction()
	return node
}

func (dn *DataNode) getStoreKeyLock(key string) *sync.RWMutex {
	dn.storeLocksMutex.Lock()
	defer dn.storeLocksMutex.Unlock()

	if _, exists := dn.storeLocks[key]; !exists {
		dn.storeLocks[key] = &sync.RWMutex{}
	}

	return dn.storeLocks[key]
}

func (dn *DataNode) Set(key string, value string) {
	lock := dn.getStoreKeyLock(key)
	lock.Lock()
	defer lock.Unlock()

	// if cache key already exists override the key and update usage tracker
	if elem, exists := dn.store[key]; exists {
		dn.usageList.MoveToFront(elem)
		elem.Value.(*CacheNode).value = value
		elem.Value.(*CacheNode).timestamp = time.Now()
		return
	}

	if dn.usageList.Len() >= dn.storeSize {
		// TODO: Use strategy pattern for adding more eviction strategies
		dn.evict()
	}

	node := &CacheNode{
		key:       key,
		value:     value,
		timestamp: time.Now(),
	}
	elem := dn.usageList.PushFront(node)
	dn.store[key] = elem
	fmt.Printf("DataNode %s: Set key=%s, value=%s\n", dn.id, key, value)

}

func (dn *DataNode) Get(key string) *string {
	lock := dn.getStoreKeyLock(key)
	lock.Lock()
	defer lock.Unlock()

	// find the cache key and move it to front to update the usage tracker
	if elem, exists := dn.store[key]; exists {
		dn.usageList.MoveToFront(elem)
		value := elem.Value.(*CacheNode).value
		return &value
	}
	return nil
}

// Simple evict strategy to remove the least recently used (LRU) cache data
func (dn *DataNode) evict() {
	if dn.usageList.Len() == 0 {
		return
	}

	elem := dn.usageList.Back()
	if elem != nil {
		dn.usageList.Remove(elem)
		node := elem.Value.(*CacheNode)
		delete(dn.store, node.key)
		fmt.Printf("DataNode %s: Evicted key=%s\n", dn.id, node.key)
	}
}

// listens for `invalidate` messages from nodes_manager to invalidate the cache entries
func (dn *DataNode) listenForEviction() {
	for msg := range dn.evictChannel {
		fmt.Printf("Node: %s - Listened to message: %s, with body: %s\n", dn.id, msg.msgType, msg.body)
		if msg.msgType == "invalidate" {
			storeKeyLock := dn.getStoreKeyLock(msg.body)
			storeKeyLock.Lock()
			for key, elem := range dn.store {
				if elem.Value.(*CacheNode).value == msg.body {
					dn.usageList.Remove(elem)
					delete(dn.store, key)
					fmt.Printf("DataNode %s: Invalidated stale key=%s\n", dn.id, key)
				}
			}
			storeKeyLock.Unlock()
		}
	}
}
