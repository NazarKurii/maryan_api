package dataStore

import (
	"context"
	"time"

	"maryan_api/internal/entity"
	"maryan_api/pkg/dbutil"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AdminDataStore interface {
	User
	Users(ctx context.Context, p dbutil.Pagination) ([]entity.User, int, error)
	NewUser(ctx context.Context, user *entity.User) error
	SetEmployeeAvailability(ctx context.Context, schedule []entity.EmployeeAvailability) error
	GetAvailableUsers(ctx context.Context, dates []time.Time, pagination dbutil.Pagination) ([]entity.User, int, error)
	IsDriverAvailable(ctx context.Context, dates []time.Time, driverID uuid.UUID) (bool, error)
}

type adminMySQL struct {
	userMySQL
}

func (ads *adminMySQL) Users(ctx context.Context, p dbutil.Pagination) ([]entity.User, int, error) {
	return dbutil.Paginate[entity.User](ctx, ads.db, p)
}

func (ads *adminMySQL) NewUser(ctx context.Context, user *entity.User) error {
	return dbutil.PossibleCreateError(ads.db.WithContext(ctx).Create(user), "user-credentials-validation")
}

func (ads *adminMySQL) GetAvailableUsers(ctx context.Context, dates []time.Time, pagination dbutil.Pagination) ([]entity.User, int, error) {
	return dbutil.Paginate[entity.User](ctx, ads.db.
		Table("users").
		Select("DISTINCT users.*").
		Joins("JOIN employee_availabilities ON employee_availabilities.user_id = users.id").
		Where("users.role = 'Driver' AND employee_availabilities.date NOT IN (?)", dates), pagination)
}

func (ads *adminMySQL) IsDriverAvailable(ctx context.Context, dates []time.Time, driverID uuid.UUID) (bool, error) {
	var available bool
	return available, dbutil.PossibleDbError(ads.db.WithContext(ctx).Select("SELECT EXISTS(SELECT 1 FROM  employee_availabilities WHERE driverID = ? AND date NOT IN (?))", driverID, dates).Scan(&available))
}

func (ads *adminMySQL) SetEmployeeAvailability(ctx context.Context, schedule []entity.EmployeeAvailability) error {
	return dbutil.PossibleRawsAffectedError(ads.db.WithContext(ctx).Create(schedule), "non-existing-employee")
}

// Declaration function
func NewAdmin(db *gorm.DB) AdminDataStore {
	return &adminMySQL{userMySQL{db}}
}
