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

func PossibleCreateError(result *gorm.DB, errType string) error {
	if result.Error != nil {
		return defineError(result.Error, errType)
	}
	return nil
}

func PossibleDeleteError(result *gorm.DB, errType string) error {
	if result.Error != nil {
		return defineError(result.Error, errType)
	}

	if result.RowsAffected == 0 {
		return rfc7807.BadRequest(
			errType,
			"Empty result set",
			"No records were found for the provided query conditions.",
		)
	}

	return nil
}
