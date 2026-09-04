package jwt_test

import (
	"context"
	"testing"
	"time"

	"github.com/agusheryanto182/redis-playground/pkg/jwt"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const redisAddr = "localhost:6379"

func newTestJWT(
	t *testing.T,
	secret string,
	duration time.Duration,
) (*jwt.Manager, *redis.Client) {
	t.Helper()

	ctx := context.Background()

	redisClient := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	require.NoError(t, redisClient.Ping(ctx).Err())

	t.Cleanup(func() {
		require.NoError(t, redisClient.Close())
	})

	return jwt.New(secret, duration, redisClient), redisClient
}

func TestJWT_GenerateAndParse(t *testing.T) {
	ctx := context.Background()

	j, redisClient := newTestJWT(t, "test-secret", time.Hour)

	token, err := j.GenerateToken(ctx, "user-123")
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	userID, err := j.ParseToken(ctx, token)
	require.NoError(t, err)
	assert.Equal(t, "user-123", userID)

	t.Cleanup(func() {
		require.NoError(t, redisClient.Del(ctx, token).Err())
	})
}

func TestJWT_ParseToken_Invalid(t *testing.T) {
	ctx := context.Background()

	j, _ := newTestJWT(t, "test-secret", time.Hour)

	_, err := j.ParseToken(ctx, "invalid-token")
	require.Error(t, err)
}

func TestJWT_ParseToken_WrongSecret(t *testing.T) {
	ctx := context.Background()

	j1, redisClient1 := newTestJWT(t, "secret-1", time.Hour)
	j2, _ := newTestJWT(t, "secret-2", time.Hour)

	token, err := j1.GenerateToken(ctx, "user-123")
	require.NoError(t, err)

	_, err = j2.ParseToken(ctx, token)
	require.Error(t, err)

	t.Cleanup(func() {
		require.NoError(t, redisClient1.Del(ctx, token).Err())
	})
}

func TestJWT_ParseToken_Expired(t *testing.T) {
	ctx := context.Background()

	j, _ := newTestJWT(t, "test-secret", -time.Hour)

	token, err := j.GenerateToken(ctx, "user-123")
	require.NoError(t, err)

	_, err = j.ParseToken(ctx, token)
	require.Error(t, err)
}

func TestJWT_ParseToken_TokenNotFoundInRedis(t *testing.T) {
	ctx := context.Background()

	j, redisClient := newTestJWT(t, "test-secret", time.Hour)

	token, err := j.GenerateToken(ctx, "user-123")
	require.NoError(t, err)

	err = redisClient.Del(ctx, token).Err()
	require.NoError(t, err)

	_, err = j.ParseToken(ctx, token)
	require.Error(t, err)
	assert.ErrorIs(t, err, jwt.ErrTokenNotFound)
}

func TestJWT_ParseToken_UserIDMismatch(t *testing.T) {
	ctx := context.Background()

	j, redisClient := newTestJWT(t, "test-secret", time.Hour)

	token, err := j.GenerateToken(ctx, "user-123")
	require.NoError(t, err)

	err = redisClient.Set(
		ctx,
		token,
		"user-456",
		time.Hour,
	).Err()
	require.NoError(t, err)

	_, err = j.ParseToken(ctx, token)
	require.Error(t, err)
	assert.ErrorIs(t, err, jwt.ErrUserIDMismatch)

	t.Cleanup(func() {
		require.NoError(t, redisClient.Del(ctx, token).Err())
	})
}

func TestJWT_RevokeToken(t *testing.T) {
	ctx := context.Background()

	j, redisClient := newTestJWT(t, "test-secret", time.Hour)

	token, err := j.GenerateToken(ctx, "user-123")
	require.NoError(t, err)

	userID, err := j.ParseToken(ctx, token)
	require.NoError(t, err)
	assert.Equal(t, "user-123", userID)

	err = j.RevokeToken(ctx, token)
	require.NoError(t, err)

	_, err = j.ParseToken(ctx, token)
	require.Error(t, err)
	assert.ErrorIs(t, err, jwt.ErrTokenNotFound)

	t.Cleanup(func() {
		require.NoError(t, redisClient.Del(ctx, token).Err())
	})
}
