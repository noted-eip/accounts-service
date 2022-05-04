package auth

import (
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
)

// Token represents the payload section of a JWT.
type Token struct {
	Role   Role      `json:"role,omitempty"`
	UserID uuid.UUID `json:"uid,omitempty"`
	jwt.StandardClaims
}
