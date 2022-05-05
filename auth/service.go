// Package auth implements all the authentication logic used by
// the accounts service.
//
// TODO: Add configurable standard claims such as the token expiration date.
package auth

import (
	"context"
	"crypto/ed25519"
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt"
	"google.golang.org/grpc/metadata"
)

var (
	ErrNoTokenInCtx    = errors.New("no token in context")
	ErrNoMetadataInCtx = errors.New("no metadata in context")
	ErrInvalidToken    = errors.New("invalid token")
)

const (
	TokenMetadataKey = "authorization"
)

// Service is used to create JWTs for use with other services or to
// verify JWTs emitted by other services. A service is safe for use
// in multiple goroutines.
type Service interface {
	// TokenFromContext verifies the token stored inside of ctx and
	// extracts the payload. If the token is missing or invalid an error
	// is returned.
	TokenFromContext(ctx context.Context) (*Token, error)

	// ContextWithToken returns a copy of parent in which a new value for the
	// key 'noted-token' is set to a string encoded JWT.
	ContextWithToken(parent context.Context, info *Token) (context.Context, error)
}

// NewService creates a new authentication service which encodes/decodes
// signed-JWTs with the key provided as argument.
func NewService(key ed25519.PrivateKey) Service {
	return &service{
		key: key,
	}
}

type service struct {
	key ed25519.PrivateKey
}

func (srv *service) TokenFromContext(ctx context.Context) (*Token, error) {
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		return nil, ErrNoMetadataInCtx
	}

	values := md.Get(TokenMetadataKey)
	if len(values) == 0 {
		return nil, ErrNoTokenInCtx
	}
	tokenString := values[0]

	tok, err := jwt.ParseWithClaims(tokenString, &Token{}, func(t *jwt.Token) (interface{}, error) {
		pub, ok := srv.key.Public().(ed25519.PublicKey)
		if !ok {
			// This should never happen
			return nil, errors.New("invalid key")
		}
		return pub, nil
	})
	if err != nil {
		return nil, fmt.Errorf("could not parse token: %v", err)
	}

	claims, ok := tok.Claims.(*Token)
	if !ok {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

func (srv *service) ContextWithToken(parent context.Context, info *Token) (context.Context, error) {
	token := jwt.NewWithClaims(&jwt.SigningMethodEd25519{}, info)
	ss, err := token.SignedString(srv.key)
	if err != nil {
		return nil, err
	}
	return metadata.AppendToOutgoingContext(parent, TokenMetadataKey, ss), nil
}
