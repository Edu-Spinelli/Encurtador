package idgen

import (
	"context"

	"github.com/redis/go-redis/v9"
)

const counterKey = "url:next_id"

type Generator struct {
	rdb         *redis.Client
	startOffset int64
}

func New(rdb *redis.Client, startOffset int64) *Generator {
	return &Generator{rdb: rdb, startOffset: startOffset}
}

func (g *Generator) Init(ctx context.Context) error {
	exists, err := g.rdb.Exists(ctx, counterKey).Result()
	if err != nil {
		return err
	}
	if exists == 0 {
		return g.rdb.Set(ctx, counterKey, g.startOffset, 0).Err()
	}
	return nil
}

func (g *Generator) Next(ctx context.Context) (int64, error) {
	return g.rdb.Incr(ctx, counterKey).Result()
}
