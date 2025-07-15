package dbutil

import (
	"context"
	"errors"
	"fmt"
	"math"

	rfc7807 "maryan_api/pkg/problem"
	"slices"
	"strconv"

	"gorm.io/gorm"
)

func Paginate[T any](ctx context.Context, db *gorm.DB, pagination Pagination, preload ...string) ([]T, int, error, bool) {
	var entities []T

	err := PossibleDbError(buildRequest(ctx, db, pagination, preload...).Find(&entities))

	if err != nil || len(entities) == 0 {

		return nil, 0, err, true
	}

	totalPages, err := countTotaPages[T](ctx, db, pagination.Size, pagination)
	if err != nil {
		return nil, 0, err, true
	}

	return entities, totalPages, nil, false
}

func countTotaPages[T any](ctx context.Context, db *gorm.DB, size int, pagination Pagination) (int, error) {
	var model T
	var totalInDbINT64 int64
	err := db.WithContext(ctx).Where(pagination.Condition.Where, pagination.Condition.Values...).Model(&model).Count(&totalInDbINT64).Error
	if err != nil {
		return 0, rfc7807.DB(err.Error())
	}
	return int(math.Ceil(float64(totalInDbINT64) / float64(size))), nil
}

func buildRequest(ctx context.Context, db *gorm.DB, pagination Pagination, preload ...string) *gorm.DB {
	pagination.Page--
	var request = db.WithContext(ctx).Limit(pagination.Size).Offset(pagination.Page*pagination.Size).Order("id, "+pagination.Order).Where(pagination.Condition.Where, pagination.Condition.Values...)

	for _, v := range preload {
		request = request.Preload(v)
	}

	return request
}

type PaginationStr struct {
	Path     string
	Page     string
	Size     string
	OrderBy  string
	OrderWay string
	Search   string
}

type Pagination struct {
	Path      string
	Page      int
	Size      int
	Order     string
	Condition Condition
}

type Condition struct {
	Where  string
	Values []any
}

func (p PaginationStr) ParseWithCondition(condition Condition, search []string, orderBy ...string) (Pagination, error) {
	pagination, params := p.parseWithParams(search, orderBy...)

	if condition.Where == "" {
		params.SetInvalidParam("Condition.Where", "Empty condition.")
	}

	if len(condition.Values) == 0 {
		params.SetInvalidParam("Condition.Values", "No values have been provided.")
	}

	if params != nil {
		return Pagination{}, rfc7807.BadRequest("invalid, data", "Invalid Pagination Data Error", "Provided data is invald.", params...)
	}

	if p.Search != "" {
		pagination.Condition.Where += "AND "
	}

	pagination.Condition.Where += condition.Where
	pagination.Condition.Values = append(pagination.Condition.Values, condition.Values...)

	return pagination, nil
}

func (pStr PaginationStr) Parse(search []string, orderBy ...string) (Pagination, error) {
	pagination, params := pStr.parseWithParams(search, orderBy...)
	if params != nil {
		return Pagination{}, rfc7807.BadRequest("invalid, data", "Invalid Pagination Data Error", "Provided data is invald.", params...)
	}
	return pagination, nil
}

func (pStr PaginationStr) parseWithParams(search []string, orderBy ...string) (Pagination, rfc7807.InvalidParams) {
	var params rfc7807.InvalidParams
	var err error
	stringToInt := func(s string, name string, destination *int) {
		*destination, err = strconv.Atoi(s)
		if err != nil {
			if errors.Is(err, strconv.ErrSyntax) {
				params.SetInvalidParam(name, err.Error())
			} else {

			}
		} else if *destination < 1 {
			params.SetInvalidParam(name, "Must be equal or greater than 1.")
		}
	}

	var pagination Pagination

	stringToInt(pStr.Page, "pageNumber", &pagination.Page)
	stringToInt(pStr.Size, "pageSize", &pagination.Size)

	if len(orderBy) == 0 {
		pagination.Order = "1"
	} else if slices.Index(orderBy, pStr.OrderBy) != -1 {
		pagination.Order += pStr.OrderBy
	} else {
		params.SetInvalidParam("orderBy", "non-existing orderBy param.")
	}

	switch pStr.OrderWay {
	case "desc":
		pagination.Order += " DESC"
	case "asc":
		pagination.Order += " ASC"
	default:
		fmt.Println(pStr.OrderWay)
		params.SetInvalidParam("orderWay", "non-existing orderWay param.")
	}

	if pStr.Path == "" {
		params.SetInvalidParam("path", "Empty.")
	}

	if pStr.Search != "" {
		pagination.Condition.Where = "("

		for i, v := range search {
			if i != 0 {
				v = "OR " + v
			}
			pagination.Condition.Where += fmt.Sprintf("%s LIKE CONCAT('%%', ?, '%%')", v)
			pagination.Condition.Values = append(pagination.Condition.Values, pStr.Search)
		}

		pagination.Condition.Where += ")"
	}

	return pagination, params
}
