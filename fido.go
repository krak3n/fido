package fido

import (
	"context"
	"errors"
	"fmt"
	"reflect"
)

// DefaultStructTag is the struct tag fido looks for to populate values from providers.
const DefaultStructTag = "fido"

// Options configures Fido behaviour.
type Options struct {
	// Enforce provider priority when settings values, e.g provider 1 cannot set values set by
	// provider 3. Default: true.
	EnforcePriority bool
	// StructTag to look for on destination struct fields. Default: fido.
	StructTag string
	// Return an error if a provider provides a value for a field that does not exist on the
	// destination struct. Default: false.
	ErrorOnFieldNotFound bool
	// Return an error if the fido struct tag is not found on a destination struct field. Default: true.
	ErrorOnMissingTag bool
}

// An Option configures Fido behaviour.
type Option interface {
	apply(*Options)
}

// OptionFunc is an adapter allowing regular methods to act as Option's.
type OptionFunc func(*Options)

func (fn OptionFunc) apply(o *Options) {
	fn(o)
}

// WithStructTag configures the struct tag Fido looks for on destination struct types.
func WithStructTag(t string) Option {
	return OptionFunc(func(o *Options) {
		o.StructTag = t
	})
}

// Callback is a function given to a Provider to call when it has values to give to Fido for
// processing.
type Callback func(path Path, value interface{}) error

// Fido is a extensible configuration loader.
type Fido struct {
	providers providers
	fields    fields
	options   Options
}

// New constructs a new Fido.
func New(dst interface{}, opts ...Option) (*Fido, error) {
	f := &Fido{
		providers: make(providers),
		fields:    make(fields),
		options: Options{
			EnforcePriority:   true,
			ErrorOnMissingTag: true,
			StructTag:         DefaultStructTag,
		},
	}

	for _, opt := range opts {
		opt.apply(&f.options)
	}

	return f, f.hydrate([]string{}, reflect.ValueOf(dst))
}

// Fetch fetches configuration values from the given providers with a background context.
func (f *Fido) Fetch(providers ...Provider) error {
	return f.FetchWithContext(context.Background(), providers...)
}

// FetchWithContext fetches configuration values from the given providers.
// A named return value is used to catch and return a wrapped recover error on panic.
func (f *Fido) FetchWithContext(ctx context.Context, providers ...Provider) (err error) {
	defer func() {
		if r := recover(); r != nil {
			switch r := r.(type) {
			case error:
				err = fmt.Errorf("%w: recovered from panic", r)
			default:
				err = fmt.Errorf("%v: recovered from panic", r)
			}
		}
	}()

	for _, provider := range providers {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err = f.fetch(ctx, provider); err != nil {
				return
			}
		}
	}

	return
}

// Close loops over providers closing them if they implement the CloseProvider interface.
func (f *Fido) Close() error {
	for provider := range f.providers {
		if closer, ok := provider.(CloseProvider); ok {
			if err := closer.Close(); err != nil {
				return err
			}
		}
	}

	return nil
}

func (f *Fido) hydrate(p Path, v reflect.Value) error {
	if len(p) == 0 {
		switch {
		case !v.IsValid():
			return ErrDestinationTypeInvalid
		case v.IsNil():
			return ErrDestinationNil
		case v.Type().Kind() != reflect.Ptr:
			return fmt.Errorf("%w: %s", ErrDestinationNotPtr, v.Type().Kind())
		case v.Elem().Kind() != reflect.Struct:
			return fmt.Errorf("%w: %s", ErrDestinationTypeInvalid, v.Elem().Kind())
		}

		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		f.fields.set(p, &field{
			path:  p,
			value: v,
		})

		return nil
	}

	for i := 0; i < v.NumField(); i++ {
		fv := v.Field(i)
		ft := v.Type().Field(i)

		tag, err := LookupTag(f.options.StructTag, ft)
		if err != nil {
			if errors.Is(err, ErrStructTagNotFound) && !f.options.ErrorOnMissingTag {
				continue
			}

			return fmt.Errorf("%w: failed parse struct tag for %s", err, ft.Name)
		}

		if err := f.hydrate(append(p, tag.Name), fv); err != nil {
			return fmt.Errorf("%w: failed parse struct tag for %s", err, ft.Name)
		}
	}

	return nil
}

func (f *Fido) fetch(ctx context.Context, p Provider) error {
	if !f.providers.exists(p) {
		f.providers.add(p)
	}

	if err := p.Values(ctx, f.callback(p)); err != nil {
		return err
	}

	return nil
}

func (f *Fido) callback(p Provider) Callback {
	return Callback(func(path Path, value interface{}) error {
		for {
			field, ok := f.fields.get(path)
			if !ok {
				if f.options.ErrorOnFieldNotFound {
					return fmt.Errorf("%w: %s", ErrFieldNotFound, field)
				}

				return nil
			}

			if !field.Path().equal(path) && field.Value().Kind() == reflect.Map {
				if err := f.initMapField(path, field); err != nil {
					return err
				}

				continue
			}

			if field.Provider() != nil && f.options.EnforcePriority {
				if f.providers[field.Provider()] > f.providers[p] {
					return nil
				}
			}

			return field.Set(value, p)
		}
	})
}

func (f *Fido) initMapField(path Path, fld Field) error {
	mp := path[len(fld.Path())-1:]

	mv, err := f.initMap(mp, fld.Value())
	if err != nil {
		return err
	}

	f.fields.set(path, &mapfield{
		field: &field{
			path:  path,
			value: reflect.New(mv.Type().Elem()).Elem(),
		},
		dst: mv,
		idx: reflect.ValueOf(mp[len(mp)-1]),
	})

	return nil
}

func (f *Fido) initMap(path Path, value reflect.Value) (reflect.Value, error) {
	// Ensure we have a map
	if value.Kind() != reflect.Map {
		return value, fmt.Errorf("%w for %s got %s", ErrExpectedMap, path, value)
	}

	// If the value is nil, initialise a new map of the correct types and set it as the fields value
	if value.IsNil() {
		if value.Type().Key().Kind() != reflect.String {
			return value, fmt.Errorf("%w for %s got %s", ErrInvalidMapKeyType, path, value.Type().Key())
		}

		if !value.CanAddr() {
			return value, fmt.Errorf("cannot initialise map: %w", ErrReflectValueNotAddressable)
		}

		value.Set(reflect.MakeMap(
			reflect.MapOf(
				value.Type().Key(),
				value.Type().Elem())))
	}

	// If this is a nested map we need to initialise a new map with the correct type.
	if value.Type().Elem().Kind() == reflect.Map {
		if len(path) <= 1 {
			return value, fmt.Errorf("%w: cannot initialise nested map with path %d length", ErrInvalidPath, len(path))
		}

		parent, children := path[0], path[1:]

		// Check if we have this map index in the map, if not create a new value for the index of
		// the correct type - this will be another map.
		m := value.MapIndex(reflect.ValueOf(parent))
		if !m.IsValid() {
			m = reflect.New(value.Type().Elem()).Elem()
		}

		// As this is a map we now init that map
		field, err := f.initMap(children, m)
		if err != nil {
			return value, err
		}

		// Set the map index of key parent to the value of the new map.
		value.SetMapIndex(reflect.ValueOf(parent), m)

		return field, nil
	}

	// Return the map for index value setting.
	return value, nil
}
