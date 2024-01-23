package provider

import (
	"context"
	"log/slog"
)

type ValueProvider interface {
	Get(ctx context.Context) (slog.Attr, bool)
}

// FromContextSource is able to read a value from the context
type FromContextSource[T any] struct {
	key string
}

// Get implement ValueSource
func (src *FromContextSource[T]) Get(ctx context.Context) (slog.Attr, bool) {
	if value, ok := ctx.Value(src.key).(T); ok {
		return slog.Any(src.key, value), true
	}
	return slog.Attr{}, false
}

func FromContext[T any](key string) ValueProvider {
	return &FromContextSource[T]{
		key: key,
	}
}

// StaticValueSource is able to add a static a value to each log statement
type StaticValueSource[T any] struct {
	key   string
	value T
}

// Get implement ValueSource
func (src *StaticValueSource[T]) Get(context.Context) (slog.Attr, bool) {
	return slog.Any(src.key, src.value), true
}

func Static[T any](key string, value T) ValueProvider {
	return &StaticValueSource[T]{
		key:   key,
		value: value,
	}
}
