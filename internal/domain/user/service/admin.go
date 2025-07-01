package service

import (
	"context"
	"fmt"
	"maryan_api/config"
	"maryan_api/internal/domain/user/repo"
	"maryan_api/internal/entity"
	"maryan_api/pkg/auth"
	"maryan_api/pkg/hypermedia"
	"net/http"
	"strconv"
)

type AdminService interface {
	UserService
	Users(cfgStr entity.UsersPaginationStr, ctx context.Context) ([]entity.User, hypermedia.Links, error)
}
type adminServiceImpl struct {
	UserService
	repo   repo.AdminRepo
	client *http.Client
}

func (as adminServiceImpl) Users(cfgStr entity.UsersPaginationStr, ctx context.Context) ([]entity.User, hypermedia.Links, error) {

	cfg, err := cfgStr.Parse()
	if err != nil {
		return nil, nil, err
	}

	users, pages, err := as.repo.Users(cfg.PageNumber, cfg.PageSize, cfg.Orderby, cfg.Roles, ctx)
	if err != nil {
		return nil, nil, err
	}

	var pagesUrls = make(hypermedia.Links, pages)
	for i := 0; i < pages; i++ {
		pagesUrls[i] = hypermedia.Link{strconv.Itoa(i + 1): hypermedia.Href{config.APIURL() + fmt.Sprintf("/admin/users/%d/%d", i+1, cfg.PageSize), http.MethodGet}}
	}

	return users, pagesUrls, nil
}

// Declaration function
func NewAdminServiceImpl(repo repo.AdminRepo, client *http.Client) AdminService {
	return &adminServiceImpl{newUserService(auth.Admin, repo), repo, client}
}
