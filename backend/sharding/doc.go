// Package sharding provides a rate-limiting backend that distributes
// requests across multiple inner backends using consistent hashing.
//
// Each unique key is deterministically routed to the same shard on every
// call, ensuring that rate-limiting counters for a given key are always
// maintained by a single backend instance. This makes it straightforward
// to scale out memory or Redis backends without cross-shard coordination.
//
// # Example
//
//	shard1, _ := memory.New(memory.Options{Limit: 100, Window: time.Minute})
//	shard2, _ := memory.New(memory.Options{Limit: 100, Window: time.Minute})
//	shard3, _ := memory.New(memory.Options{Limit: 100, Window: time.Minute})
//
//	limiter, err := sharding.New(sharding.Options{
//		Shards: []backend.Backend{shard1, shard2, shard3},
//	})
//	if err != nil {
//		log.Fatal(err)
//	}
//
// Each key will always be routed to the same shard, providing consistent
// rate limiting without any inter-shard communication.
package sharding
