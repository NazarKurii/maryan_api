package service

import (
	"context"
	"maryan_api/config"
	"maryan_api/internal/domain/user/repo"
	"maryan_api/internal/entity"
	"maryan_api/pkg/auth"
	"maryan_api/pkg/hypermedia"
	"maryan_api/pkg/images"
	"maryan_api/pkg/pagination"
	rfc7807 "maryan_api/pkg/problem"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

type AdminService interface {
	UserService
	Users(ctx context.Context, cfgStr pagination.CfgStr, rolesStr string) ([]entity.User, hypermedia.Links, error)
	NewUser(ctx context.Context, ru entity.RegistrantionUser, image *multipart.FileHeader, role auth.Role) (string, error)
}

type adminServiceImpl struct {
	UserService
	repo   repo.AdminRepo
	client *http.Client
}

func (as adminServiceImpl) Users(ctx context.Context, cfgStr pagination.CfgStr, rolesStr string) ([]entity.User, hypermedia.Links, error) {
	roles := strings.Split(rolesStr, "+")
	length := len(roles)
	var rolesAny = make([]any, length)
	for i := 0; i < length; i++ {
		rolesAny[i] = roles[i]
	}

	cfg, err := cfgStr.ParseWithCondition(pagination.Condition{"role IN ?", rolesAny})
	if err != nil {
		return nil, nil, err
	}

	users, pages, err := as.repo.Users(ctx, cfg)
	if err != nil {
		return nil, nil, err
	}

	return users, pagination.Links(pages, cfg.Size, "/admin/users"), nil
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
