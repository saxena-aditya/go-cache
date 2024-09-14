package main

import (
	"fmt"
	"sync"
)

type DataNode struct {
	id              string
	store           map[string]string
	storeLocks      map[string]*sync.RWMutex
	storeLocksMutex sync.Mutex
}

func NewDataNode(id string) *DataNode {
	node := &DataNode{
		id:         id,
		store:      make(map[string]string),
		storeLocks: make(map[string]*sync.RWMutex),
	}

	return node
}

func (dn *DataNode) getStoreKeyLock(key string) *sync.RWMutex {
	dn.storeLocksMutex.Lock()
	defer dn.storeLocksMutex.Lock()

	if _, exists := dn.storeLocks[key]; !exists {
		dn.storeLocks[key] = &sync.RWMutex{}
	}

	return dn.storeLocks[key]
}

func (dn *DataNode) Set(key string, value string) {
	lock := dn.getStoreKeyLock(key)
	lock.Lock()
	defer lock.Lock()

	dn.store[key] = value
	fmt.Printf("DataNode: %s, SET %s = %s", dn.id, key, value)
}

func (dn *DataNode) Get(key string) *string {
	lock := dn.getStoreKeyLock(key)
	lock.Lock()
	defer lock.Lock()

	if value, exists := dn.store[key]; exists {
		return &value
	}
	return nil
}
