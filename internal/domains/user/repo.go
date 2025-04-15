package user

import (
	"errors"
	"maryan_api/config"
	rfc7807 "maryan_api/pkg/problem"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// userRepo defines basic login methods and the getByID method, that available for all users.
type userRepo interface {
	database() *gorm.DB
	repo() *userRepoMySQL
	getByID(id uuid.UUID) (User, error)
	login(email string) (uuid.UUID, string, error)
	loginJWT(email string, id uuid.UUID) (bool, error)
}

// customerRepo embeds userRepo and defines additional Customer functionality.
type customerRepo interface {
	userRepo
	save(user *User) error
	delete(id uuid.UUID) (bool, error)
	emailExists(email string) (bool, error)
	userExists(email string) (uuid.UUID, bool, error)
}

// adminrRepo embeds userRepo and defines additional Customer functionality.
type adminRepo interface {
	userRepo
	getUsers(pageNumber, pageSize int) ([]User, int64, error)
}

// driverRepo only embeds userRepo methods, so basicly is an allies for better role structure.
type driverRepo interface {
	userRepo
}

// driverRepo only embeds userRepo methods, so basicly is an allies for better role structure.
type supportEmployeeRepo interface {
	userRepo
}

// MYSQL IMPLEMENTATION
type userRepoMySQL struct {
	db *gorm.DB
}

func (mscr *userRepoMySQL) repo() *userRepoMySQL {
	return mscr
}

func (mscr *userRepoMySQL) database() *gorm.DB {
	return mscr.db
}

func (mscr *userRepoMySQL) getByID(id uuid.UUID) (User, error) {
	var u = User{ID: id}
	result := mscr.db.First(&u)

	if result.Error != nil {
		return User{}, rfc7807.DB(result.Error.Error())
	}

	if result.RowsAffected == 0 {
		return User{}, rfc7807.BadRequest(
			"unexistant-user",
			"Unexistant user Error",
			"There is no user assosiated with provided id.",
		)
	}

	return u, nil
}

func (mscr *userRepoMySQL) login(email string) (uuid.UUID, string, error) {
	var user User

	result := mscr.db.Select("id", "password").Where("email = ?", email).Find(&user)
	if result.Error != nil {
		return uuid.Nil, "", rfc7807.DB(result.Error.Error())
	}

	if result.RowsAffected == 0 {
		return uuid.Nil, "", rfc7807.BadRequest(
			"unexistant-user",
			"Invalid user Email",
			"There is no user assosiated with provided email",
		)
	}

	return user.ID, user.Password, nil
}

func (mscr *userRepoMySQL) loginJWT(email string, id uuid.UUID) (bool, error) {
	var exists bool
	err := mscr.db.Raw("SELECT EXISTS(SELECT 1 FROM users WHERe email = ? AND id = ?)", email, id).Scan(&exists).Error
	if err != nil {
		return false, rfc7807.DB(err.Error())
	}

	return exists, nil
}

type customerRepoMySQL struct {
	userRepoMySQL
}

func (mscr *customerRepoMySQL) save(u *User) error {
	var imageUrl string
	err := mscr.db.Raw("SELECT image_url	A FROM users WHERE id = ?", u.ID).Scan(&imageUrl).Error
	if err != nil {
		return rfc7807.DB(err.Error())
	}

	if imageUrl == "" {
		imageUrl = config.APIURL() + "/images/default-avatar.jpg"
	}

	u.ImageUrl = imageUrl

	err = mscr.db.Save(u).Error
	if err != nil {
		if errors.Is(gorm.ErrInvalidData, err) {
			return rfc7807.DB(err.Error())
		}
		return rfc7807.BadRequest("user-credentials-validation", "Invalid user Credentials", err.Error())
	}

	return nil
}

func (mscr *customerRepoMySQL) delete(id uuid.UUID) (bool, error) {
	result := mscr.db.Delete(&User{ID: id})
	if result.RowsAffected == 0 {
		return false, rfc7807.BadRequest("unexistant-user", "Unexistant user Error", "There is no user with provided id")
	}

	if result.Error != nil {
		return false, rfc7807.DB(result.Error.Error())
	}

	return true, nil
}

func (mscr *customerRepoMySQL) emailExists(email string) (bool, error) {
	var exists bool
	err := mscr.db.Raw("SELECT EXISTS(SELECT 1 FROM users WHERE email = ?) ", email).Scan(&exists).Error
	if err != nil {
		return false, rfc7807.DB(err.Error())
	}

	return exists, nil
}

func (mscr *customerRepoMySQL) userExists(email string) (uuid.UUID, bool, error) {
	var user User
	err := mscr.db.Select("id").Where("email = ?", email).Find(&user).Error
	if err != nil {
		return uuid.Nil, false, rfc7807.DB(err.Error())
	}

	return user.ID, user.ID != uuid.Nil, nil
}

type adminRepoMySQL struct {
	userRepoMySQL
}

func (mscr *adminRepoMySQL) getUsers(pageNumber, pageSize int) ([]User, int64, error) {
	var users []User
	var totalUsers int64

	err := mscr.db.Offset(int(pageNumber) * pageSize).Limit(pageSize).Find(&users).Error
	if err != nil {
		return nil, 0, rfc7807.DB(err.Error())
	}

	if len(users) == 0 {
		return nil, 0, rfc7807.BadRequest(
			"users-order-data",
			"user Order Data Error",
			"pageNumber is invalid (too big)",
		)
	}

	err = mscr.db.Model(&User{}).Count(&totalUsers).Error
	if err != nil {
		return nil, 0, rfc7807.DB(err.Error())
	}

	return users, totalUsers / int64(pageSize), nil
}

type driverRepoMySQL struct {
	userRepoMySQL
}

type supportEmployeeRepoMySQL struct {
	userRepoMySQL
}

//Declaration functions

func newCustomerRepoMySQL(db *gorm.DB) customerRepo {
	return &customerRepoMySQL{userRepoMySQL{db}}
}

func newAdminRepoMySQL(db *gorm.DB) adminRepo {
	return &adminRepoMySQL{userRepoMySQL{db}}
}

func newDriverRepoMySQL(db *gorm.DB) driverRepo {
	return &driverRepoMySQL{userRepoMySQL{db}}
}

func newSupportEmployeeRepoMySQL(db *gorm.DB) supportEmployeeRepo {
	return &supportEmployeeRepoMySQL{userRepoMySQL{db}}
}
