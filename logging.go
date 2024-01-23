package log

import (
	"bytes"
	"context"
	"egyptnmaster/golang-x-log/formatter"
	"egyptnmaster/golang-x-log/provider"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
)

type Config struct {
	LogLevel   string `json:"level"`
	AddSource  bool
	IsDefault  bool
	TimeFormat string
	UTC        bool
	Values     map[string]any
}

func (l Config) Level() slog.Level {
	level := slog.LevelDebug
	if err := level.UnmarshalText([]byte(l.LogLevel)); err != nil {
		panic(fmt.Errorf("the level is not support %v", l.LogLevel))
	}
	return level
}

type LoggerBuilder struct {
	formatter  []formatter.Formatter
	providers  []provider.ValueProvider
	setDefault bool
	addSource  bool
	level      slog.Level
	w          io.Writer
}

func NewBuilderFromFile(cfgFile string) (*LoggerBuilder, error) {
	data, err := os.ReadFile(cfgFile)
	if err != nil {
		return nil, err
	}
	return NewBuilderFromJson(data)
}

func NewBuilderFromJson(encoded []byte) (*LoggerBuilder, error) {
	var cfg Config
	if err := json.NewDecoder(bytes.NewBuffer(encoded)).Decode(&cfg); err != nil {
		return nil, err
	}
	return NewBuilderFromCfg(cfg)
}

func NewBuilderFromCfg(cfg Config) (*LoggerBuilder, error) {
	level := slog.LevelDebug
	if err := level.UnmarshalText([]byte(cfg.LogLevel)); err != nil {
		return nil, err
	}

	builder := NewBuilder()
	builder.SetLevel(level)
	builder.SetDefault(cfg.IsDefault)
	builder.addSource = cfg.AddSource

	if len(cfg.TimeFormat) > 0 {
		builder.AddFormatter(formatter.TimeFormat(cfg.TimeFormat, cfg.UTC))
	}

	if cfg.Values != nil && len(cfg.Values) > 0 {
		for key, value := range cfg.Values {
			builder.AddProvider(provider.Static(key, value))
		}
	}

	return builder, nil
}

func NewBuilder() *LoggerBuilder {
	return &LoggerBuilder{
		level:      slog.LevelInfo,
		addSource:  false,
		w:          os.Stdout,
		setDefault: false,
	}
}

func (l *LoggerBuilder) SetLevel(lvl slog.Level) *LoggerBuilder {
	l.level = lvl
	return l
}

func (l *LoggerBuilder) SetDefault(setDefault bool) *LoggerBuilder {
	l.setDefault = setDefault
	return l
}

func (l *LoggerBuilder) AddFormatter(formatters ...formatter.Formatter) *LoggerBuilder {
	l.formatter = append(l.formatter, formatters...)
	return l
}

func (l *LoggerBuilder) AddProvider(prov ...provider.ValueProvider) *LoggerBuilder {
	l.providers = append(l.providers, prov...)
	return l
}

func (l *LoggerBuilder) SetWriter(w io.Writer) *LoggerBuilder {
	l.w = w
	return l
}

func (l *LoggerBuilder) Build() (*slog.Logger, error) {
	// create the handler options
	options := &slog.HandlerOptions{
		AddSource: l.addSource,
		Level:     l.level,
	}

	// add formatter as replacer
	if len(l.formatter) > 0 {
		replacer := func(groups []string, a slog.Attr) slog.Attr {
			for _, logFmt := range l.formatter {
				if logFmt.ShouldHandle(a) {
					logFmt.Format(&a)
				}
			}
			return a
		}
		options.ReplaceAttr = replacer
	}

	// create a new JSON encoding handler and maybe warp by context handler
	var handler slog.Handler = slog.NewJSONHandler(l.w, options)
	if len(l.providers) > 0 {
		handler = newExtendedHandler(handler, l.providers...)
	}

	// finally create the logger and set to default if required
	logger := slog.New(handler)
	if l.setDefault {
		slog.SetDefault(logger)
	}

	return logger, nil
}

// extendedHandler implements the slog.Handler interface to enable usage of additional ValueProviders
type extendedHandler struct {
	slog.Handler
	providers []provider.ValueProvider
}

func newExtendedHandler(handler slog.Handler, providers ...provider.ValueProvider) *extendedHandler {
	var h = &extendedHandler{
		providers: providers,
	}
	h.Handler = handler
	return h
}

// Handle adds contextual attributes to the Record before calling the underlying handler
func (h *extendedHandler) Handle(ctx context.Context, r slog.Record) error {

	for _, valueSource := range h.providers {
		if attr, ok := valueSource.Get(ctx); ok {
			r.AddAttrs(attr)
		}
	}

	return h.Handler.Handle(ctx, r)
}
