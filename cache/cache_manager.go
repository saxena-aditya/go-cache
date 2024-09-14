package cache

import (
	"fmt"
	"hash/fnv"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type CacheManager struct {
	dataNodes     []*DataNode   // list of all data nodes
	nodeCount     int           // count of attached nodes
	pubSub        *PubSub       // pubSub protocol to send and listen messages
	lock          sync.RWMutex  // makes sure no two process can operate on same object
	autoEvictTime time.Duration // cleans cache after this interval
}

func NewCacheManager(nodeCount int, autoEvictTime time.Duration) *CacheManager {
	dataNodes := make([]*DataNode, nodeCount)
	pubSub := NewPubSub()
	for i := 0; i < nodeCount; i++ {
		dataNodes[i] = NewDataNode("node-"+strconv.Itoa(i), 2, pubSub)
	}

	manager := &CacheManager{
		dataNodes:     dataNodes,
		pubSub:        pubSub,
		nodeCount:     nodeCount,
		autoEvictTime: autoEvictTime,
	}

	go manager.periodicCleanup()
	return manager
}

func (cm *CacheManager) Set(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	value := r.URL.Query().Get("value")

	node := cm.getNode(key)
	node.Set(key, value)
	fmt.Fprintf(w, "Set key=%s, value=%s\n", key, value)
	// TODO: handle error while setting cache
}

func (cm *CacheManager) Get(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	node := cm.getNode(key)
	value := node.Get(key)

	if value == nil {
		message := fmt.Sprintf("Key %s not found", key)
		http.Error(w, message, http.StatusNotFound)
	} else {
		fmt.Fprintf(w, "DataNode %s: Get key=%s, value=%s\n", node.id, key, *value)
	}
}

// periodicCleanup publishes messages for invalidation of stale entries
func (cm *CacheManager) periodicCleanup() {
	for {
		time.Sleep(200 * time.Second)
		cm.lock.Lock()

		for _, node := range cm.dataNodes {
			for key, elem := range node.store {
				if time.Since(elem.Value.(*CacheNode).timestamp) > cm.autoEvictTime {
					// Publish an invalidation message for the node
					message := Message{
						msgType: "invalidate",
						body:    elem.Value.(*CacheNode).value,
					}
					cm.pubSub.Publish(fmt.Sprintf("evict-%s", cm.getNode(key).id), message)
					fmt.Printf("CacheManager: Published eviction for key=%s\n", key)
				}
			}
		}
		cm.lock.Unlock()
	}
}

func (cm *CacheManager) getNode(key string) *DataNode {
	hash := fnv.New32a()
	hash.Write([]byte(key))
	nodeIndex := int(hash.Sum32()) % cm.nodeCount

	return cm.dataNodes[nodeIndex]
}
