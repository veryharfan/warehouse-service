package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	handler "warehouse-service/app/handler/api"
	"warehouse-service/app/middleware"
	"warehouse-service/app/repository/broker"
	"warehouse-service/app/repository/db"
	"warehouse-service/app/usecase"
	"warehouse-service/config"
	"warehouse-service/pkg/logger"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/healthcheck"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	slogfiber "github.com/samber/slog-fiber"
)

func main() {
	// init logger
	logger.InitLogger()

	ctx := context.Background()
	// init config
	cfg, err := config.InitConfig(ctx)
	if err != nil {
		slog.Error("failed to init config", "error", err)
		return
	}

	// init database
	dbConn, err := db.NewPostgres(cfg.Db)
	if err != nil {
		slog.Error("DB connection failed", "error", err)
	}
	defer dbConn.Close()

	// Connect to NATS server
	nc, err := nats.Connect(cfg.Nats.Url) // default is nats://localhost:4222
	if err != nil {
		slog.Error("Error connecting to NATS", "error", err)
		return
	}
	defer nc.Drain()

	js, err := jetstream.New(nc)
	if err != nil {
		slog.Error("Error creating JetStream context", "error", err)
		return
	}
	_, err = js.CreateStream(ctx, jetstream.StreamConfig{
		Name:     strings.ToUpper(cfg.Nats.StreamName),
		Subjects: []string{fmt.Sprintf("%s.*", strings.ToLower(cfg.Nats.StreamName))},
		Storage:  jetstream.FileStorage,
	})
	if err != nil && !errors.Is(err, jetstream.ErrStreamNameAlreadyInUse) {
		slog.Error("create STOCK stream failed", "error", err)
		return
	}

	reqValidator := validator.New()
	warehouseRepo := db.NewWarehouseRepository(dbConn)
	stockRepo := db.NewStockRepository(dbConn)
	stockBroker := broker.NewStockBrokerPublisher(js)

	warehouseUsecase := usecase.NewWarehouseUsecase(warehouseRepo, cfg)
	stockUsecase := usecase.NewStockUsecase(stockRepo, warehouseRepo, stockBroker, cfg)

	warehouseHandler := handler.NewWarehouseHandler(warehouseUsecase, reqValidator)
	stockHandler := handler.NewStockHandler(stockUsecase, reqValidator)

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

	handler.SetupRouter(app, warehouseHandler, stockHandler, cfg)

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
