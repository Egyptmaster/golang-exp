package formatter

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

type Formatter interface {
	ShouldHandle(attr slog.Attr) bool
	Format(*slog.Attr)
}

type ByKeyFormatter struct {
	Formatter
	key string
}

func (f *ByKeyFormatter) ShouldHandle(attr slog.Attr) bool {
	return attr.Key == f.key
}

// RewriteKeyFormatter will format the key
type RewriteKeyFormatter struct {
	ByKeyFormatter
	newKey string
}

func (r *RewriteKeyFormatter) Format(attr *slog.Attr) {
	attr.Key = r.newKey // change the key
}

// RewriteKey provides a formatter which will rename any occurrence of the given key by the new key
func RewriteKey(key string, newKey string) Formatter {
	f := &RewriteKeyFormatter{
		newKey: newKey,
	}
	f.key = key
	return f
}

// RelativeSource will convert the absolute source to relative path
func RelativeSource() Formatter {
	f := &RelativeSourceLogFormatter{}
	f.key = slog.SourceKey
	return f
}

type RelativeSourceLogFormatter struct {
	ByKeyFormatter
}

func (r *RelativeSourceLogFormatter) Format(attr *slog.Attr) {
	val, ok := attr.Value.Any().(*slog.Source)
	if !ok {
		return
	}
	exec, err := os.Getwd()
	if err != nil {
		return
	}

	path := filepath.Dir(exec)
	path = filepath.ToSlash(path)
	if strings.HasPrefix(val.File, path) {
		val.File = val.File[len(path)+1:]
	}
}

// TimeFormatFormatter is able to format time log attributes
type TimeFormatFormatter struct {
	UTC    *bool
	Layout *string
}

func (r *TimeFormatFormatter) Format(attr *slog.Attr) {
	time := attr.Value.Time()
	// convert to UTC
	if r.UTC != nil && *r.UTC {
		time = time.UTC()
	}
	// layout the time
	if r.Layout == nil {
		attr.Value = slog.TimeValue(time)
	} else {
		attr.Value = slog.StringValue(time.Format(*r.Layout))
	}
}
func (r *TimeFormatFormatter) Key() string {
	return slog.TimeKey
}

// ByKeyTimeFormatFormatter is able to format the time log attribute
type ByKeyTimeFormatFormatter struct {
	ByKeyFormatter
	TimeFormatFormatter
}

func TimeFormat(layout string, utc bool) Formatter {
	f := &ByKeyTimeFormatFormatter{}
	f.UTC = &utc
	f.Layout = &layout
	f.key = slog.TimeKey
	return f
}

// EveryTimeFormatFormatter is able to format all log attributes of kind time
type EveryTimeFormatFormatter struct {
	TimeFormatFormatter
}

func (f *EveryTimeFormatFormatter) ShouldHandle(attr slog.Attr) bool {
	return attr.Value.Kind() == slog.KindTime
}

func TimesFormat(layout string, utc bool) Formatter {
	f := &EveryTimeFormatFormatter{}
	f.UTC = &utc
	f.Layout = &layout
	return f
}
