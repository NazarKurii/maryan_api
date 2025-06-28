package auth

import (
	"errors"
	rfc7807 "maryan_api/pkg/problem"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
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

func GenerateAccessToken(id uuid.UUID, secretKey []byte, duration time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":      id.String(),
		"expires": time.Now().Add(duration).Unix(),
	})

	signedToken, err := token.SignedString(secretKey)

	if err != nil {
		return "", rfc7807.DB("Could not generate access token.")
	}

	return signedToken, nil
}

func VerifyAccessToken(token string, secretKey []byte) (uuid.UUID, error) {
	parsedToken, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("Unexpected signing method")
		}

		return secretKey, nil
	})

	if err != nil {
		return uuid.Nil, err
	}

	if !parsedToken.Valid {
		return uuid.Nil, errors.New("Invalid token")
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return uuid.Nil, errors.New("Invalid token")
	}

	id, err := uuid.Parse(claims["id"].(string))
	if err != nil {
		return uuid.Nil, errors.New("Invalid token")
	}

	expires, ok := claims["expires"].(float64)
	if !ok {
		return uuid.Nil, errors.New("Invalid token")
	}

	if time.Unix(int64(expires), 0).Before(time.Now()) {
		return uuid.Nil, errors.New("The token has expired")
	}

	return id, nil
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
