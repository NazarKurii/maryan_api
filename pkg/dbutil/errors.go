package dbutil

import (
	"errors"
	rfc7807 "maryan_api/pkg/problem"

	"gorm.io/gorm"
)

func defineError(err error, errType string) error {
	switch {
	case errors.Is(err, gorm.ErrInvalidData):
		return rfc7807.BadRequest(errType, "Invalid data", err.Error())

	case errors.Is(err, gorm.ErrPrimaryKeyRequired):
		return rfc7807.BadRequest(errType, "Primary key required", err.Error())

	case errors.Is(err, gorm.ErrInvalidField):
		return rfc7807.BadRequest(errType, "Invalid field in request", err.Error())

	case errors.Is(err, gorm.ErrModelValueRequired):
		return rfc7807.BadRequest(errType, "Model value required", err.Error())

	case errors.Is(err, gorm.ErrEmptySlice):
		return rfc7807.BadRequest(errType, "Empty slice provided", err.Error())

	default:
		return rfc7807.DB(err.Error())
	}
}

func PossibleRawsAffectedError(result *gorm.DB, errType string) error {
	if result.RowsAffected == 0 {
		return rfc7807.BadRequest(
			errType,
			"Non-existing Resourse Error",
			"There is no Resourse assosiated with provided data.",
		)
	}

	if result.Error != nil {
		return defineError(result.Error, errType)
	}

	return nil
}

func PossibleDbError(result *gorm.DB) error {

	if result.Error != nil {
		return rfc7807.DB(result.Error.Error())
	}

	return nil
}

func PossibleFirstError(result *gorm.DB, errType string) error {
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return rfc7807.BadRequest(
				errType,
				"Non-existing Resourse Error",
				"There is no Resourse assosiated with provided data.",
			)
		}
		return defineError(result.Error, errType)
	}

	return nil
}

func PossibleForeignKeyError(result *gorm.DB, nonExistingParentErrType, nonExistingChildErrType, errType string) error {
	if result.Error == nil {
		return nil
	}

	if errors.Is(result.Error, gorm.ErrForeignKeyViolated) {
		return rfc7807.BadRequest(nonExistingChildErrType, "Non-existing Resourse Error", "There is no resourse assosiated with provided data.")
	} else if result.RowsAffected == 0 {
		return rfc7807.BadRequest(
			nonExistingParentErrType,
			"Non-existing Resourse Error",
			"There is no Resourse assosiated with provided data.",
		)
	}

	return defineError(result.Error, errType)
}

func PossibleForeignKeyCreateError(result *gorm.DB, nonExistingParentErrType, errType string) error {
	if result.Error == nil {
		return nil
	}

	if errors.Is(result.Error, gorm.ErrForeignKeyViolated) {
		return rfc7807.BadRequest(nonExistingParentErrType, "Non-existing Resourse Error", "There is no resourse assosiated with provided data.")
	}

	return defineError(result.Error, errType)
}

func PossibleCreateError(result *gorm.DB, errType string) error {
	if result.Error != nil {
		return defineError(result.Error, errType)
	}
	return nil
}

func Preload(query *gorm.DB, preload ...string) *gorm.DB {
	for _, v := range preload {
		query.Preload(v)
	}
	return query
}
