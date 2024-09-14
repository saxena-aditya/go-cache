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
	id              string
	store           map[string]*list.Element
	storeSize       int
	usageList       *list.List
	storeLocks      map[string]*sync.RWMutex
	storeLocksMutex sync.Mutex
	pubSub          *PubSub
	evictChannel    chan Message
	autoEvictTime   int
}

func NewDataNode(id string, cacheSize int, pubSub *PubSub) *DataNode {
	node := &DataNode{
		id:            id,
		store:         make(map[string]*list.Element),
		storeLocks:    make(map[string]*sync.RWMutex),
		usageList:     list.New(),
		storeSize:     cacheSize,
		pubSub:        pubSub,
		evictChannel:  pubSub.Subscribe(fmt.Sprintf("evict-%s", id)),
		autoEvictTime: 10,
	}

	//go node.listenEvictionMessages()
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

	if elem, exists := dn.store[key]; exists {
		dn.usageList.MoveToFront(elem)
		elem.Value.(*CacheNode).value = value
		elem.Value.(*CacheNode).timestamp = time.Now()
		return
	}

	if dn.usageList.Len() >= dn.storeSize {
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

	if elem, exists := dn.store[key]; exists {
		dn.usageList.MoveToFront(elem)
		value := elem.Value.(*CacheNode).value
		return &value
	}
	return nil
}

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
