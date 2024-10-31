package shard_manager

import (
	"fmt"

	internal_errors "route256/loms/internal/pkg/errors"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spaolacci/murmur3"
)

type ShardKey string
type ShardIndex int

type ShardFn func(ShardKey) ShardIndex

type ShardManager struct {
	shardFn ShardFn
	shards  []*pgxpool.Pool
}

// NewShardManager creates a new ShardManager instance.
func NewShardManager(shardFn ShardFn, shards []*pgxpool.Pool) *ShardManager {
	return &ShardManager{
		shardFn: shardFn,
		shards:  shards,
	}
}

// GetShardIndex returns the shard index for a given key.
func (sm *ShardManager) GetShardIndex(key ShardKey) ShardIndex {
	return sm.shardFn(key)
}

// GetShard returns the connection pool for the given shard index.
func (sm *ShardManager) GetShard(index ShardIndex) (*pgxpool.Pool, error) {
	if int(index) < len(sm.shards) {
		return sm.shards[index], nil
	}
	return nil, fmt.Errorf("%w: given index=%d, len=%d", internal_errors.ErrShardIndexOutOfRange, index, len(sm.shards))
}

// HashShardFn is a hash function for sharding.
func HashShardFn(shardCount int) ShardFn {
	return func(key ShardKey) ShardIndex {
		h := murmur3.New32()
		h.Write([]byte(key))
		return ShardIndex(uint32(h.Sum32()) % uint32(shardCount))
	}
}

// GetMurmur3ShardFn returns a ShardFn based on the Murmur3 hash function.
func GetMurmur3ShardFn(shardCount int) ShardFn {
	return func(key ShardKey) ShardIndex {
		hasher := murmur3.New32()
		hasher.Write([]byte(key))
		return ShardIndex(hasher.Sum32() % uint32(shardCount))
	}
}
