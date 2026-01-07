package ginutil

import "github.com/nazarkurii/marshrutka_api/pkg/hypermedia"

type Response struct {
	Message string           `json:"message"`
	Links   hypermedia.Links `json:"links,omitempty"`
}
