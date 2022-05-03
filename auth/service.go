// Package auth implements all the authentication logic used by
// the accounts service.
//
// TODO: Add configurable standard claims such as the token expiration date.
package auth

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"google.golang.org/grpc/metadata"
)

var (
	ErrExpiredToken     = errors.New("token is expired")
	ErrInvalidSignature = errors.New("signature is invalid")
	ErrNoTokenInCtx     = errors.New("no token in context")
)

const (
	authorizationMetadataKey = "authorization"
)

type TokenInfo struct {
	UserID uuid.UUID
	Role   Role
	jwt.StandardClaims
}

// Service is used to create JWTs for use with other services or to
// verify JWTs emitted by other services. A service is safe for use
// in multiple goroutines.
type Service interface {
	// TokenInfoFromContext verifies the token stored inside of ctx and
	// extracts the payload. If the token is missing or invalid an error
	// is returned.
	TokenInfoFromContext(ctx context.Context) (*TokenInfo, error)

	// ContextWithToken returns a copy of parent in which a new value for the
	// key 'noted-token' is set to a string encoded JWT.
	ContextWithTokenInfo(parent context.Context, info *TokenInfo) (context.Context, error)

	// DefaultStandardClaims fills in the default claims structure according to
	// RFC 7519.
	DefaultStandardClaims() jwt.StandardClaims
}

// NewService creates a new authentication service which encodes/decodes
// signed-JWTs with the key provided as argument.
func NewService(key []byte) Service {
	return &service{
		key: key,
	}
}

type service struct {
	key []byte
}

type jwtCustomClaims struct {
	Role
	jwt.StandardClaims
}

func (srv service) TokenInfoFromContext(ctx context.Context) (*TokenInfo, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, ErrNoTokenInCtx
	}
	values := md.Get(authorizationMetadataKey)
	if len(values) == 0 {
		return nil, ErrNoTokenInCtx
	}
	tokenString := values[0]
	info := &TokenInfo{}
	token, err := jwt.ParseWithClaims(tokenString, info.StandardClaims, func(t *jwt.Token) (interface{}, error) {
		return srv.key, nil
	})
	if err != nil {
		return nil, ErrInvalidSignature
	}

	return nil, nil
}

func (srv service) ContextWithTokenInfo(parent context.Context, info *TokenInfo) (context.Context, error) {
	return nil, nil
}

func (srv service) DefaultStandardClaims() jwt.StandardClaims {
	return jwt.StandardClaims{
		Audience:  "noted",
		Issuer:    "noted",
		ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
		IssuedAt:  time.Now().Unix(),
	}
}
