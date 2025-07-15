package repo

import (
	"context"

	"github.com/google/uuid"
)

type Trip interface {
}

type Bus interface {
	Exists(ctx context.Context, id uuid.UUID) (bool, error)
}

type Driver interface {
	Exists(ctx context.Context, id uuid.UUID) (bool, error)
}
type Stop interface {
	Complete(ctx context.Context, id uuid.UUID) error
	MakeMissed(ctx context.Context, id uuid.UUID) error
}
