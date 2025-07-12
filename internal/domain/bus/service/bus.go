package service

import (
	"context"
	"fmt"
	"maryan_api/config"
	"time"

	"maryan_api/internal/domain/bus/repo"
	"maryan_api/internal/entity"
	"maryan_api/pkg/dbutil"
	"maryan_api/pkg/hypermedia"
	"maryan_api/pkg/images"
	"maryan_api/pkg/timeutil"

	rfc7807 "maryan_api/pkg/problem"
	"mime/multipart"

	"github.com/google/uuid"
)

type Bus interface {
	Create(ctx context.Context, bus entity.Bus, busImages []*multipart.FileHeader) (uuid.UUID, error)
	GetByID(ctx context.Context, id string) (entity.Bus, error)
	GetBuses(ctx context.Context, cfgStr dbutil.PaginationStr) ([]entity.Bus, hypermedia.Links, error)
	Delete(ctx context.Context, id string) error
	ChangeDriver(driverType driverType) func(ctx context.Context, busIDStr, driverIDStr string) error
	GetAvailable(ctx context.Context, paginationStr dbutil.PaginationStr, fromStr, toStr string) ([]entity.Bus, hypermedia.Links, error)
	SetSchedule(ctx context.Context, schedule []entity.BusAvailability) error
}

type busServiceImpl struct {
	bus    repo.Bus
	driver repo.Driver
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

	return bus.ID, b.bus.Create(ctx, &bus)
}

func (b *busServiceImpl) GetByID(ctx context.Context, id string) (entity.Bus, error) {
	uuid, err := uuid.Parse(id)
	if err != nil {
		return entity.Bus{}, rfc7807.UUID(err.Error())
	}
	return b.bus.GetByID(ctx, uuid)
}

func (b *busServiceImpl) GetBuses(ctx context.Context, paginationStr dbutil.PaginationStr) ([]entity.Bus, hypermedia.Links, error) {
	pagination, err := paginationStr.Parse([]string{"model", "registration_number", "manufacturer", "year"}, "manufacturer", "year")
	if err != nil {
		return nil, nil, err
	}

	buses, total, err := b.bus.GetBuses(ctx, pagination)
	if err != nil {
		return nil, nil, err
	}

	return buses, hypermedia.Pagination(paginationStr, total), nil

}

func (b *busServiceImpl) Delete(ctx context.Context, id string) error {
	uuid, err := uuid.Parse(id)
	if err != nil {
		return rfc7807.UUID(err.Error())
	}
	return b.bus.Delete(ctx, uuid)
}

func (b *busServiceImpl) GetAvailable(ctx context.Context, paginationStr dbutil.PaginationStr, fromStr, toStr string) ([]entity.Bus, hypermedia.Links, error) {
	pagination, err := paginationStr.Parse([]string{"model", "registration_number", "manufacturer", "year"}, "manufacturer", "year")
	if err != nil {
		return nil, nil, err
	}

	from, err := time.Parse("2006-01-02T15:04:05Z", fromStr)
	if err != nil {
		return nil, nil, rfc7807.BadRequest("invalid-from-time", "Invalid From Time Error", err.Error())
	}

	to, err := time.Parse("2006-01-02T15:04:05Z", toStr)
	if err != nil {
		return nil, nil, rfc7807.BadRequest("invalid-to-time", "Invalid To Time Error", err.Error())
	}

	buses, total, err := b.bus.GetAvailable(ctx, timeutil.FromTo(from, to), pagination)
	if err != nil {
		return nil, nil, err
	}

	return buses, hypermedia.Pagination(paginationStr, total,
		hypermedia.DefaultParam{
			"from",
			"",
			fromStr,
		}, hypermedia.DefaultParam{
			"to",
			"",
			toStr,
		}), nil
}

type driverType int

const (
	AssistantDriver driverType = 0
	LeadDriver      driverType = 0
)

func (b *busServiceImpl) ChangeDriver(driverType driverType) func(ctx context.Context, busIDStr, driverIDStr string) error {
	var driverChanginFunc func(context.Context, uuid.UUID, uuid.UUID) error

	if driverType == AssistantDriver {
		driverChanginFunc = b.bus.ChangeAssistantDriver
	} else {
		driverChanginFunc = b.bus.ChangeLeadDriver
	}

	return func(ctx context.Context, busIDStr, driverIDStr string) error {
		var params rfc7807.InvalidParams

		busID, err := uuid.Parse(busIDStr)
		if err != nil {
			params.SetInvalidParam("busId", err.Error())
		}

		driverID, err := uuid.Parse(driverIDStr)
		if err != nil {
			params.SetInvalidParam("budrievrId", err.Error())
		}

		if params != nil {
			return rfc7807.BadRequest("ivalid-id", "Invalid ID Error", "Provided id is not valid", params...)
		}

		exists, err := b.driver.Exists(ctx, driverID)
		if err != nil {
			return err
		} else if !exists {
			return rfc7807.BadRequest("non-existing-user", "Non-existing User Error", "There is no driver assosiated with provided id.")
		}

		exists, err = b.bus.Exists(ctx, busID)
		if err != nil {
			return err
		} else if !exists {
			return rfc7807.BadRequest("non-existing-bus", "Non-existing Bus Error", "There is no bus assosiated with provided id.")
		}

		return driverChanginFunc(ctx, busID, driverID)
	}

}

func (b *busServiceImpl) SetSchedule(ctx context.Context, schedule []entity.BusAvailability) error {
	var invalidParams rfc7807.InvalidParams
	busID := schedule[0].BusID
	for _, availability := range schedule {
		if !availability.Status.IsValid() {
			invalidParams.SetInvalidParam(availability.Date.String(), "invalid status.")
		}

		if busID != availability.BusID {
			invalidParams.SetInvalidParam(availability.Date.String(), "busID differs.")
		}
	}

	if invalidParams != nil {
		return rfc7807.BadRequest("invalid-bus-schedule", "Invalid Bus Schedule Error", "Provided params are not valid.", invalidParams...)
	}

	exists, err := b.bus.Exists(ctx, busID)
	if err != nil {
		return err
	}

	if !exists {
		return rfc7807.BadRequest("non-existing-bus", "Non-existring Bus Error", "There is no bus assosiated with provided id.")
	}

	return b.bus.SetSchedule(ctx, schedule)
}

// --------------------Services Initialization Functions

func NewBusService(bus repo.Bus, driver repo.Driver) Bus {
	return &busServiceImpl{bus, driver}
}
