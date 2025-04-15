package rfc7807

import (
	"fmt"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func (e Problem) Map() map[string]any {
	var object = map[string]any{}
	object["type"] = e.Type
	object["title"] = e.Title
	object["status"] = e.Status
	object["detail"] = e.Detail
	if e.Instance != "" {
		object["intance"] = e.Instance
	}

	for key, v := range e.Extensions {
		object[key] = v
	}

	return object
}

func Cmp(check Problem, target Problem) string {
	return cmp.Diff(check, target, cmpopts.IgnoreFields(Problem{}, "Instance"))
}

func CmpMaps(check map[string]any, target map[string]any) (string, error) {
	problemCheck, err := ProblemFromMap(check)
	if err != nil {
		return "", err
	}

	problemTarget, err := ProblemFromMap(target)
	if err != nil {
		return "", err
	}

	return Cmp(problemCheck, problemTarget), nil
}

func ProblemFromMap(m map[string]any) (Problem, error) {
	var p Problem
	p.Extensions = make(Extensions)

	requiredFields := []string{"type", "title", "status", "detail"}
	missing := []string{}

	for _, key := range requiredFields {
		if _, ok := m[key]; !ok {
			missing = append(missing, key)
		}
	}

	if len(missing) > 0 {
		return Problem{}, fmt.Errorf("missing required fields: %v", missing)
	}

	var ok bool

	if p.Type, ok = m["type"].(string); !ok {
		return Problem{}, fmt.Errorf(`field "type" must be a string`)
	}
	if p.Title, ok = m["title"].(string); !ok {
		return Problem{}, fmt.Errorf(`field "title" must be a string`)
	}
	if p.Detail, ok = m["detail"].(string); !ok {
		return Problem{}, fmt.Errorf(`field "detail" must be a string`)
	}

	switch v := m["status"].(type) {
	case int:
		p.Status = v
	case float64:
		p.Status = int(v)
	default:
		return Problem{}, fmt.Errorf(`field "status" must be a number`)
	}

	if inst, ok := m["instance"].(string); ok && inst != "" {
		p.Instance = inst
	}

	for k, v := range m {
		if k != "type" && k != "title" && k != "status" && k != "detail" && k != "instance" {
			p.Extensions[k] = v
		}
	}

	return p, nil
}
