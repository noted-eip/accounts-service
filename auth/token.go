package auth

import (
	"github.com/golang-jwt/jwt"
)

// Token represents the payload section of a JWT.
type Token struct {
	AccountID string `json:"uid,omitempty"`
	jwt.StandardClaims
}
