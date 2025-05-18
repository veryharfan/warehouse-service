package config

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

type Config struct {
	Port               string    `mapstructure:"PORT" validate:"required"`
	InternalAuthHeader string    `mapstructure:"INTERNAL_AUTH_HEADER" validate:"required"`
	Db                 DbConfig  `mapstructure:",squash"`
	Jwt                JwtConfig `mapstructure:",squash"`
}

type DbConfig struct {
	Host     string `mapstructure:"DB_HOST" validate:"required"`
	Port     string `mapstructure:"DB_PORT" validate:"required"`
	Username string `mapstructure:"DB_USERNAME" validate:"required"`
	Password string `mapstructure:"DB_PASSWORD" validate:"required"`
	DbName   string `mapstructure:"DB_DBNAME" validate:"required"`
	SSLMode  string `mapstructure:"DB_SSLMODE"`
}

type JwtConfig struct {
	SecretKey string `mapstructure:"JWT_SECRETKEY" validate:"required"`
	Expire    int64  `mapstructure:"JWT_EXPIRE" validate:"required"`
}

func InitConfig(ctx context.Context) (*Config, error) {
	var cfg Config

	// Reset viper to avoid any previous configuration
	viper.Reset()

	// Make viper case insensitive for environment variables
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Set the configuration type
	viper.SetConfigType("env")

	// Try to load from .env file if it exists
	envFile := os.Getenv("ENV_FILE")
	if envFile == "" {
		envFile = ".env"
	}

	_, err := os.Stat(envFile)
	if !os.IsNotExist(err) {
		viper.SetConfigFile(envFile)

		if err := viper.ReadInConfig(); err != nil {
			slog.WarnContext(ctx, "[InitConfig] ReadInConfig warning, continuing with env vars only", "error", err)
			// Continue with just environment variables instead of returning error
		} else {
			slog.InfoContext(ctx, "[InitConfig] Successfully loaded config file", "file", envFile)
		}
	} else {
		slog.InfoContext(ctx, "[InitConfig] No config file found, using environment variables")
	}

	// Load environment variables
	viper.AutomaticEnv()

	// Debug: Print environment variables we're looking for
	envVars := []string{
		"PORT",
		"DB_HOST",
		"DB_PORT",
		"DB_USERNAME",
		"DB_PASSWORD",
		"DB_DBNAME",
		"DB_SSLMODE",
		"JWT_SECRETKEY",
		"JWT_EXPIRE",
		"INTERNAL_AUTH_HEADER",
	}

	slog.InfoContext(ctx, "[InitConfig] Environment variables debug:")

	// Bind environment variables explicitly to ensure they're mapped correctly
	for _, key := range envVars {
		viper.BindEnv(key)
	}

	// Unmarshal configuration
	if err := viper.Unmarshal(&cfg); err != nil {
		slog.ErrorContext(ctx, "[InitConfig] Unmarshal", "failed bind config", err)
		return nil, err
	}

	// Log the entire configuration after binding
	slog.InfoContext(ctx, "[InitConfig] Configuration after binding",
		"PORT", cfg.Port,
		"DB_HOST", cfg.Db.Host,
		"DB_PORT", cfg.Db.Port,
		"DB_USERNAME", cfg.Db.Username,
		"DB_DBNAME", cfg.Db.DbName,
		"DB_SSLMODE", cfg.Db.SSLMode,
		"JWT_EXPIRE", cfg.Jwt.Expire)

	// Validate configuration
	validate := validator.New()
	if err := validate.Struct(cfg); err != nil {
		validationErrs, ok := err.(validator.ValidationErrors)
		if ok {
			for _, validationErr := range validationErrs {
				slog.ErrorContext(ctx, "[InitConfig] Validation error",
					"field", validationErr.Field(),
					"namespace", validationErr.Namespace(),
					"tag", validationErr.Tag(),
					"value", validationErr.Value())
			}
		} else {
			slog.ErrorContext(ctx, "[InitConfig] Validation", "error", err)
		}
		return nil, err
	}

	slog.InfoContext(ctx, "[InitConfig] Config loaded successfully")
	return &cfg, nil
}
