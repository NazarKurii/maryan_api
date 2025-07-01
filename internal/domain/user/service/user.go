package service

import (
	"context"
	"maryan_api/internal/domain/user/repo"
	"maryan_api/internal/entity"
	"maryan_api/pkg/auth"
	rfc7807 "maryan_api/pkg/problem"
	"maryan_api/pkg/security"

	"github.com/asaskevich/govalidator"
	"github.com/google/uuid"
)

type UserService interface {
	//----------Not authentificated------------------
	Login(email, password string, ctx context.Context) (string, error)
	LoginJWT(id uuid.UUID, email string, ctx context.Context) (string, error)
	GetByID(id uuid.UUID, ctx context.Context) (entity.UserSimplified, error)
	//------------Authentificated--------------------

	//Aditional functionality
	//UserService() UserService
	SecretKey() []byte
	Role() auth.Role
}

type userServiceImpl struct {
	repo repo.UserRepo
	role auth.Role
}

func (us *userServiceImpl) SecretKey() []byte {
	return us.role.SecretKey()
}

func (us *userServiceImpl) Role() auth.Role {
	return us.role
}

func (us *userServiceImpl) Login(email, password string, ctx context.Context) (string, error) {
	if !govalidator.IsEmail(email) {
		return "", rfc7807.BadRequest(
			"email-invalid",
			"Invalid Email Error",
			"Provided email contains forbidden characters or is not an email at all.",
		)
	}

	id, passwordHashed, err := us.repo.Login(email, ctx)
	if err != nil {
		return "", err
	}

	if ok := security.VerifyPassword(password, passwordHashed); !ok {
		return "", rfc7807.Unauthorized(
			"invalid-password",
			"Invalid Password Error",
			"Invalid password for user assosiated with the provided email.",
		)
	}

	return us.role.GenerateToken(email, id)

}

func (us *userServiceImpl) LoginJWT(id uuid.UUID, email string, ctx context.Context) (string, error) {
	if !govalidator.IsEmail(email) {
		return "", rfc7807.BadRequest(
			"email-invalid",
			"Invalid Email Error",
			"Provided email contains forbidden characters or is not an email at all.",
		)
	}

	validID, exists, err := us.repo.UserExists(email, ctx)
	if err != nil {
		return "", err
	}

	if !exists {
		return "", rfc7807.BadRequest(
			"non-existing-user",
			"Non-existing User Error",
			"There is no user assosiated with the provided email.",
		)
	}

	if id != validID {
		return "", rfc7807.Unauthorized(
			"unauthorized",
			"Unauthorized",
			"Invalid token.",
		)
	}

	token, err := us.role.GenerateToken(email, id)

	return token, err
}

func (us *userServiceImpl) GetByID(id uuid.UUID, ctx context.Context) (entity.UserSimplified, error) {
	user, err := us.repo.GetByID(id, ctx)
	if err != nil {
		return entity.UserSimplified{}, err
	}

	return user.ToSimplified(), nil
}

func newUserService(role auth.Role, repo repo.UserRepo) UserService {
	return &userServiceImpl{repo, role}
}
