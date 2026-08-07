package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const sessionTTL = 30 * 24 * time.Hour

type Claims struct {
	UserID int `json:"user_id"`
	jwt.RegisteredClaims
}

type Signer struct {
	secret []byte
}

func NewSigner(secret string) *Signer {
	return &Signer{secret: []byte(secret)}
}

// IssueSession returns a JWT identifying userID, valid for 30 days. This is
// our own session token, separate from Google's ID token — Google is only
// consulted once at login, not on every request.
func (s *Signer) IssueSession(userID int) (string, error) {
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(sessionTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

func (s *Signer) VerifySession(tokenString string) (int, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return s.secret, nil
	})
	if err != nil {
		return 0, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return 0, errors.New("invalid token")
	}
	return claims.UserID, nil
}
