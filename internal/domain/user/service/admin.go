package service

import (
	"context"
	"maryan_api/config"
	"maryan_api/internal/domain/user/repo"
	"maryan_api/internal/entity"
	"maryan_api/pkg/auth"
	"maryan_api/pkg/dbutil"
	"maryan_api/pkg/hypermedia"
	"maryan_api/pkg/images"

	rfc7807 "maryan_api/pkg/problem"
	"mime/multipart"
	"net/http"

	"github.com/google/uuid"
)

type AdminService interface {
	UserService
	Users(ctx context.Context, paginationStr dbutil.PaginationStr, rolesStr string) ([]entity.User, hypermedia.Links, error)
	NewUser(ctx context.Context, ru entity.RegistrantionUser, image *multipart.FileHeader, role auth.Role) (string, error)
}

type adminServiceImpl struct {
	UserService
	repo   repo.AdminRepo
	client *http.Client
}

func (as adminServiceImpl) GetCustomers(ctx context.Context, paginationStr dbutil.PaginationStr) ([]entity.User, hypermedia.Links, error) {

	pagination, err := paginationStr.ParseWithCondition(dbutil.Condition{"role IN ?", []any{rolesAny}})
	if err != nil {
		return nil, nil, err
	}

	return as.repo.Users(ctx, pagination)
}

func (as *adminServiceImpl) NewUser(ctx context.Context, ru entity.RegistrantionUser, image *multipart.FileHeader, role auth.Role) (string, error) {
	u, invalidParams := ru.ToNewUser(role)

	if invalidParams != nil {
		return "", rfc7807.BadRequest(
			"user-credentials-validation",
			"user Credentials Error",
			"Could not save the user due to invalid credentials.",
			invalidParams...,
		)
	}

	u.ID = uuid.New()

	if image != nil {
		err := images.Save("../../../../static/imgs/"+u.ID.String(), image)
		if err != nil {
			return "", rfc7807.Internal("image-saving-error", err.Error())
		}
		u.ImageUrl = config.APIURL() + "/imgs/" + u.ID.String()
	} else {
		u.ImageUrl = config.APIURL() + "/imgs/guest-female.png"
	}

	u.Role.Val = auth.Customer
	err := as.repo.NewUser(ctx, &u)
	if err != nil {
		return "", err
	}

	token, err := u.Role.Val.GenerateToken(u.Email, u.ID)
	return token, err
}

// Constructor function
func NewAdminServiceImpl(repo repo.AdminRepo, client *http.Client) AdminService {
	return &adminServiceImpl{
		UserService: NewUserService(auth.Admin, repo),
		repo:        repo,
		client:      client,
	}
}
