package interceptors

import (
	"context"
	"log"
	"os"
	"fmt"

	"github.com/SternKater/carsharing/pkg/auth"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UnaryBruteForceLimiterInterceptor struct {
	rdb 	*redis.Client
	limit 	int
	window 	int
	sha 	string
}

func (u *UnaryBruteForceLimiterInterceptor) Allow(ctx context.Context, login string) (bool, error) {
	result, err := u.rdb.EvalSha(ctx, u.sha, []string{fmt.Sprintf("brute:login:%s", login)}, u.limit, u.window).Result()
	if err != nil {
		return false, err
	}

	return result.(int64) == 1, nil
}

func NewUnaryBruteForceLimiterInterceptor(r *redis.Client, l int, w int, s string) (*UnaryBruteForceLimiterInterceptor, error) {
	content, err := os.ReadFile(s)
	if err != nil {
		return nil, err
	}
	sha, err := r.ScriptLoad(context.Background(), string(content)).Result()
	if err != nil {
		return nil, err
	}
	return &UnaryBruteForceLimiterInterceptor{rdb: r, limit: l, window: w, sha: sha}, nil
}

func (u *UnaryBruteForceLimiterInterceptor)UnaryServerInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
    log.Printf("[INTERCEPTOR][BruteLimiter]: Received request on method: %s", info.FullMethod)

	login := ""
// sign in
	if loginReq, ok := req.(*auth.SignInRequest); ok {
		login = loginReq.UserName
	}
// signup
	if loginReq, ok := req.(*auth.SignUpRequest); ok {
		login = loginReq.UserName
	}

	if login != "" {
		log.Printf("[BruteLimiter] Login detected: %s", login)

		allowed, err := u.Allow(ctx, login)
		if err != nil {
			return nil, err
		}

		if !allowed {
			return nil, status.Error(codes.PermissionDenied, "Limit exhausted")
		}
	}

    resp, err := handler(ctx, req)
    log.Printf("[INTERCEPTOR][BruteLimiter]: Sending response from method: %s", info.FullMethod)

    return resp, err
}
