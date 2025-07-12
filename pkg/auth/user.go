package auth

import (
	"fmt"
	"maryan_api/config"
	rfc7807 "maryan_api/pkg/problem"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Role interface {
	Name() string
	SecretKey() []byte
	TokenDuration() time.Duration
	GenerateToken(email string, id uuid.UUID) (string, error)
}

type CustomerRole string
type AdminRole string
type DriverRole string
type SupportRole string

const (
	Customer CustomerRole = "Customer"
	Admin    AdminRole    = "Admin"
	Driver   DriverRole   = "Driver"
	Support  SupportRole  = "Support"
)

func (r CustomerRole) Name() string                 { return string(r) }
func (r CustomerRole) SecretKey() []byte            { return config.CustomerSecretKey() }
func (r CustomerRole) TokenDuration() time.Duration { return 7 * 24 * time.Hour }
func (r CustomerRole) GenerateToken(email string, id uuid.UUID) (string, error) {
	return generateToken(email, id, r)
}

func (r AdminRole) Name() string                 { return string(r) }
func (r AdminRole) SecretKey() []byte            { return config.AdminSecretKey() }
func (r AdminRole) TokenDuration() time.Duration { return 24 * time.Hour }
func (r AdminRole) GenerateToken(email string, id uuid.UUID) (string, error) {
	return generateToken(email, id, r)
}

func (r DriverRole) Name() string                 { return string(r) }
func (r DriverRole) SecretKey() []byte            { return config.DriverSecretKey() }
func (r DriverRole) TokenDuration() time.Duration { return 3 * 24 * time.Hour }
func (r DriverRole) GenerateToken(email string, id uuid.UUID) (string, error) {
	return generateToken(email, id, r)
}

func (r SupportRole) Name() string                 { return string(r) }
func (r SupportRole) SecretKey() []byte            { return config.SupportEmployeeSecretKey() }
func (r SupportRole) TokenDuration() time.Duration { return 24 * time.Hour }
func (r SupportRole) GenerateToken(email string, id uuid.UUID) (string, error) {
	return generateToken(email, id, r)
}

func DefineRole(role string) (Role, error) {
	switch role {
	case Customer.Name():
		return Customer, nil
	case Admin.Name():
		return Admin, nil
	case Driver.Name():
		return Driver, nil
	case Support.Name():
		return Support, nil
	default:
		return nil, fmt.Errorf("unknown role: %s", role)
	}
}

func SplitIntoRoles(rolesStr string) ([]string, error) {
	roles := strings.Split(rolesStr, "+")
	for i, role := range roles {
		role = strings.ToUpper(string(role[0])) + strings.ToLower(role[0:])
		if _, err := DefineRole(role); err != nil {
			return nil, rfc7807.BadRequest("non-existing-role", "Non-existing Role Error", "There is no role with provided name.")
		}

		roles[i] = role
	}

	return roles, nil
}
