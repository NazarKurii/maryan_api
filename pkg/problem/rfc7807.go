package rfc7807

import (
	"encoding/json"
	"fmt"
	"maryan_api/config"

	"net/http"
)

type Default struct {
	Type     string `json:"type"`
	Title    string `json:"title"`
	Status   int    `json:"status"`
	Detail   string `json:"detail"`
	Instance string `json:"instance,omitemty"`
}

type Problem struct {
	Default
	Extensions Extensions `json:"-,omitempty"`
}

func (err Problem) Error() string {
	return err.Detail
}

func Is(err error) (Problem, bool) {
	rfc8707, ok := err.(Problem)
	return rfc8707, ok
}

type InvalidParam struct {
	Name   string `json:"name"`
	Reason string `json:"reason"`
}

func (e *Problem) SetInvalidParam(name, reason string) {
	params, ok := e.Extensions["invalidParams"]

	if !ok {
		e.Extensions["invalidParams"] = []InvalidParam{{name, reason}}
	}

	if params, ok := params.([]InvalidParam); ok {
		e.Extensions["invalidParams"] = append(params, InvalidParam{name, reason})
	} else {
		e.Extensions["invalidParams"] = []InvalidParam{{name, reason}}
	}
}

func (e Problem) SetInvalidParams(paramsToSet []InvalidParam) Problem {
	if len(paramsToSet) == 0 {
		return e
	}

	params, ok := e.Extensions["invalidParams"]

	if !ok {
		e.Extensions["invalidParams"] = paramsToSet
	}

	if params, ok := params.([]InvalidParam); ok {
		e.Extensions["invalidParams"] = append(params, paramsToSet...)
	} else {
		e.Extensions["invalidParams"] = paramsToSet
	}
	return e
}

func StartSettingInvalidParams() (*[]InvalidParam, func(name, reason string), func() bool) {
	var invalidParams []InvalidParam
	var isNil = true
	return &invalidParams, func(name, reason string) {
		invalidParams = append(invalidParams, InvalidParam{name, reason})
		isNil = false
	}, func() bool { return isNil }
}

func New(status int, problemType, title, detail string, instance ...string) Problem {
	p, ok := validate(status, problemType, title, detail)

	if ok {
		p = Problem{
			Default: Default{
				Type:   config.APIURL() + "/problems/" + problemType,
				Title:  title,
				Status: status,
				Detail: detail,
			},
			Extensions: make(Extensions),
		}
	}

	if len(instance) > 0 {
		p.SetInstance(instance[0])
	}

	return p
}

func validate(status int, problemType, title, detail string) (Problem, bool) {
	var p Problem
	var ok = true

	if http.StatusText(status) == "" {
		p.SetInvalidParam("status", fmt.Sprintf(`Invalid http response status: "%v".`, status))
		ok = false
	}
	if problemType == "" {
		p.SetInvalidParam("type", fmt.Sprintf("Invalid problem type, got empty string."))
		ok = false
	}
	if title == "" {
		p.SetInvalidParam("title", fmt.Sprintf("Invalid problem title, got empty string."))
		ok = false
	}
	if detail == "" {
		p.SetInvalidParam("detail", fmt.Sprintf("Invalid problem detail, got empty string."))
		ok = false
	}

	return p, ok
}

func (p *Problem) SetInstance(instance string) {
	if length := len(instance); length != 36 {
		p.SetInvalidParam("instance", fmt.Sprintf(`Invalid instance format. Want 36-char string, got "%s(%v)"`, instance, length))
	} else {
		p.Instance = config.APIURL() + "/instances/" + instance
	}
}

func (p Problem) isComposing() bool {
	return p.Title == "Problem Composing Error" &&
		p.Detail == "Could not compose the error due to invalid params." &&
		p.Status == http.StatusInternalServerError &&
		p.Type == "internal-server-error"
}

func composing() Problem {
	return Internal("Problem Composing Error", "Could not compose the error due to invalid params.")
}
func Internal(title, detail string, instance ...string) Problem {
	return New(http.StatusInternalServerError, "internal-server-error", title, detail, instance...)

}

func BadGateway(problemType, title, detail string, instance ...string) Problem {
	return New(http.StatusBadGateway, problemType, title, detail, instance...)
}

func BadRequest(problemType, title, detail string, instance ...string) Problem {
	return New(http.StatusBadRequest, problemType, title, detail, instance...)
}

func Unauthorized(problemType, title, detail string, instance ...string) Problem {
	return New(http.StatusUnauthorized, problemType, title, detail, instance...)
}

func Forbidden(problemType, title, detail string, instance ...string) Problem {
	return New(http.StatusForbidden, problemType, title, detail, instance...)
}

func DB(detail string) Problem {
	return BadGateway("database", "Database Error", detail)
}

type Extensions map[string]any

func (e Problem) MarshalExtensions() []byte {
	if len(e.Extensions) == 0 {
		return nil
	}

	extsJSON, err := json.Marshal(e.Extensions)

	if err != nil {
		e.Extensions = map[string]any{"ExtensionsError": fmt.Errorf(`Error while parsing Extensions: %s`, err.Error())}
		extsJSON, _ := json.Marshal(e.Extensions)
		return extsJSON
	}

	return extsJSON
}

func (e *Problem) Extend(name string, value any) {
	e.Extensions[name] = value
}

func (e Problem) MarshalJSON() ([]byte, error) {
	baseJSON, err := json.Marshal(e.Default)

	if err != nil {
		return nil, err
	}

	extsJSON := e.MarshalExtensions()
	if extsJSON == nil {
		return baseJSON, nil
	}
	return append(baseJSON[:len(baseJSON)-1], append([]byte(","), extsJSON[1:]...)...), nil
}
