package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func generateToken(email string, userID uuid.UUID, role Role) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email":   email,
		"userID":  userID.String(),
		"expires": time.Now().Add(role.TokenDuration()).Unix(),
	})

	signedToken, err := token.SignedString(role.SecretKey())

	if err != nil {
		return "", problem(role.Role(), err)
	}
	return signedToken, nil
}

func VerifyToken(token string, secretKey []byte) (uuid.UUID, string, error) {
	parsedToken, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("Unexpected signing method")
		}

		return secretKey, nil
	})

	if err != nil {
		return uuid.Nil, "", err
	}

	if !parsedToken.Valid {
		return uuid.Nil, "", errors.New("Invalid token")
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return uuid.Nil, "", errors.New("Invalid token")
	}

	id, err := uuid.Parse(claims["userID"].(string))
	if err != nil {
		return uuid.Nil, "", errors.New("Invalid token")
	}

	email, ok := claims["email"].(string)
	if !ok {
		return uuid.Nil, "", errors.New("Invalid token")
	}

	expires, ok := claims["expires"].(float64)
	if !ok {
		return uuid.Nil, "", errors.New("Invalid token")
	}

	if time.Unix(int64(expires), 0).Before(time.Now()) {
		return uuid.Nil, "", errors.New("The token has expired")
	}

	return id, email, nil
}
