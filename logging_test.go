package log

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/Egyptmaster/golang-x-log/formatter"
	"github.com/Egyptmaster/golang-x-log/provider"
	"github.com/stretchr/testify/assert"
)

func TestSlog(t *testing.T) {
	cfg := Config{
		IsDefault:  true,
		AddSource:  true,
		LogLevel:   "DEBUG",
		UTC:        true,
		TimeFormat: time.RFC3339,
		Values: map[string]any{
			"company": "cid",
			"version": 1.0,
		},
	}

	builder, err := NewBuilderFromCfg(cfg)
	assert.NoError(t, err)

	builder.SetDefault(true)
	builder.SetWriter(os.Stdout)
	builder.AddFormatter(formatter.RelativeSource())
	builder.AddFormatter(formatter.RewriteKey(slog.SourceKey, "logger"))
	builder.AddProvider(provider.FromContext[string]("uuid"))
	logger, err := builder.Build()
	assert.NoError(t, err)

	logger.Error("this is an error log without context")

	ctx := context.WithValue(context.Background(), "uuid", "1A2B")
	logger.ErrorContext(ctx, "this is an error log with context")

}
