package tokens

import (
	"time"
	"context"
	"strings"

	"github.com/SternKater/carsharing/internal/domain"

	"google.golang.org/grpc/metadata"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type TokenManager struct {
	jwtSecret []byte
}

func NewTokenManager(secret []byte) *TokenManager {
	return &TokenManager{jwtSecret: secret}
}

func (t *TokenManager) JWTByUserID(id int64) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{"user_id": id,
			"exp": time.Now().Add(time.Minute * 15).Unix()},
	)
	jwtToken, _ := token.SignedString(t.jwtSecret)

	return jwtToken
}

func (t *TokenManager) ExtractAuthToken(ctx context.Context) (string, error) {
    md, ok := metadata.FromIncomingContext(ctx)
    if !ok {
		return "", domain.ErrInvalidAccessToken
    }

    authHeader := md.Get("authorization")
    if len(authHeader) == 0 {
		return "", domain.ErrInvalidAccessToken
    }

    if len(authHeader) != 1 {
		return "", domain.ErrInvalidAccessToken
    }

    parts := strings.SplitN(authHeader[0], " ", 2)

    if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", domain.ErrInvalidAccessToken
    }

    return parts[1], nil
}

func (t *TokenManager) GenerateRefreshToken() string {
	newUUID := uuid.New()
	return newUUID.String()
}
