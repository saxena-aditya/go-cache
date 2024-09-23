package main

import (
	"fmt"
	"hash/crc32"
	"sort"
	"strconv"
)

// import (
// 	"go-cache/cache"
// 	"log"
// 	"net/http"
// 	"os"
// 	"time"
// )

// var logger = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)

// func main() {

// 	cacheManager := cache.NewCacheManager(3, 60)

// 	http.HandleFunc("/cache/get", cacheManager.Get)
// 	http.HandleFunc("/cache/set", cacheManager.Set)
// 	port := ":8080"

// 	logger.Printf("Starting server on port %s at %s", port, time.Now().Format(time.RFC3339))

// 	err := http.ListenAndServe(port, nil)
// 	if err != nil {
// 		log.Fatalf("Server failed: %s", err)
// 	}
// }

type HashFunc func([]byte) uint32
type ConsistentHashing struct {
	nodes       []int
	nodeHashMap map[int]string
	hashFunc    HashFunc
	replica     int
}

func NewConsistentHashing(replica int, hashFunc HashFunc) *ConsistentHashing {
	if hashFunc == nil {
		hashFunc = crc32.ChecksumIEEE
	}

	return &ConsistentHashing{
		nodeHashMap: make(map[int]string),
		replica:     replica,
		hashFunc:    hashFunc,
	}
}

func (c *ConsistentHashing) AddServer(server string) {

	for i := 0; i < c.replica; i++ {
		hash := int(c.hashFunc([]byte(strconv.Itoa(i) + server)))
		c.nodes = append(c.nodes, hash)
		c.nodeHashMap[hash] = server
	}

	sort.Ints(c.nodes)
}

func (c *ConsistentHashing) RemoveServer(server string) {

	for i := 0; i < c.replica; i++ {
		hash := int(c.hashFunc([]byte(strconv.Itoa(i) + server)))
		index := sort.SearchInts(c.nodes, hash)
		c.nodes = append(c.nodes[:index], c.nodes[index+1:]...)
		delete(c.nodeHashMap, hash)
	}
}

func (c *ConsistentHashing) FindServer(key string) string {

	hash := int(c.hashFunc([]byte(key)))
	fmt.Printf("KeyHash: %s : %s\n", key, hash)
	index := sort.Search(len(c.nodes), func(i int) bool {
		return c.nodes[i] >= hash
	})

	if index == len(c.nodes) {
		index = 0
	}

	return c.nodeHashMap[c.nodes[index]]
}

func main() {
	ch := NewConsistentHashing(1, nil)

	ch.AddServer("london")
	ch.AddServer("paris")
	ch.AddServer("germany")

	keys := []string{"a", "b", "c", "golang"}

	fmt.Printf("Nodes: %+v\n", ch.nodes)
	fmt.Printf("Keys: %+v\n", ch.nodeHashMap)

	for _, key := range keys {
		fmt.Printf("Key: %s got server: %s\n", key, ch.FindServer(key))
	}
}
