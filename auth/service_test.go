package auth_test

import (
	"accounts-service/auth"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
)

func Test_service_ContextWithToken(t *testing.T) {
	srv := auth.NewService([]byte("secret"))

	ctx, err := srv.ContextWithToken(context.TODO(), &auth.Token{})
	assert.NoError(t, err)
	md, ok := metadata.FromOutgoingContext(ctx)
	assert.True(t, ok)
	assert.True(t, len(md) > 0)
	assert.True(t, md.Get(auth.TokenMetadataKey)[0] == "<token>")
}
