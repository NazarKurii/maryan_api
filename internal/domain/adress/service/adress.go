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

type Adress interface {
	Create(ctx context.Context, adress entity.Adress, userID uuid.UUID) (uuid.UUID, error)
	Update(ctx context.Context, adress entity.Adress) (uuid.UUID, error)
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (entity.Adress, error)
	GetAdresses(ctx context.Context, paginationStr dbutil.PaginationStr, userID uuid.UUID) ([]entity.Adress, hypermedia.Links, error)
}

type adressServiceImpl struct {
	repo   repo.Adress
	client *http.Client
}

func (a *adressServiceImpl) Create(ctx context.Context, adress entity.Adress, userID uuid.UUID) (uuid.UUID, error) {
	err := adress.Prepare(userID)
	if err != nil {
		return uuid.Nil, err
	}

	return adress.ID, a.repo.Create(ctx, &adress)
}

func (a *adressServiceImpl) Update(ctx context.Context, adress entity.Adress) (uuid.UUID, error) {
	params := adress.Validate()
	if params != nil {
		return uuid.Nil, rfc7807.BadRequest("adress-invalid-data", "Adress Data Error", "Provided data is not valid.", params...)
	}

	exists, usedByTicket, err := a.repo.Status(ctx, adress.ID)
	if err != nil {
		return uuid.Nil, err
	} else if !exists {
		return uuid.Nil, rfc7807.BadRequest("non-existing-adress", "Non-existing Adress Error", "There is no adress associated with provided id.")
	}

	if !usedByTicket {
		err = a.repo.Update(ctx, &adress)
		if err != nil {
			return uuid.Nil, err
		}
	} else {
		err = a.repo.SoftDelete(ctx, adress.ID)
		if err != nil {
			return uuid.Nil, err
		}

		adress.ID = uuid.New()

		err = a.repo.Create(ctx, &adress)
		if err != nil {
			return uuid.Nil, err
		}
	}

	return adress.ID, nil
}

func (a *adressServiceImpl) Delete(ctx context.Context, idStr string) error {

	id, err := uuid.Parse(idStr)
	if err != nil {
		return rfc7807.BadRequest("invalid-id", "Invalid ID Error", err.Error())
	}

	exists, usedByTicket, err := a.repo.Status(ctx, id)
	if err != nil {
		return err
	} else if !exists {
		return rfc7807.BadRequest("non-existing-adress", "Non-existing Adress Error", "There is no adress associated with provided id.")
	}

	if !usedByTicket {
		err = a.repo.ForseDelete(ctx, id)
		if err != nil {
			return err
		}
	} else {
		err = a.repo.SoftDelete(ctx, id)
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *adressServiceImpl) GetByID(ctx context.Context, idStr string) (entity.Adress, error) {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return entity.Adress{}, rfc7807.BadRequest("invalid-id", "Invalid ID Error", err.Error())
	}
	return a.repo.GetByID(ctx, id)
}

func (a *adressServiceImpl) GetAdresses(ctx context.Context, paginationStr dbutil.PaginationStr, userID uuid.UUID) ([]entity.Adress, hypermedia.Links, error) {
	pagination, err := paginationStr.ParseWithCondition(dbutil.Condition{"user_id IN ?", []any{userID}}, "city", "country", "street", "post_code")
	if err != nil {
		return nil, nil, err
	}

	return a.repo.GetAdresses(ctx, pagination)
}

func NewAdressService(repo repo.Adress, client *http.Client) Adress {
	return &adressServiceImpl{repo, client}
}
