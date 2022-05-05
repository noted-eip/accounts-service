package auth_test

import (
	"accounts-service/auth"
	"context"
	"crypto/ed25519"
	"testing"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
)

func Test_service_ContextWithToken(t *testing.T) {
	// Given
	pub, priv := genKeyOrFail(t)
	srv := auth.NewService(priv)
	uid := uuid.New()

	// When
	ctx, err := srv.ContextWithToken(context.TODO(), &auth.Token{
		UserID: uid,
		Role:   auth.RoleAdmin,
	})

	// Then
	require.NoError(t, err)

	var tokenString string

	t.Run("there should be metadata in the context", func(t *testing.T) {
		md, ok := metadata.FromOutgoingContext(ctx)
		require.True(t, ok)
		tokenString = md.Get(auth.TokenMetadataKey)[0]
		require.NotZero(t, tokenString)
	})

	var claims *auth.Token

	t.Run("the token should be valid and decodable with the public key", func(t *testing.T) {
		tok, err := jwt.ParseWithClaims(tokenString, &auth.Token{}, func(*jwt.Token) (interface{}, error) {
			return pub, nil
		})
		require.NoError(t, err)
		var ok bool
		claims, ok = tok.Claims.(*auth.Token)
		require.True(t, ok)
	})

	t.Run("the token should contain user data", func(t *testing.T) {
		require.Equal(t, claims.UserID, uid)
	})
}

func Test_service_TokenFromContext(t *testing.T) {
	// Given
	_, priv := genKeyOrFail(t)
	srv := auth.NewService(priv)
	uid := uuid.New()
	ctx, err := srv.ContextWithToken(context.TODO(), &auth.Token{
		UserID: uid,
		Role:   auth.RoleAdmin,
	})
	require.NoError(t, err)

	// When
	token, err := srv.TokenFromContext(ctx)

	// Then
	require.NoError(t, err)

	t.Run("the token should contain the user data", func(t *testing.T) {
		require.Equal(t, token.UserID, uid)
	})
}

func genKeyOrFail(t *testing.T) (ed25519.PublicKey, ed25519.PrivateKey) {
	pub, priv, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)
	return pub, priv
}
