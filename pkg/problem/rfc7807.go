package rfc7807

import (
	"encoding/json"
	"fmt"
	"maryan_api/config"

	"net/http"
)

type Problem struct {
	Type          string        `json:"type"`
	Title         string        `json:"title"`
	Status        int           `json:"status"`
	Detail        string        `json:"detail"`
	InvalidParams InvalidParams `json:"invalidParams,omitempty"`
}

func (p Problem) Error() string {
	json, _ := json.Marshal(p)
	return string(json)
}

func Is(err error) (Problem, bool) {
	rfc8707, ok := err.(Problem)
	return rfc8707, ok
}

type InvalidParam struct {
	Name   string `json:"name"`
	Reason string `json:"reason"`
}

type InvalidParams []InvalidParam

func (ip *InvalidParams) SetInvalidParam(name, reason string) {
	*ip = append(*ip, InvalidParam{name, reason})
}

func New(status int, problemType, title, detail string, invalidParams ...InvalidParam) Problem {
	p, ok := validate(status, problemType, title, detail)

	if ok {
		p = Problem{
			Type:          config.APIURL() + "/problems/" + problemType,
			Title:         title,
			Status:        status,
			Detail:        detail,
			InvalidParams: invalidParams,
		}
	}

	return p
}

func validate(status int, problemType, title, detail string, invalidParams ...InvalidParam) (Problem, bool) {
	var p Problem
	var ok = true

	if http.StatusText(status) == "" {
		p.InvalidParams.SetInvalidParam("status", fmt.Sprintf(`Invalid http response status: "%v".`, status))
		ok = false
	}
	if problemType == "" {
		p.InvalidParams.SetInvalidParam("type", fmt.Sprintf("Invalid problem type, got empty string."))
		ok = false
	}
	if title == "" {
		p.InvalidParams.SetInvalidParam("title", fmt.Sprintf("Invalid problem title, got empty string."))
		ok = false
	}
	if detail == "" {
		p.InvalidParams.SetInvalidParam("detail", fmt.Sprintf("Invalid problem detail, got empty string."))
		ok = false
	}
	if len(invalidParams) != 0 {
		for _, param := range invalidParams {
			if param.Name == "" {
				ok = false
				p.InvalidParams.SetInvalidParam("param", "Empty invalidParam name")
			}
			if param.Reason == "" {
				ok = false
				p.InvalidParams.SetInvalidParam("param", "Empty invalidParam reason")
			}
		}
	}

	return p, ok
}

// func (p Problem) isComposing() bool {
// 	return p.Title == "Problem Composing Error" &&
// 		p.Detail == "Could not compose the error due to invalid params." &&
// 		p.Status == http.StatusInternalServerError &&
// 		p.Type == "internal-server-error"
// }

func composing() Problem {
	return Internal("Problem Composing Error", "Could not compose the error due to invalid params.")
}
func Internal(title, detail string) Problem {
	return New(http.StatusInternalServerError, "internal-server-error", title, detail)

}

func BadGateway(problemType, title, detail string) Problem {
	return New(http.StatusBadGateway, problemType, title, detail)
}

func BadRequest(problemType, title, detail string, invalidParams ...InvalidParam) Problem {
	return New(http.StatusBadRequest, problemType, title, detail, invalidParams...)
}

func Unauthorized(problemType, title, detail string, invalidParams ...InvalidParam) Problem {
	return New(http.StatusUnauthorized, problemType, title, detail, invalidParams...)
}

func Forbidden(problemType, title, detail string) Problem {
	return New(http.StatusForbidden, problemType, title, detail)
}

func DB(detail string) Problem {
	return BadGateway("database", "Database Error", detail)
}

func UUID(detail string) error {
	return BadRequest("invalid-id-format", "Invalid ID Format", detail)
}

type Extensions map[string]any
