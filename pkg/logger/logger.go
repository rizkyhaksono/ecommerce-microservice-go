package logger

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	gormlogger "gorm.io/gorm/logger"
)

type Logger struct {
	Log *zap.Logger
}

// baseEncoderConfig is the shared JSON encoder config used by every logger
// variant (stdout-only and the OTel tee).
func baseEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.FullCallerEncoder,
		EncodeName:     zapcore.FullNameEncoder,
	}
}

// stdoutCore builds the JSON-to-stdout core at the given level.
func stdoutCore(level zapcore.Level) zapcore.Core {
	return zapcore.NewCore(
		zapcore.NewJSONEncoder(baseEncoderConfig()),
		zapcore.AddSync(os.Stdout),
		zap.NewAtomicLevelAt(level),
	)
}

func NewLogger() (*Logger, error) {
	return &Logger{Log: zap.New(stdoutCore(zap.InfoLevel))}, nil
}

func NewDevelopmentLogger() (*Logger, error) {
	return &Logger{Log: zap.New(stdoutCore(zap.DebugLevel), zap.AddStacktrace(zap.ErrorLevel))}, nil
}

func (l *Logger) Info(msg string, fields ...zap.Field)  { l.Log.Info(msg, fields...) }
func (l *Logger) Error(msg string, fields ...zap.Field) { l.Log.Error(msg, fields...) }
func (l *Logger) Fatal(msg string, fields ...zap.Field) { l.Log.Fatal(msg, fields...) }
func (l *Logger) Panic(msg string, fields ...zap.Field) { l.Log.Panic(msg, fields...) }
func (l *Logger) Warn(msg string, fields ...zap.Field)  { l.Log.Warn(msg, fields...) }
func (l *Logger) Debug(msg string, fields ...zap.Field) { l.Log.Debug(msg, fields...) }

func (l *Logger) SetupGinWithZapLogger() {
	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = &ZapWriter{logger: l.Log}
	gin.DefaultErrorWriter = &ZapErrorWriter{logger: l.Log}
}

func (l *Logger) SetupGinWithZapLoggerInDevelopment() {
	gin.SetMode(gin.DebugMode)
	gin.DefaultWriter = &ZapWriter{logger: l.Log}
	gin.DefaultErrorWriter = &ZapErrorWriter{logger: l.Log}
}

type ZapWriter struct{ logger *zap.Logger }

func (w *ZapWriter) Write(p []byte) (n int, err error) {
	w.logger.Info("Gin-log", zap.String("message", string(p)))
	return len(p), nil
}

type ZapErrorWriter struct{ logger *zap.Logger }

func (w *ZapErrorWriter) Write(p []byte) (n int, err error) {
	w.logger.Error("Gin-error", zap.String("error", string(p)))
	return len(p), nil
}

func (l *Logger) GinZapLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		latency := time.Since(start)
		// Ctx adds trace_id/span_id from the span created by otelgin upstream,
		// so access logs in Loki link back to the trace in Tempo.
		l.Ctx(c.Request.Context()).Info("HTTP request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", latency),
			zap.String("client_ip", c.ClientIP()),
		)
	}
}

type GormZapLogger struct {
	zap    *zap.SugaredLogger
	config gormlogger.Config
}

func NewGormLogger(base *zap.Logger) *GormZapLogger {
	return &GormZapLogger{
		zap: base.Sugar(),
		config: gormlogger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  gormlogger.Error,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	}
}

func (l *GormZapLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	newCfg := l.config
	newCfg.LogLevel = level
	return &GormZapLogger{zap: l.zap, config: newCfg}
}

func (l *GormZapLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.config.LogLevel >= gormlogger.Info {
		l.zap.Infof(msg, data...)
	}
}

func (l *GormZapLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.config.LogLevel >= gormlogger.Warn {
		l.zap.Warnf(msg, data...)
	}
}

func (l *GormZapLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.config.LogLevel >= gormlogger.Error &&
		(!l.config.IgnoreRecordNotFoundError || msg != gormlogger.ErrRecordNotFound.Error()) {
		l.zap.Errorf(msg, data...)
	}
}

func (l *GormZapLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	elapsed := time.Since(begin)
	if err != nil {
		if l.config.IgnoreRecordNotFoundError && errors.Is(err, gormlogger.ErrRecordNotFound) {
			return
		}
		if l.config.LogLevel >= gormlogger.Error {
			sql, rows := fc()
			l.zap.Errorf("Error: %v | %.3fms | rows:%d | %s", err, float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
		return
	}
	if elapsed > l.config.SlowThreshold && l.config.LogLevel >= gormlogger.Warn {
		sql, rows := fc()
		l.zap.Warnf("SLOW ≥ %s | %.3fms | rows:%d | %s", l.config.SlowThreshold, float64(elapsed.Nanoseconds())/1e6, rows, sql)
	}
}
