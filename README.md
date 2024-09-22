# go-cache
Distributed cache written using Go

### About current implementation

1. The cache module consists of two main interfaces: CacheManager and DataNode. CacheManager is responsible for coordinating and tracking all DataNode instances, while each DataNode represents an individual cache node that stores cache data.
1. For every Set and Get request, the cache uses the [32-bit FNV-1a (Fowler–Noll–Vo)](https://en.wikipedia.org/wiki/Fowler%E2%80%93Noll%E2%80%93Vo_hash_function) hash function to compute the address of the DataNode for each {Key, Value} pair.
1. Currently, we store only one copy of each data entry in its respective DataNode.

1. The cache employs a simple LRU (Least Recently Used) eviction policy to remove data when the cache size limit is reached.

1. The cache is designed to prevent concurrent writes for a given key, ensuring data consistency. However, it allows concurrent reads.

1. An auto-cleanup policy or TTL (Time-To-Live) mechanism is implemented using a pub-sub model and Go’s channel construct. The CacheManager publishes an eviction message to all nodes containing stale data, and each DataNode listens to these messages and takes responsibility for invalidating stale entries from its data store.

### Steps to run the cache

The cache can be accessed behind a web server. 
1. Clone `go-cache` pacakge using `$ git clone https://github.com/saxena-aditya/go-cache.git`
1. Make sure your system has `Go` installed.
1. Make sure you are inside `go-cache` directory; execute `$ go run server.go`. 
   * This will start the service on port 8080.
1. In the cache you can Get & Set entries using following end points 
   * `GET http://localhost:8080/cache/get?key=foo` to Get the value of previously set `key` from the cache
   * `GET http://localhost:8080/cache/set?key=foo&value` to Set the `key` & `value` pair in the cache. 

### Notice

Currently, this cache implements a TTL (Time To Live) policy of 60 seconds. The clean up happens every 200 seconds automatically, so make sure you are accessing the values before it for expected results. 
> Above TTL configurations can be modified using CacheManager & DataNode configurations respectively. 