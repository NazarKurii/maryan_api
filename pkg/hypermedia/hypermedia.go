package hypermedia

type Href struct {
	Href   string `json:"href"`
	Method string `json:"string"`
}

type Link map[string]Href

type Links []Link

func (l *Links) Add(name, url, method string) {
	*l = append(*l, Link{name: Href{url, method}})
}

func (l *Links) AddLink(link Link) {
	*l = append(*l, link)
}
