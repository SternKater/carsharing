package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"


	"github.com/SternKater/carsharing/internal/delivery/grpc/handler"
	"github.com/SternKater/carsharing/internal/delivery/grpc/interceptors"
	"github.com/SternKater/carsharing/internal/repository/postgres"
	"github.com/SternKater/carsharing/internal/service"
	"github.com/SternKater/carsharing/internal/tokens"
	"github.com/SternKater/carsharing/pkg/auth"
	"github.com/SternKater/carsharing/pkg/cars"
	"github.com/redis/go-redis/v9"
	"github.com/jackc/pgx/v5/pgxpool"

	"google.golang.org/grpc"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

// Reddis
	redisConnStr := os.Getenv("REDIS_CONN_STR")
	if redisConnStr == "" {
		redisConnStr = "localhost:6380"
	}
	rdb := redis.NewClient(&redis.Options{Addr: redisConnStr})
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("[SERVER]: Failed to ping Redis: %v", err)
	}
	defer rdb.Close()
// Rate Limit check
	rateLimiterInterceptor, err := interceptors.NewUnaryRateLimiterInterceptor(rdb, 10, 60, "scripts/lua/rate_limiter.lua")	
	if err != nil {
		log.Fatalf("[SERVER]: Failed to create RateLimiterInterceptor: %v", err)
	}
// BruteForce Limit check
	bruteForceLimiterInterceptor, err := interceptors.NewUnaryBruteForceLimiterInterceptor(rdb, 5, 30, "scripts/lua/rate_limiter.lua")	
	if err != nil {
		log.Fatalf("[SERVER]: Failed to create BruteForceLimiterInterceptor: %v", err)
	}
// JWT-token check
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "jwt-secret"
	}

	tkMgr := tokens.NewTokenManager([]byte(jwtSecret))
	tokenInterceptor := interceptors.NewUnaryTokenInterceptor(rdb, tkMgr)
	
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			rateLimiterInterceptor.UnaryServerInterceptor, 
			bruteForceLimiterInterceptor.UnaryServerInterceptor,
			tokenInterceptor.UnaryServerInterceptor,
		),
	)

// DB routine
	postgresConnStr := os.Getenv("POSTGRES_CONN_STR")
	if postgresConnStr == "" {
		postgresConnStr = "postgres://user:password@localhost:5433/carsharing_db"
	}
	postgresPool, err := pgxpool.New(ctx, postgresConnStr)
	if err != nil {
		log.Fatalf("[SERVER]: Failed to connect to db: %v", err)
	}
	if err := postgresPool.Ping(ctx); err != nil {
		log.Fatalf("[SERVER]: DB ping error: %v", err)
	}
	defer postgresPool.Close()

// repository
	txManager := postgres.NewTxManager(postgresPool)
	authRepo := postgres.NewAuthRepository(postgresPool)
// service
	authService := service.NewAuthService(authRepo, tkMgr, txManager)
// handler	
	authHandler := handler.NewAuthHandler(authService)
	auth.RegisterAuthServiceServer(grpcServer, authHandler)

// same as prev	
	carsRepo := postgres.NewCarsRepository(postgresPool)
	carsService := service.NewCarsService(carsRepo)
	carsHandler := handler.NewCarsHandler(carsService)
	cars.RegisterCarsServiceServer(grpcServer, carsHandler)

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	log.Println("🚀[SERVER]: gRPC server is started on :50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("[SERVER]: Failed to serve: %v", err)
	}
}