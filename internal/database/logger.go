package database

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/librarease/librarease/internal/config"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type SlogGormLogger struct {
	Logger        *slog.Logger
	LogLevel      logger.LogLevel
	SlowThreshold time.Duration
}

func NewSlogGormLogger(l *slog.Logger) *SlogGormLogger {
	// Map slog level to gorm logger level based on ENV_KEY_LOG_LEVEL
	gormLevel := logger.Info // default: log all SQL queries

	if lvl := os.Getenv(config.ENV_KEY_LOG_LEVEL); lvl != "" {
		switch lvl {
		case "DEBUG":
			gormLevel = logger.Info // log all SQL queries
		case "INFO":
			gormLevel = logger.Info // log all SQL queries
		case "WARN":
			gormLevel = logger.Warn // log only SQL errors
		case "ERROR":
			gormLevel = logger.Error // log only SQL errors
		}
	}

	return &SlogGormLogger{
		Logger:        l,
		LogLevel:      gormLevel,
		SlowThreshold: 200 * time.Millisecond,
	}
}

func (l *SlogGormLogger) LogMode(level logger.LogLevel) logger.Interface {
	newLogger := *l
	newLogger.LogLevel = level
	return &newLogger
}

func (l *SlogGormLogger) Info(ctx context.Context, msg string, args ...interface{}) {
	if l.LogLevel >= logger.Info {
		l.Logger.InfoContext(ctx, fmt.Sprintf(msg, args...))
	}
}

func (l *SlogGormLogger) Warn(ctx context.Context, msg string, args ...interface{}) {
	if l.LogLevel >= logger.Warn {
		l.Logger.WarnContext(ctx, fmt.Sprintf(msg, args...))
	}
}

func (l *SlogGormLogger) Error(ctx context.Context, msg string, args ...interface{}) {
	if l.LogLevel >= logger.Error {
		l.Logger.ErrorContext(ctx, fmt.Sprintf(msg, args...))
	}
}

func (l *SlogGormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel <= logger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	fields := []any{
		slog.String("sql", sql),
		slog.Int64("rows", rows),
		slog.Duration("latency", elapsed),
		slog.String("latency_human", elapsed.String()),
	}

	switch {
	case err != nil && l.LogLevel >= logger.Error && !errors.Is(err, gorm.ErrRecordNotFound):
		fields = append(fields, slog.String("source", l.getSource()), slog.String("err", err.Error()))
		l.Logger.ErrorContext(ctx, "sql_error", fields...)
	case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= logger.Warn:
		fields = append(fields, slog.String("source", l.getSource()), slog.String("slow_threshold", l.SlowThreshold.String()))
		l.Logger.WarnContext(ctx, "sql_slow", fields...)
	case l.LogLevel == logger.Info:
		fields = append(fields, slog.String("source", l.getSource()))
		l.Logger.InfoContext(ctx, "sql", fields...)
	}
}

func (l *SlogGormLogger) getSource() string {
	for i := 2; i < 15; i++ {
		_, file, line, ok := runtime.Caller(i)
		if ok && (!strings.Contains(file, "gorm.io") && !strings.HasSuffix(file, "internal/database/logger.go")) {
			return file + ":" + strconv.Itoa(line)
		}
	}
	return ""
}
