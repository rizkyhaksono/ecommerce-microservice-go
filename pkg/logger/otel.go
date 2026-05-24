package logger

import (
	"context"

	"go.opentelemetry.io/contrib/bridges/otelzap"
	otellog "go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewLoggerWithOTel returns a Logger whose records are written to BOTH the JSON
// stdout core (so `docker logs` still works) AND the OpenTelemetry LoggerProvider,
// which exports them over OTLP to the collector and on to Loki.
func NewLoggerWithOTel(lp otellog.LoggerProvider, scopeName string, dev bool) (*Logger, error) {
	level := zap.InfoLevel
	if dev {
		level = zap.DebugLevel
	}
	otelCore := otelzap.NewCore(scopeName, otelzap.WithLoggerProvider(lp))
	tee := zapcore.NewTee(stdoutCore(level), otelCore)

	var opts []zap.Option
	if dev {
		opts = append(opts, zap.AddStacktrace(zap.ErrorLevel))
	}
	return &Logger{Log: zap.New(tee, opts...)}, nil
}

// Ctx returns a child *zap.Logger annotated with trace_id/span_id from the
// active span in ctx, so log lines can be correlated with traces in Grafana
// (Loki derived field -> Tempo). When ctx has no valid span, the base logger is
// returned unchanged.
func (l *Logger) Ctx(ctx context.Context) *zap.Logger {
	sc := trace.SpanContextFromContext(ctx)
	if !sc.IsValid() {
		return l.Log
	}
	return l.Log.With(
		zap.String("trace_id", sc.TraceID().String()),
		zap.String("span_id", sc.SpanID().String()),
	)
}
