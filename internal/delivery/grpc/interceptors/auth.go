package interceptors

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/SternKater/carsharing/internal/domain"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type jwtExtractor interface {
    ExtractAuthToken(ctx context.Context) (string, error)
}

type AuthTokenFields struct {
    UserId  int64 `json:"user_id"`
    Exp     int64 `json:"exp"`
    jwt.RegisteredClaims
}

type UnaryTokenInterceptor struct {
    rdb *redis.Client
    tokenExtractor jwtExtractor
}

func NewUnaryTokenInterceptor(db *redis.Client, extractor jwtExtractor) *UnaryTokenInterceptor {
    return &UnaryTokenInterceptor{rdb: db, tokenExtractor: extractor}
}

func (u *UnaryTokenInterceptor)UnaryServerInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
    log.Printf("[INTERCEPTOR][JWTToken]: Received request on method: %s", info.FullMethod)
    
// no token needed
    if info.FullMethod == "/auth.AuthService/SignIn" || 
    info.FullMethod == "/auth.AuthService/SignUp" {
        resp, err := handler(ctx, req)
        log.Printf("[INTERCEPTOR][JWTToken]: Sending response from method: %s", info.FullMethod)
        return resp, err
    }

// no authorization field or invalid field format
    token, err := u.tokenExtractor.ExtractAuthToken(ctx)
    if err != nil {
        return nil, err
    }

// token has been added to stop-list
    _, err = u.rdb.Get(ctx,fmt.Sprintf("stoplist:token:%s", token)).Result()
    if err != redis.Nil {
        return nil, status.Error(codes.Unauthenticated, "Token in stop-list")
    } else if err != nil {
        return nil, err
    }

    jwtToken, err := jwt.ParseWithClaims(token, &AuthTokenFields{}, func(token *jwt.Token) (interface{}, error) {
    	jwtSecret := os.Getenv("JWT_SECRET")
	    if jwtSecret == "" {
		    jwtSecret = "jwt-secret"
	    }
        return []byte(jwtSecret), nil
    })

// corrupted token string    
    if err != nil {
        return nil, err
    }
    
// wrong token string format
    claims, ok := jwtToken.Claims.(*AuthTokenFields)
    if !ok || !jwtToken.Valid {
        return nil, status.Error(codes.Unauthenticated, "Invalid token")
    }

// token has expired(it;s excessive but let it be)
    if claims.Exp < time.Now().Unix() {
        return nil, status.Error(codes.Unauthenticated, "Token has expired")
    }

    resp, err := handler(context.WithValue(ctx, domain.UserIDKey, claims.UserId), req)
    log.Printf("[INTERCEPTOR][JWTToken]: Sending response from method: %s", info.FullMethod)
    return resp, err
}
