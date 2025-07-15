package hypermedia

import (
	"encoding/json"
)

type LinkData struct {
	Href   string `json:"href"`
	Method string `json:"method"`
}

type Link struct {
	Name string
	Data LinkData
}

type Links []Link

func (l Links) MarshalJSON() ([]byte, error) {
	var links = map[string]LinkData{}
	for _, link := range l {
		links[link.Name] = link.Data
	}

	return json.Marshal(links)
}

func (l *Links) Add(link Link) {
	*l = append(*l, link)
}
