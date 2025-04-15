package auth

import (
	"errors"
	rfc7807 "maryan_api/pkg/problem"
	"time"

	"github.com/d3code/uuid"
	"github.com/golang-jwt/jwt/v5"
)

func generateToken(email string, userID uuid.UUID, role Role) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email":   email,
		"userID":  userID.String(),
		"expires": time.Now().Add(role.TokenDuration()).Unix(),
		"role":    role.Name(),
	})

	signedToken, err := token.SignedString(role.SecretKey())

	if err != nil {
		return "", problem(role.Name(), err)
	}

	return signedToken, nil
}

func GenerateAccessToken(secretKey []byte, claims jwt.MapClaims) (string, error) {
	claims["expires"] = time.Now().Add(time.Minute * 10).Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString(secretKey)

	if err != nil {
		return "", rfc7807.Internal("Token Generation Error", "Could not generate access token.")
	}

	return signedToken, nil
}

type ClaimValidation struct {
	Name       string
	Returnable bool
	Type       claimType
}

type claimType int

const (
	ClaimString  claimType = 0
	ClaimFloat64 claimType = 1
	ClaimInt64   claimType = 2
	ClaimUUID    claimType = 3
)

func VerifyAccessToken(token string, secretKey []byte, claimsValidations []ClaimValidation) ([]any, error) {
	parsedToken, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("Unexpected signing method")
		}

		return secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !parsedToken.Valid {
		return nil, errors.New("Invalid token")
	}

	tokenClaims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("Invalid token")
	}

	expires, ok := tokenClaims["expires"].(float64)
	if !ok {
		return nil, errors.New("Invalid token")
	}

	if time.Unix(int64(expires), 0).Before(time.Now()) {
		return nil, errors.New("The token has expired")
	}

	var claims []any
	var value any

	for _, claimValidation := range claimsValidations {
		switch claimValidation.Type {
		case ClaimString:
			value, ok = tokenClaims[claimValidation.Name].(string)
		case ClaimInt64:
			value, ok = tokenClaims[claimValidation.Name].(int64)
		case ClaimFloat64:
			value, ok = tokenClaims[claimValidation.Name].(float64)
		case ClaimUUID:
			valueStr, exists := tokenClaims[claimValidation.Name].(string)
			if !exists {
				return nil, errors.New("Invalid token")
			}
			var err error
			value, err = uuid.Parse(valueStr)
			if err != nil {
				ok = false
			}
		default:
			return nil, errors.New("Invalid token")
		}

		if !ok {
			return nil, errors.New("Invalid token")
		}

		if claimValidation.Returnable {
			claims = append(claims, value)
		}
	}

	return claims, nil
}

func verifyUserToken(token string, secretKey []byte) (uuid.UUID, string, Role, error) {
	parsedToken, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("Unexpected signing method")
		}

		return secretKey, nil
	})

	if err != nil {
		return uuid.Nil, "", nil, err
	}

	if !parsedToken.Valid {
		return uuid.Nil, "", nil, errors.New("Invalid token")
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return uuid.Nil, "", nil, errors.New("Invalid token")
	}

	id, err := uuid.Parse(claims["userID"].(string))
	if err != nil {
		return uuid.Nil, "", nil, errors.New("Invalid token")
	}

	email, ok := claims["email"].(string)
	if !ok {
		return uuid.Nil, "", nil, errors.New("Invalid token")
	}

	expires, ok := claims["expires"].(float64)
	if !ok {
		return uuid.Nil, "", nil, errors.New("Invalid token")
	}

	roleString, ok := claims["role"].(string)
	if !ok {
		return uuid.Nil, "", nil, errors.New("Invalid token")
	}

	role, err := DefineRole(roleString)
	if err != nil {
		return uuid.Nil, "", nil, errors.New("Invalid token")
	}

	if time.Unix(int64(expires), 0).Before(time.Now()) {
		return uuid.Nil, "", nil, errors.New("The token has expired")
	}

	return id, email, role, nil
}
