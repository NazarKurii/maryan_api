package service

import (
	"context"
	"fmt"
	"maryan_api/config"

	"maryan_api/internal/domain/bus/repo"
	"maryan_api/internal/entity"
	"maryan_api/pkg/dbutil"
	"maryan_api/pkg/hypermedia"
	"maryan_api/pkg/images"

	rfc7807 "maryan_api/pkg/problem"
	"mime/multipart"

	"github.com/google/uuid"
)

type Bus interface {
	Create(ctx context.Context, bus entity.Bus, busImages []*multipart.FileHeader) (uuid.UUID, error)
	GetByID(ctx context.Context, id string) (entity.Bus, error)
	GetBuses(ctx context.Context, cfgStr dbutil.PaginationStr) ([]entity.Bus, hypermedia.Links, error)
	Delete(ctx context.Context, id string) error
	MakeActive(ctx context.Context, id string) error
	MakeInActive(ctx context.Context, id string) error
}

type busServiceImpl struct {
	repo repo.Bus
}

func (b *busServiceImpl) Create(ctx context.Context, bus entity.Bus, busImages []*multipart.FileHeader) (uuid.UUID, error) {
	params := bus.Prepare()
	if len(busImages) == 0 {
		params.SetInvalidParam("images", "No images attached")
	}
	if len(params) != 0 {
		return uuid.Nil, rfc7807.BadRequest("invalid-bus-data", "Invalid Bus Data Error", "Invalid params.", params...)
	}

	for i, image := range busImages {
		id := uuid.NewString()
		if err := images.Save("../../../../static/imgs/"+id, image); err != nil {
			params.SetInvalidParam(fmt.Sprintf("image(index:%d)", i), err.Error())
		} else {
			bus.Images = append(bus.Images, entity.BusImage{bus.ID, config.APIURL() + "/imgs/" + id})
		}
	}

	if len(params) != 0 {
		return uuid.Nil, rfc7807.BadRequest("invalid-bus-data", "Invalid Bus Data Error", "Invalid params.", params...)
	}

	return bus.ID, b.repo.Create(ctx, &bus)
}

func (b *busServiceImpl) GetByID(ctx context.Context, id string) (entity.Bus, error) {
	uuid, err := uuid.Parse(id)
	if err != nil {
		return entity.Bus{}, rfc7807.UUID(err.Error())
	}
	return b.repo.GetByID(ctx, uuid)
}

func (b *busServiceImpl) GetBuses(ctx context.Context, paginationStr dbutil.PaginationStr) ([]entity.Bus, hypermedia.Links, error) {
	pagination, err := paginationStr.Parse("name", "manufacturer", "date")
	if err != nil {
		return nil, nil, err
	}
	return b.repo.GetBuses(ctx, pagination)

}

func (b *busServiceImpl) Delete(ctx context.Context, id string) error {
	uuid, err := uuid.Parse(id)
	if err != nil {
		return rfc7807.UUID(err.Error())
	}
	return b.repo.Delete(ctx, uuid)
}

func (b *busServiceImpl) MakeActive(ctx context.Context, id string) error {
	uuid, err := uuid.Parse(id)
	if err != nil {
		return rfc7807.UUID(err.Error())
	}

	isActive, err := b.repo.IsActive(ctx, uuid)
	if err != nil {
		return err
	}

	if isActive {
		return rfc7807.BadRequest("already-active", "Already Active Bus Error", "The bus already has an 'Active' status.")
	}

	return b.repo.MakeActive(ctx, uuid)
}

func (b *busServiceImpl) MakeInActive(ctx context.Context, id string) error {
	uuid, err := uuid.Parse(id)
	if err != nil {
		return rfc7807.UUID(err.Error())
	}

	isActive, err := b.repo.IsActive(ctx, uuid)
	if err != nil {
		return err
	}

	if !isActive {
		return rfc7807.BadRequest("already-inactive", "Already Inactive Bus Error", "The bus already has an 'Inactive' status.")
	}

	return b.repo.MakeInactive(ctx, uuid)
}

// --------------------Services Initialization Functions

func NewBusService(repo repo.Bus) Bus {
	return &busServiceImpl{repo}
}
