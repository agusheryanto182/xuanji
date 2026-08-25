package jwt

import (
	"context"
	"errors"
	"fmt"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
)

var (
	// ErrUnexpectedSigningMethod is returned when the JWT signing method is not expected.
	ErrUnexpectedSigningMethod = errors.New("unexpected signing method")

	// ErrTokenNotFound is returned when the JWT token does not exist in Redis.
	ErrTokenNotFound = errors.New("token not found in Redis")

	// ErrUserIDMismatch is returned when the JWT subject does not match the Redis user ID.
	ErrUserIDMismatch = errors.New("user ID mismatch")
)

// Manager handles JWT token generation and parsing.
type Manager struct {
	secret      string
	duration    time.Duration
	redisClient *redis.Client
}

// New creates a new JWT manager.
func New(secret string, duration time.Duration, redisClient *redis.Client) *Manager {
	return &Manager{
		secret:      secret,
		duration:    duration,
		redisClient: redisClient,
	}
}

// GenerateToken creates a new JWT token for the given user ID.
func (m *Manager) GenerateToken(ctx context.Context, userID string) (string, error) {
	token := jwtlib.NewWithClaims(
		jwtlib.SigningMethodHS256,
		jwtlib.RegisteredClaims{
			Subject:   userID,
			ExpiresAt: jwtlib.NewNumericDate(time.Now().Add(m.duration)),
		},
	)

	tokenString, err := token.SignedString([]byte(m.secret))
	if err != nil {
		return "", fmt.Errorf(
			"jwt - GenerateToken - token.SignedString: %w",
			err,
		)
	}

	if err := m.redisClient.Set(
		ctx,
		tokenString,
		userID,
		m.duration,
	).Err(); err != nil {
		return "", fmt.Errorf(
			"jwt - GenerateToken - redisClient.Set: %w",
			err,
		)
	}

	return tokenString, nil
}

// ParseToken validates a JWT token and returns the user ID.
func (m *Manager) ParseToken(
	ctx context.Context,
	tokenString string,
) (string, error) {
	token, err := jwtlib.Parse(
		tokenString,
		func(token *jwtlib.Token) (any, error) {
			if token.Method != jwtlib.SigningMethodHS256 {
				return nil, fmt.Errorf(
					"%w: %v",
					ErrUnexpectedSigningMethod,
					token.Header["alg"],
				)
			}

			return []byte(m.secret), nil
		},
	)
	if err != nil {
		return "", fmt.Errorf(
			"jwt - ParseToken - jwtlib.Parse: %w",
			err,
		)
	}

	if !token.Valid {
		return "", errors.New("jwt - ParseToken - invalid token")
	}

	sub, err := token.Claims.GetSubject()
	if err != nil {
		return "", fmt.Errorf(
			"jwt - ParseToken - GetSubject: %w",
			err,
		)
	}

	redisUserID, err := m.redisClient.Get(
		ctx,
		tokenString,
	).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", fmt.Errorf(
				"jwt - ParseToken - %w",
				ErrTokenNotFound,
			)
		}

		return "", fmt.Errorf(
			"jwt - ParseToken - redisClient.Get: %w",
			err,
		)
	}

	if redisUserID != sub {
		return "", fmt.Errorf(
			"jwt - ParseToken - %w",
			ErrUserIDMismatch,
		)
	}

	return sub, nil
}

func (m *Manager) RevokeToken(ctx context.Context, tokenString string) error {
	if err := m.redisClient.Del(ctx, tokenString).Err(); err != nil {
		return fmt.Errorf(
			"jwt - RevokeToken - redisClient.Del: %w",
			err,
		)
	}

	return nil
}
