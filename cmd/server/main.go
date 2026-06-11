package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"go.uber.org/zap"

	"github.com/yourusername/user-api/config"
	"github.com/yourusername/user-api/db/sqlc"
	"github.com/yourusername/user-api/internal/handler"
	"github.com/yourusername/user-api/internal/logger"
	"github.com/yourusername/user-api/internal/middleware"
	"github.com/yourusername/user-api/internal/repository"
	"github.com/yourusername/user-api/internal/routes"
	"github.com/yourusername/user-api/internal/service"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		fmt.Println("No .env file found, using system environment variables")
	}

	// Load config
	cfg := config.Load()

	// Initialize logger
	log := logger.NewLogger(cfg.Env)
	defer log.Sync()

	log.Info("Starting server", zap.String("env", cfg.Env), zap.String("port", cfg.Port))

	// Connect to database
	dbPool, err := connectDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer dbPool.Close()

	log.Info("Connected to database successfully")

	// Initialize SQLC queries
	queries := db.New(dbPool)

	// Initialize layers
	userRepo := repository.NewUserRepository(queries)
	userSvc := service.NewUserService(userRepo, log)
	userHandler := handler.NewUserHandler(userSvc, log)

	// Setup Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: handler.ErrorHandler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	})

	// Global middleware
	app.Use(recover.New())
	app.Use(middleware.RequestID())
	app.Use(middleware.RequestLogger(log))

	// Serve static files from /public directory
	app.Static("/", "./public")

	// Register routes
	routes.RegisterUserRoutes(app, userHandler)

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "timestamp": time.Now().UTC()})
	})

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		addr := fmt.Sprintf(":%s", cfg.Port)
		log.Info("Server listening", zap.String("addr", addr))
		if err := app.Listen(addr); err != nil {
			log.Fatal("Server failed to start", zap.Error(err))
		}
	}()

	<-quit
	log.Info("Shutting down server...")

	if err := app.Shutdown(); err != nil {
		log.Error("Error shutting down server", zap.Error(err))
	}

	log.Info("Server exited cleanly")
}

func connectDB(databaseURL string) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	config.MaxConns = 25
	config.MinConns = 5
	config.MaxConnLifetime = 30 * time.Minute
	config.MaxConnIdleTime = 5 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return pool, nil
}
