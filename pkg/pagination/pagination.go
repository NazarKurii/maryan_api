package pagination

import (
	"errors"
	"fmt"
	"maryan_api/config"
	"maryan_api/pkg/hypermedia"
	rfc7807 "maryan_api/pkg/problem"
	"net/http"
	"slices"
	"strconv"
)

type CfgStr struct {
	Page     string
	Size     string
	OrderBy  string
	OrderWay string
}

type Cfg struct {
	Page  int
	Size  int
	Order string
}

func (cfgStr CfgStr) Parse(orderBy ...string) (Cfg, error) {
	cfg, params := cfgStr.parseWithParams(orderBy...)
	if params != nil {
		return Cfg{}, rfc7807.BadRequest("invalid-pagination-data", "Invalid Pagination Data Error", "Provided data is invald.", params...)
	}
	return cfg, nil
}

func (cfgStr CfgStr) parseWithParams(orderBy ...string) (Cfg, rfc7807.InvalidParams) {
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

	var cfg Cfg

	stringToInt(cfgStr.Page, "pageNumber", &cfg.Page)
	stringToInt(cfgStr.Size, "pageSize", &cfg.Size)

	if len(orderBy) == 0 {
		cfg.Order = "1"
	} else if slices.Index(orderBy, cfgStr.OrderBy) != -1 {
		cfg.Order += cfgStr.OrderBy
	} else {
		params.SetInvalidParam("orderBy", "non-existing orderBy param.")
	}

	switch cfgStr.OrderWay {
	case "DESC", "ASC":
		cfg.Order += " " + cfgStr.OrderWay
	default:
		params.SetInvalidParam("orderWay", "non-existing orderWay param.")
	}

	return cfg, params
}

func Links(total, pageSize int, path string) hypermedia.Links {
	var pagesUrls = make(hypermedia.Links, total)
	for i := 0; i < total; i++ {
		pagesUrls[i] = hypermedia.Link{strconv.Itoa(i + 1): hypermedia.Href{config.APIURL() + path + fmt.Sprintf("/%d/%d", i+1, pageSize), http.MethodGet}}
	}
	return pagesUrls
}

type CfgCondition struct {
	Cfg
	Condition
}

type Condition struct {
	Where  string
	Values []any
}

func (cfgStr CfgStr) ParseWithCondition(condition Condition, orderBy ...string) (CfgCondition, error) {
	var cfgCon CfgCondition
	var params rfc7807.InvalidParams
	cfgCon.Cfg, params = cfgStr.parseWithParams(orderBy...)

	valuesLength := len(condition.Values)

	if condition.Where == "" {
		params.SetInvalidParam("Condition.Where", "Empty condition.")
	}

	if valuesLength == 0 {
		params.SetInvalidParam("Condition.Values", "No values have been provided.")
	}

	if params != nil {
		return CfgCondition{}, rfc7807.BadRequest("invalid-pagination-data", "Invalid Pagination Data Error", "Provided data is invald.", params...)
	}

	cfgCon.Condition = condition
	return cfgCon, nil
}
