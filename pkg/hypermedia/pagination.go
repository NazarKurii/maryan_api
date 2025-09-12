package hypermedia

import (
	"fmt"
	"maryan_api/config"
	"maryan_api/pkg/dbutil"
	"strconv"
)

type DefaultParam struct {
	Name    string
	Default string
	Value   string
}

func (dp DefaultParam) IsDefault() bool {
	return dp.Default == dp.Value
}

func Pagination(pagination dbutil.PaginationStr, total int, params ...DefaultParam) Links {
	var links = make(Links, 0, total)
	base := fmt.Sprintf(
		"%s%s?size=%s&order_by=%s&order_way=%s",
		config.APIURL(), pagination.Path, pagination.Size, pagination.OrderBy, pagination.OrderWay,
	)

	if pagination.Search != "" {
		base += "&search=" + pagination.Search
	}

	for page := 1; page <= total; page++ {
		pageString := strconv.Itoa(page)

		var url = base + "&page=" + pageString

		for _, param := range params {

			url += "&" + param.Name + "=" + param.Value
		}

		links.Add(Link{strconv.Itoa(page - 1), LinkData{url, "GET"}})
	}

	return links
}
