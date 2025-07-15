package service

import (
	"context"
	"maryan_api/internal/domain/adress/repo"
	"maryan_api/internal/entity"
	"maryan_api/pkg/dbutil"
	"maryan_api/pkg/hypermedia"

	rfc7807 "maryan_api/pkg/problem"
	"net/http"

	"github.com/google/uuid"
)

type Address interface {
	// Create(ctx context.Context, Address entity.NewAddress, userID uuid.UUID) (uuid.UUID, error)
	Update(ctx context.Context, Address entity.Address) (uuid.UUID, error)
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (entity.Address, error)
	GetAddresses(ctx context.Context, paginationStr dbutil.PaginationStr, userID uuid.UUID) ([]entity.Address, hypermedia.Links, error)
}

type AddressServiceImpl struct {
	repo   repo.Address
	client *http.Client
}

// func (a *AddressServiceImpl) Create(ctx context.Context, newAddress entity.NewAddress, userID uuid.UUID) (uuid.UUID, error) {
// 	Address := newAddress.ToAddress()
// 	err := Address.Prepare(userID)
// 	if err != nil {
// 		return uuid.Nil, err
// 	}

// 	return Address.ID, a.repo.Create(ctx, &Address)
// }

func (a *AddressServiceImpl) Update(ctx context.Context, Address entity.Address) (uuid.UUID, error) {
	params := Address.Validate()
	if params != nil {
		return uuid.Nil, rfc7807.BadRequest("Address-invalid-data", "Address Data Error", "Provided data is not valid.", params...)
	}

	exists, usedByTicket, err := a.repo.Status(ctx, Address.ID)
	if err != nil {
		return uuid.Nil, err
	} else if !exists {
		return uuid.Nil, rfc7807.BadRequest("non-existing-Address", "Non-existing Address Error", "There is no Address associated with provided id.")
	}

	if !usedByTicket {
		err := a.repo.Update(ctx, &Address)
		if err != nil {
			return uuid.Nil, err
		}
	} else {
		err = a.repo.SoftDelete(ctx, Address.ID)
		if err != nil {
			return uuid.Nil, err
		}

		Address.ID = uuid.New()

		err = a.repo.Create(ctx, &Address)
		if err != nil {
			return uuid.Nil, err
		}
	}

	return Address.ID, nil
}

func (a *AddressServiceImpl) Delete(ctx context.Context, idStr string) error {

	id, err := uuid.Parse(idStr)
	if err != nil {
		return rfc7807.BadRequest("invalid-id", "Invalid ID Error", err.Error())
	}

	// exists, usedByTicket, err := a.repo.Status(ctx, id)
	// if err != nil {
	// 	return err
	// } else if !exists {
	// 	return rfc7807.BadRequest("non-existing-Address", "Non-existing Address Error", "There is no Address associated with provided id.")
	// }

	// if !usedByTicket {
	err = a.repo.ForseDelete(ctx, id)
	if err != nil {
		return err
	}
	// } else {
	// 	err = a.repo.SoftDelete(ctx, id)
	// 	if err != nil {
	// 		return err
	// 	}
	// }

	return nil
}

func (a *AddressServiceImpl) GetByID(ctx context.Context, idStr string) (entity.Address, error) {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return entity.Address{}, rfc7807.BadRequest("invalid-id", "Invalid ID Error", err.Error())
	}
	return a.repo.GetByID(ctx, id)
}

func (a *AddressServiceImpl) GetAddresses(ctx context.Context, paginationStr dbutil.PaginationStr, userID uuid.UUID) ([]entity.Address, hypermedia.Links, error) {

	pagination, err := paginationStr.ParseWithCondition(
		dbutil.Condition{
			Where:  "user_id = ?",
			Values: []any{userID},
		},
		[]string{"country", "city", "street"},
		"city", "country", "street", "post_code", "created_at",
	)
	if err != nil {
		return nil, nil, err
	}

	Addresses, total, err, empty := a.repo.GetAddresses(ctx, pagination)
	if err != nil || empty {
		return nil, nil, err
	}

	return Addresses, hypermedia.Pagination(paginationStr, total), nil
}

func NewAddressService(repo repo.Address, client *http.Client) Address {
	return &AddressServiceImpl{repo, client}
}
