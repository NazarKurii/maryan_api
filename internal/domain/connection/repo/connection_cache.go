package repo

import (
	"context"
	"time"

	"github.com/d3code/uuid"
	"github.com/nazarkurii/marshrutka_api/internal/entity"
	cache "github.com/nazarkurii/marshrutka_api/internal/infrastructure/cache/connection"
	"github.com/redis/go-redis/v9"
)

type ConnectionCache interface {
	GetFindConnections(ctx context.Context, from, to uuid.UUID, date time.Time) (entity.FindConnectionsResponse, error)
	SetFindConnections(ctx context.Context, from, to uuid.UUID, date time.Time, value entity.FindConnectionsResponse) error
}

type redisCeche struct {
	connection cache.Connection
}

func (r *redisCeche) GetFindConnections(ctx context.Context, from, to uuid.UUID, date time.Time) (entity.FindConnectionsResponse, error) {
	return r.connection.GetFindConnection(ctx, from, to, date)
}

func (r *redisCeche) SetFindConnections(ctx context.Context, from, to uuid.UUID, date time.Time, value entity.FindConnectionsResponse) error {
	return r.connection.SetFindConnection(ctx, from, to, date, value)
}

func NewConnectionCacheRepo(db *redis.Client) ConnectionCache {
	return &redisCeche{
		cache.NewConnection(db),
	}
}
