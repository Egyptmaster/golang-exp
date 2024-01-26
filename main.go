package main

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"log/slog"
	"os"
	"reflect"
)

type X struct {
	S string
}

type Z struct {
	Q string
}

type Y struct {
	Ptr    *Z
	NonPtr X
	B      string
	X      int
}

type container struct {
	creators map[string]func() (*any, error)
}

func main() {

	// override logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)

	// create some container with "instantiation" methods
	c := &container{creators: make(map[string]func() (*any, error))}
	c.creators["X"] = func() (*any, error) {
		var x any = X{S: uuid.NewString()}
		return &x, nil
	}
	c.creators["Z"] = func() (*any, error) {
		var z any = Z{Q: uuid.NewString()}
		return &z, nil
	}

	// create a map with some default values
	defaults := map[string]any{
		"B": "go",
		"X": 42,
	}
	ins, err := Create[Y](c, defaults)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", *ins)
}

func Create[T any](c *container, defaults map[string]any) (*T, error) {
	var value T
	typeOf := reflect.TypeOf(value)

	// type has to be a struct
	if typeOf.Kind() != reflect.Struct {
		return nil, errors.New(fmt.Sprintf("type %v is not of a struct", typeOf.Name()))
	}

	// create an addressable pointer to instance of the struct
	ptr := reflect.ValueOf(&value).Elem()
	for i := 0; i < typeOf.NumField(); i++ {
		field := typeOf.Field(i)
		settable := ptr.FieldByName(field.Name)

		// check if there is a setter for the field
		if !settable.CanSet() {
			slog.Debug(fmt.Sprintf("field %v has no setter", field.Name))
			continue
		}
		// check if there is any value specified for the field by its name -> assign it
		if d, ok := defaults[field.Name]; ok {
			settable.Set(reflect.ValueOf(d))
			slog.Debug(fmt.Sprintf("set value '%v' of type %T to field %s", d, d, field.Name))
		}
		// check if there is any registration for the type
		typeName := field.Type.Name()
		isPtr := field.Type.Kind() == reflect.Ptr
		if isPtr {
			typeName = field.Type.Elem().Name()
		}
		r, ok := c.creators[typeName]
		if !ok {
			continue
		}
		val, err := r()
		if err != nil {
			return nil, err
		}
		if isPtr {
			panic("not yet supported")
		} else {
			slog.Debug(fmt.Sprintf("set struct '%+v' of type %T to field %s", *val, val, field.Name))
			settable.Set(reflect.ValueOf(*val))
		}

	}
	// cast the created value to requested interface type
	var asAny any = value
	service, err := as[T](&asAny)
	return service, err
}

func nameOf[T any]() string {
	var ptr *T = nil
	return reflect.TypeOf(ptr).Elem().Name()
}

func as[T any](value *any) (*T, error) {
	srv := nameOf[T]()
	if t, ok := (*value).(T); !ok {
		return nil, errors.New(fmt.Sprintf("cannot convert %v of type %T to %v", value, value, srv))
	} else {
		return &t, nil
	}
}
