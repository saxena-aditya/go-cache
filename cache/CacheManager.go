package cache

import (
	"fmt"
	"hash/fnv"
	"net/http"
	"strconv"
)

type CacheManager struct {
	dataNodes []*DataNode
	nodeCount int
}

func NewCacheManager(nodeCount int) *CacheManager {
	dataNodes := make([]*DataNode, nodeCount)
	for i := 0; i < nodeCount; i++ {
		dataNodes[i] = NewDataNode("node-" + strconv.Itoa(i))
	}

	return &CacheManager{
		dataNodes: dataNodes,
		nodeCount: nodeCount,
	}
}

func (cm *CacheManager) GetNode(key string) *DataNode {
	hash := fnv.New32a()
	hash.Write([]byte(key))
	nodeIndex := int(hash.Sum32()) % cm.nodeCount

	return cm.dataNodes[nodeIndex]
}

func (cm *CacheManager) Set(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	value := r.URL.Query().Get("value")

	node := cm.GetNode(key)
	node.Set(key, value)
	// TODO: What if set in datanode fails
}

func (cm *CacheManager) Get(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	node := cm.GetNode(key)
	value := node.Get(key)

	if value == nil {
		message := fmt.Sprintf("Key %s not found", key)
		http.Error(w, message, http.StatusNotFound)
	} else {
		fmt.Fprintf(w, *value)
	}
}
