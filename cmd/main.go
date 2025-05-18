package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	handler "warehouse-service/app/handler/api"
	"warehouse-service/app/middleware"
	"warehouse-service/app/repository/db"
	"warehouse-service/app/usecase"
	"warehouse-service/config"
	"warehouse-service/pkg/logger"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/healthcheck"
	"github.com/gofiber/fiber/v2/middleware/recover"
	slogfiber "github.com/samber/slog-fiber"
)

func main() {
	// init logger
	logger.InitLogger()

	// init config
	cfg, err := config.InitConfig(context.Background())
	if err != nil {
		slog.Error("failed to init config", "error", err)
		return
	}

	// init database
	dbConn, err := db.NewPostgres(cfg.Db)
	if err != nil {
		log.Fatalf("DB connection failed: %v", err)
	}
	defer dbConn.Close()

	reqValidator := validator.New()
	userRepo := db.NewUserRepository(dbConn)

	userUsecase := usecase.NewWarehouseUsecase(userRepo, cfg)

	userHandler := handler.NewWarehouseHandler(userUsecase, reqValidator)

	// Initialize HTTP web framework
	app := fiber.New()
	app.Use(healthcheck.New(healthcheck.Config{
		LivenessProbe: func(c *fiber.Ctx) bool {
			return true
		},
		LivenessEndpoint: "/live",
		ReadinessProbe: func(c *fiber.Ctx) bool {
			return true
		},
		ReadinessEndpoint: "/ready",
	}))
	webLogger := slog.New(&logger.RequestIDHandler{Handler: slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})})
	app.Use(slogfiber.New(webLogger))
	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
	}))
	app.Use(middleware.RequestIDMiddleware())

	handler.SetupRouter(app, userHandler, cfg)

	go func() {
		if err := app.Listen(":" + cfg.Port); err != nil {
			slog.Error("Failed to listen", "port", cfg.Port)
			return
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	slog.Info("Gracefully shutdown")
	err = app.Shutdown()
	if err != nil {
		slog.Warn("Unfortunately the shutdown wasn't smooth", "err", err)
	}
}
