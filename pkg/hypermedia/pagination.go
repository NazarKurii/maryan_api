package hypermedia

import (
	"maryan_api/config"
	"strconv"
)

func PaginationParams(base string, total int, params ...string) Links {
	var links = make(Links, total)
	base = config.APIURL() + base + "/"

	for page := 1; page <= total; page++ {
		pageString := strconv.Itoa(page)

		var url = base + pageString

		for _, param := range params {
			url += "/" + param
		}

		links[page-1] = Link{pageString: Href{url, "GET"}}
	}

	return links
}

type DefaultParam struct {
	Name    string
	Default string
	Value   string
}

func (dp DefaultParam) IsDefault() bool {
	return dp.Default == dp.Value
}

func PaginationDefaultParams(base string, total int, params []DefaultParam) Links {
	var links = make(Links, total)
	base = config.APIURL() + base + "?"

	for page := 1; page <= total; page++ {
		pageString := strconv.Itoa(page)

		var url = base + pageString

		for i, param := range params {
			if i != 0 {
				url += "&"
			}
			url += param.Name + "=" + param.Value
		}

		links[page-1] = Link{pageString: Href{url, "GET"}}
	}
	return links
}
