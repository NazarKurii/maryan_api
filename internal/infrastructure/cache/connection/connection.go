package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/d3code/uuid"
	"github.com/nazarkurii/marshrutka_api/internal/entity"
	connectionpb "github.com/nazarkurii/marshrutka_api/internal/proto/proto/connection"
	rfc7807 "github.com/nazarkurii/marshrutka_api/pkg/problem"
	"github.com/redis/go-redis/v9"
	"google.golang.org/protobuf/proto"
)

type Connection interface {
	GetFindConnection(ctx context.Context, from, to uuid.UUID, date time.Time) (entity.FindConnectionsResponse, error)
	SetFindConnection(ctx context.Context, from, to uuid.UUID, date time.Time, value entity.FindConnectionsResponse) error
}

type connectionsRedis struct {
	db *redis.Client
}

func (cr *connectionsRedis) GetFindConnection(ctx context.Context, from, to uuid.UUID, date time.Time) (entity.FindConnectionsResponse, error) {
	data, err := cr.db.Get(ctx, fmt.Sprintf("findConnections:%s-%s-%s", from.String(), to.String(), date.String())).Bytes()
	if err != nil {
		return entity.FindConnectionsResponse{}, rfc7807.DB(err.Error())
	}

	var pbResp connectionpb.FindConnectionsResponse
	if err := proto.Unmarshal(data, &pbResp); err != nil {
		return entity.FindConnectionsResponse{}, rfc7807.DB(err.Error())
	}

	return fromProto(&pbResp), nil
}

func (cr *connectionsRedis) SetFindConnection(ctx context.Context, from, to uuid.UUID, date time.Time, value entity.FindConnectionsResponse) error {
	pbResp := toProto(value)

	data, err := proto.Marshal(pbResp)
	if err != nil {
		return err
	}

	result := cr.db.Set(ctx, fmt.Sprintf("findConnections:%s-%s-%s", from.String(), to.String(), date.String()), data, 0)
	if err := result.Err(); err != nil {
		return rfc7807.DB(err.Error())
	}

	return nil
}

func NewConnection(db *redis.Client) Connection {
	return &connectionsRedis{db}
}
