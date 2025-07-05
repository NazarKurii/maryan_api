package dataStore

import (
	"context"
	"maryan_api/internal/entity"

	"github.com/google/uuid"
)

type Trip interface {
	Create(ctx context.Context, trip *entity.Trip) error
	GetByID(ctx context.Context, id uuid.UUID) (entity.Trip, error)
	Delete(ctx context.Context, id uuid.UUID) error
	AddStop(ctx context.Context, stop entity.Stop, id uuid.UUID) error
	RemoveStop(ctx context.Context, stopID, tripID uuid.UUID) error
	StopMissed(ctx context.Context, id uuid.UUID) error
	StopCompleted(ctx context.Context, id uuid.UUID) error
	Cancel(ctx context.Context, id uuid.UUID) error
	ChangeDriver(ctx context.Context, id, driverID uuid.UUID) error
}
