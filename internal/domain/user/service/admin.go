package service

import (
	"context"
	"fmt"
	"maryan_api/config"
	"maryan_api/internal/domain/user/repo"
	"maryan_api/internal/entity"
	"maryan_api/pkg/auth"
	"maryan_api/pkg/dbutil"
	"maryan_api/pkg/hypermedia"
	"maryan_api/pkg/images"
	"maryan_api/pkg/timeutil"
	"slices"
	"time"

	rfc7807 "maryan_api/pkg/problem"
	"mime/multipart"
	"net/http"

	"github.com/google/uuid"
)

type AdminService interface {
	UserService
	NewEmployee(ctx context.Context, ru entity.RegistrantionEmployee, image *multipart.FileHeader, role auth.Role) (string, error)
	GetUsers(ctx context.Context, paginationStr dbutil.PaginationStr, rolesStr string) ([]entity.UserSimplified, hypermedia.Links, error)
	SetEmployeeAvailability(ctx context.Context, availability []entity.EmployeeAvailability) error
	GetUserByID(ctx context.Context, id string) (entity.User, error)
	GetAvailableEmployees(ctx context.Context, paginationStr dbutil.PaginationStr, rolesStr, from, to string) ([]entity.UserSimplified, hypermedia.Links, error)
}

type adminServiceImpl struct {
	UserService
	repo   repo.AdminRepo
	client *http.Client
}

func (as adminServiceImpl) GetAvailableEmployees(ctx context.Context, paginationStr dbutil.PaginationStr, rolesStr, fromStr, toStr string) ([]entity.UserSimplified, hypermedia.Links, error) {
	roles, err := auth.SplitIntoRoles(rolesStr)
	if err != nil {
		return nil, nil, err
	}

	if slices.Contains(roles, "Customer") {
		return nil, nil, rfc7807.BadRequest("invalid-role", "Invalid Role Error", fmt.Sprintf("'Customer' role is not allowed."))
	}

	pagination, err := paginationStr.ParseWithCondition(
		dbutil.Condition{
			Where:  "role IN (?)",
			Values: []any{roles},
		},
		[]string{"first_name", "last_name", "email", "phone_number"},
		"first_name", "last_name", "email", "date_of_birth",
	)
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

	customers, total, err := as.repo.GetAvailableUsers(ctx, timeutil.FromTo(from, to), pagination)
	if err != nil {
		return nil, nil, err
	}

	return entity.SimplifyUsers(customers), hypermedia.Pagination(paginationStr, total, hypermedia.DefaultParam{
		Name:    "roles",
		Default: "admin+driver+support",
		Value:   rolesStr,
	}, hypermedia.DefaultParam{
		"from",
		"",
		fromStr,
	}, hypermedia.DefaultParam{
		"to",
		"",
		toStr,
	}), nil
}

func (as adminServiceImpl) GetUsers(ctx context.Context, paginationStr dbutil.PaginationStr, rolesStr string) ([]entity.UserSimplified, hypermedia.Links, error) {
	roles, err := auth.SplitIntoRoles(rolesStr)
	if err != nil {
		return nil, nil, err
	}

	pagination, err := paginationStr.ParseWithCondition(
		dbutil.Condition{
			Where:  "role = IN (?)",
			Values: []any{roles},
		},
		[]string{"first_name", "last_name", "email", "phone_number"},
		"first_name", "last_name", "email", "date_of_birth",
	)

	if err != nil {
		return nil, nil, err
	}

	customers, total, err := as.repo.Users(ctx, pagination)
	if err != nil {
		return nil, nil, err
	}

	return entity.SimplifyUsers(customers), hypermedia.Pagination(paginationStr, total, hypermedia.DefaultParam{
		Name:    "roles",
		Default: "admin+driver+support",
		Value:   rolesStr,
	}), nil
}

func (as *adminServiceImpl) NewEmployee(ctx context.Context, ru entity.RegistrantionEmployee, image *multipart.FileHeader, role auth.Role) (string, error) {
	user, starts := ru.ToUser(role)
	availability, err := user.PrepareNewEmployee(starts)

	if err != nil {
		return "", err
	}

	if image != nil {
		err := images.Save("../../../../static/imgs/"+user.ID.String(), image)
		if err != nil {
			return "", rfc7807.Internal("image-saving-error", err.Error())
		}
		user.ImageUrl = config.APIURL() + "/imgs/" + user.ID.String()
	} else {
		user.ImageUrl = config.APIURL() + "/imgs/guest-female.png"
	}

	err = as.repo.NewUser(ctx, &user)
	if err != nil {
		return "", err
	}

	err = as.repo.SetEmployeeAvailability(ctx, availability)
	if err != nil {
		return "", err
	}

	return user.Role.Val.GenerateToken(user.Email, user.ID)
}

func (as *adminServiceImpl) SetEmployeeAvailability(ctx context.Context, schedule []entity.EmployeeAvailability) error {
	var invalidParams rfc7807.InvalidParams
	userID := schedule[0].UserID
	for _, availability := range schedule {
		if !availability.Status.IsValid() {
			invalidParams.SetInvalidParam(availability.Date.String(), "invalid status.")
		}

		if userID != availability.UserID {
			invalidParams.SetInvalidParam(availability.Date.String(), "UserID differs.")
		}
	}

	if invalidParams != nil {
		return rfc7807.BadRequest("invalid-employee-schedule", "Invalid Employee Schedule Error", "Provided params are not valid.", invalidParams...)
	}

	exists, err := as.repo.UserExists(ctx, userID)
	if err != nil {
		return err
	}

	if !exists {
		return rfc7807.BadRequest("non-existing-user", "Non-existring User Error", "There is no user assosiated with provided id.")
	}

	return as.repo.SetEmployeeAvailability(ctx, schedule)
}

func (as *adminServiceImpl) GetUserByID(ctx context.Context, idStr string) (entity.User, error) {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return entity.User{}, rfc7807.BadRequest("invalid-id", "Invalid ID Error", err.Error())
	}
	return as.UserService.GetByID(ctx, id)
}

// Constructor function
func NewAdminServiceImpl(repo repo.AdminRepo, client *http.Client) AdminService {
	return &adminServiceImpl{
		UserService: NewUserService(auth.Admin, repo),
		repo:        repo,
		client:      client,
	}
}
