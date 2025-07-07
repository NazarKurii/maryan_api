package dbutil

import (
	"context"
	"errors"
	"fmt"
	"maryan_api/config"
	"maryan_api/pkg/hypermedia"
	"math"

	rfc7807 "maryan_api/pkg/problem"
	"net/http"
	"slices"
	"strconv"

	"gorm.io/gorm"
)

func Paginate[T any](ctx context.Context, db *gorm.DB, pagination Pagination, preload ...string) (entities []T, urls hypermedia.Links, err error) {
	err = PossibleRawsAffectedError(buildRequest(ctx, db, pagination, preload...).Find(&entities), "non-existing-page")
	if err != nil {
		return nil, nil, err
	}

	links, err := links[T](db, pagination.Path, pagination.Size)
	return entities, links, err
}

func PaginateWithCondition[T any](ctx context.Context, db *gorm.DB, conditionPagination CondtionPagination, preload ...string) (entities []T, urls hypermedia.Links, err error) {
	err = PossibleRawsAffectedError(
		buildRequest(ctx, db, conditionPagination.Pagination, preload...).
			Where(conditionPagination.Condition.Where, conditionPagination.Condition.Values...).
			Find(&entities), "non-existing-page",
	)

	if err != nil {
		return nil, nil, err
	}

	links, err := links[T](db, conditionPagination.Path, conditionPagination.Size)
	return entities, links, err
}

func countTotaPages[T any](db *gorm.DB, size int) (int, error) {
	var model T
	var totalInDbINT64 int64
	err := db.Model(&model).Count(&totalInDbINT64).Error
	if err != nil || totalInDbINT64 == 0 {
		return 0, rfc7807.DB("Could not count buses.")
	}
	return int(math.Ceil(float64(totalInDbINT64) / float64(size))), nil
}

func buildRequest(ctx context.Context, db *gorm.DB, pagination Pagination, preload ...string) *gorm.DB {
	pagination.Page--

	var request = db.WithContext(ctx).Limit(pagination.Size).Offset(pagination.Page * pagination.Size).Order(pagination.Order)

	if len(preload) != 0 {
		for _, v := range preload {
			request.Preload(v)
		}
	}

	return request
}

type PaginationStr struct {
	Path     string
	Page     string
	Size     string
	OrderBy  string
	OrderWay string
}

type Pagination struct {
	Path  string
	Page  int
	Size  int
	Order string
}

type CondtionPagination struct {
	Pagination
	Condition Condition
}

type Condition struct {
	Where  string
	Values []any
}

func (p PaginationStr) ParseWithCondition(condition Condition, orderBy ...string) (CondtionPagination, error) {
	pagination, params := p.parseWithParams(orderBy...)

	if condition.Where == "" {
		params.SetInvalidParam("Condition.Where", "Empty condition.")
	}

	if len(condition.Values) == 0 {
		params.SetInvalidParam("Condition.Values", "No values have been provided.")
	}

	if params != nil {
		return CondtionPagination{}, rfc7807.BadRequest("invalid, data", "Invalid Pagination Data Error", "Provided data is invald.", params...)
	}

	return CondtionPagination{pagination, condition}, nil
}

func (pStr PaginationStr) Parse(orderBy ...string) (Pagination, error) {
	pagination, params := pStr.parseWithParams(orderBy...)
	if params != nil {
		return Pagination{}, rfc7807.BadRequest("invalid, data", "Invalid Pagination Data Error", "Provided data is invald.", params...)
	}
	return pagination, nil
}

func (pStr PaginationStr) parseWithParams(orderBy ...string) (Pagination, rfc7807.InvalidParams) {
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
	case "DESC", "ASC":
		pagination.Order += " " + pStr.OrderWay
	default:
		params.SetInvalidParam("orderWay", "non-existing orderWay param.")
	}

	if pStr.Path == "" {
		params.SetInvalidParam("path", "Empty.")
	}

	return pagination, params
}

func links[T any](db *gorm.DB, path string, size int) (hypermedia.Links, error) {
	total, err := countTotaPages[T](db, size)
	if err != nil {
		return nil, err
	}

	var pagesUrls = make(hypermedia.Links, total)
	for i := 0; i < total; i++ {
		pagesUrls[i] = hypermedia.Link{strconv.Itoa(i + 1): hypermedia.Href{config.APIURL() + path + fmt.Sprintf("/%d/%d", i+1, size), http.MethodGet}}
	}
	return pagesUrls, nil
}
