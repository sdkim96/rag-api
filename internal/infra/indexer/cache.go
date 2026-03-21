package indexer

import (
	"context"
	"fmt"

	"github.com/sdkim96/indexing/cache"
	"github.com/sdkim96/rag-api/internal/infra/db"
)

type DBCache struct {
	db *db.Engine
}

func NewDBCache(db *db.Engine) *DBCache {
	return &DBCache{db: db}
}

func (c *DBCache) GetOrSet(ctx context.Context, key string, fn func() ([]byte, error)) ([]byte, error) {
	// 캐시 조회
	val, err := c.db.GetCache(ctx, key)
	if err != nil {
		return nil, err
	}

	// 캐시 히트
	if val != nil {
		fmt.Printf("[CACHE HIT] %s\n", key)
		return val, nil
	}

	// 캐시 미스 → fn 실행
	fmt.Printf("[CACHE MISS] %s\n", key)
	val, err = fn()
	if err != nil {
		return nil, err
	}

	// 캐시 저장
	if err := c.db.UpsertCache(ctx, key, val, nil); err != nil {
		return nil, err
	}

	return val, nil
}

var _ cache.Cache = (*DBCache)(nil)
