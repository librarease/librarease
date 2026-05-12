package bootstrap

import (
	"fmt"
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"

	"github.com/librarease/librarease/internal/config"
)

type Config struct {
	LogLevel string
	Port     int

	DB DBConfig

	Redis RedisConfig
	SMTP  SMTPConfig
	MinIO MinIOConfig

	Worker WorkerConfig
}

type DBConfig struct {
	Host               string
	Port               string
	Database           string
	User               string
	Password           string
	MaxOpenConnections int
	MaxIdleConnections int
	ConnMaxLifetime    time.Duration
	ConnMaxIdleTime    time.Duration
}

func (c DBConfig) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", c.User, c.Password, c.Host, c.Port, c.Database)
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
}

func (c RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

type SMTPConfig struct {
	Host     string
	Port     string
	Username string
	Password string
}

type MinIOConfig struct {
	Bucket     string
	TempPath   string
	PublicPath string
	Endpoint   string
	AccessKey  string
	SecretKey  string
}

type WorkerConfig struct {
	Concurrency int
}

func LoadConfig() Config {
	return Config{
		LogLevel: os.Getenv(config.ENV_KEY_LOG_LEVEL),
		Port:     intEnv(config.ENV_KEY_PORT, 0),
		DB: DBConfig{
			Host:               os.Getenv(config.ENV_KEY_DB_HOST),
			Port:               os.Getenv(config.ENV_KEY_DB_PORT),
			Database:           os.Getenv(config.ENV_KEY_DB_DATABASE),
			User:               os.Getenv(config.ENV_KEY_DB_USER),
			Password:           os.Getenv(config.ENV_KEY_DB_PASSWORD),
			MaxOpenConnections: intEnv(config.ENV_KEY_DB_MAX_OPEN_CONNECTIONS, 25),
			MaxIdleConnections: intEnv(config.ENV_KEY_DB_MAX_IDLE_CONNECTIONS, 10),
			ConnMaxLifetime:    durationMinutesEnv(config.ENV_KEY_DB_CONN_MAX_LIFETIME_MINUTES, 5*time.Minute),
			ConnMaxIdleTime:    durationMinutesEnv(config.ENV_KEY_DB_CONN_MAX_IDLE_TIME_MINUTES, 2*time.Minute),
		},
		Redis: RedisConfig{
			Host:     os.Getenv(config.ENV_KEY_REDIS_HOST),
			Port:     os.Getenv(config.ENV_KEY_REDIS_PORT),
			Password: os.Getenv(config.ENV_KEY_REDIS_PASSWORD),
		},
		SMTP: SMTPConfig{
			Host:     os.Getenv(config.ENV_KEY_SMTP_HOST),
			Port:     os.Getenv(config.ENV_KEY_SMTP_PORT),
			Username: os.Getenv(config.ENV_KEY_SMTP_USERNAME),
			Password: os.Getenv(config.ENV_KEY_SMTP_PASSWORD),
		},
		MinIO: MinIOConfig{
			Bucket:     os.Getenv(config.ENV_KEY_MINIO_BUCKET),
			TempPath:   os.Getenv(config.ENV_KEY_MINIO_TEMP_PATH),
			PublicPath: os.Getenv(config.ENV_KEY_MINIO_PUBLIC_PATH),
			Endpoint:   os.Getenv(config.ENV_KEY_MINIO_ENDPOINT),
			AccessKey:  os.Getenv(config.ENV_KEY_MINIO_ACCESS_KEY),
			SecretKey:  os.Getenv(config.ENV_KEY_MINIO_SECRET_KEY),
		},
		Worker: WorkerConfig{
			Concurrency: intEnv(config.ENV_KEY_WORKER_CONCURRENCY, 10),
		},
	}
}

func intEnv(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	n, err := strconv.Atoi(value)
	if err != nil || n <= 0 {
		return fallback
	}
	return n
}

func durationMinutesEnv(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	n, err := strconv.Atoi(value)
	if err != nil || n <= 0 {
		return fallback
	}
	return time.Duration(n) * time.Minute
}
