package fido

import (
	"context"
	"fmt"
	"reflect"
)

// A Writer is given to a Provider to write values to Fido.
type Writer interface {
	Write(path Path, value interface{}) error
}

// A WriterFunc is an adapter that allows regular functions to act as Writers.
type WriterFunc func(path Path, value interface{}) error

// Write calls the wrapped function implementing the Writer interface.
func (fn WriterFunc) Write(path Path, value interface{}) error {
	return fn(path, value)
}

// WriterMiddleware is a function that allows Writers to be wrapped with other Writers.
type WriterMiddleware func(Writer) Writer

// WrapWriter wraps a Writer with the provided writer middleware functions.
func WrapWriter(writer Writer, middlewares ...WriterMiddleware) Writer {
	return WriterFunc(func(path Path, value interface{}) error {
		for i := len(middlewares) - 1; i >= 0; i-- {
			writer = middlewares[i](writer)
		}

		return writer.Write(path, value)
	})
}

func (f *Fido) writer(ctx context.Context, provider Provider) Writer {
	return WriterFunc(func(path Path, value interface{}) error {
		field, ok := f.fields.get(path)
		if !ok {
			return fmt.Errorf("%w: %s", ErrFieldNotFound, field)
		}

		current := field.Value().Interface()

		if value != current {
			return nil
		}

		if err := field.Set(value, provider); err != nil {
			return fmt.Errorf("%w: failed to set field %s value %v", err, path, value)
		}

		return nil
	})
}

func (f *Fido) initMapMiddleware() WriterMiddleware {
	return WriterMiddleware(func(next Writer) Writer {
		return WriterFunc(func(path Path, value interface{}) error {
			field, ok := f.fields.get(path)
			if !ok {
				return fmt.Errorf("%w: %s", ErrFieldNotFound, field)
			}

			if !field.Path().equal(path) && field.Value().Kind() == reflect.Map {
				if err := f.initMapField(path, field); err != nil {
					return err
				}
			}

			return next.Write(path, value)
		})
	})
}

func (f *Fido) notificationMiddleware(provider Provider, ch chan<- *FieldUpdate) WriterMiddleware {
	return WriterMiddleware(func(next Writer) Writer {
		return WriterFunc(func(path Path, value interface{}) error {
			field, ok := f.fields.get(path)
			if !ok {
				return fmt.Errorf("%w: %s", ErrFieldNotFound, field)
			}

			current := field.Value().Interface()

			if err := next.Write(path, value); err != nil {
				return err
			}

			if current != value {
				ch <- &FieldUpdate{
					Path:     path,
					Old:      current,
					New:      value,
					Provider: provider,
				}
			}

			return nil
		})
	})
}

func (f *Fido) enforcePriorityMiddleware(provider Provider) WriterMiddleware {
	return WriterMiddleware(func(next Writer) Writer {
		return WriterFunc(func(path Path, value interface{}) error {
			field, ok := f.fields.get(path)
			if !ok {
				return fmt.Errorf("%w: %s", ErrFieldNotFound, field)
			}

			if field.Provider() != nil && f.options.EnforcePriority {
				if f.providers[field.Provider()] > f.providers[provider] {
					return nil
				}
			}

			return next.Write(path, value)
		})
	})
}
