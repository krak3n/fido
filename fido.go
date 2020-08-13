package fido

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
)

// DefaultStructTag is the struct tag fido looks for to populate values from providers.
const DefaultStructTag = "fido"

// Options configures Fido behaviour.
type Options struct {
	// AutoWatch sets Fido to automatically start watching providers that support the NotifyProvider
	// optional extension interface. Default: true.
	AutoWatch bool
	// AutoUpdate sets Fido to automatically update configuration values when a Provider notifies
	// Fido of a change. Default: true.
	AutoUpdate bool
	// EnforcePriority will ensure that providers with lower priority cannot set values set by
	// a higher priority provider, e.g provider 1 cannot set values set by provider 3. Default: true.
	EnforcePriority bool
	// StructTag is the name of the struct tag to look for on destination struct fields. Default: fido.
	StructTag string
	// ErrorOnFieldNotFound configures Fido to return an error if a provider provides a value for a field
	// that does not exist on the destination struct. Default: false.
	ErrorOnFieldNotFound bool
	// ErrorOnMissingTag configures Fido to return an error if the Fido struct tag is not found on a
	// destination struct field. Default: true.
	ErrorOnMissingTag bool
}

// DefaultOptions returns the default configuration options for Fido.
func DefaultOptions() Options {
	return Options{
		AutoWatch:            true,
		AutoUpdate:           true,
		EnforcePriority:      true,
		ErrorOnMissingTag:    true,
		ErrorOnFieldNotFound: false,
		StructTag:            DefaultStructTag,
	}
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

// SetAutoWatch sets Fidos AutoWatch behaviour.
func SetAutoWatch(v bool) Option {
	return OptionFunc(func(o *Options) {
		o.AutoWatch = v
	})
}

// SetAutoUpdate sets Fidos AutoUpdate behaviour.
func SetAutoUpdate(v bool) Option {
	return OptionFunc(func(o *Options) {
		o.AutoUpdate = v
	})
}

// SetPriorityEnforcement configures provider priority enforcement. Set to false to
// disable, true to enable enforcement.
func SetPriorityEnforcement(enforce bool) Option {
	return OptionFunc(func(o *Options) {
		o.EnforcePriority = enforce
	})
}

// SetErrorOnMissingTag configures Fido behaviour when it finds a struct field without tag, if
// false Fido will ignore the field and carry on processing the other fields on the struct, if set
// true a ErrStructTagNotFound will be returned.
func SetErrorOnMissingTag(err bool) Option {
	return OptionFunc(func(o *Options) {
		o.ErrorOnMissingTag = err
	})
}

// Callback is a function given to a Provider to call when it has values to give to Fido for
// processing.
type Callback func(path Path, value interface{}) error

// FieldUpdate holds meta data about a change to a fields value.
// Old and New values are not guarantee to be populated. Always check the value of Err.
type FieldUpdate struct {
	Path     Path        // Path to the field that has changed
	Old      interface{} // The previous value
	New      interface{} // The new value
	Provider Provider    // The provider that set the value
}

// A Notification holds meta data about a change to a field.
type Notification interface {
	Updates() ([]*FieldUpdate, error)
}

// A FieldUpdateError satisfies the Notification interface and is published when an error occurs
// when updating a fields value.
type FieldUpdateError struct {
	Err error
}

// Updates returns a nil slice of *FieldUpdate and the Error that occured.
func (e *FieldUpdateError) Updates() ([]*FieldUpdate, error) {
	return nil, e.Err
}

// FieldUpdates is a slice of pointers to FieldUpdate values. It implements the Notification
// interface allowing FieldUpdates to be published to subscribers.
type FieldUpdates []*FieldUpdate

// Updates returns the slices of *FieldUpdate with no error.
func (u FieldUpdates) Updates() ([]*FieldUpdate, error) {
	return u, nil
}

// Fido is a extensible configuration loader.
type Fido struct {
	wg          sync.WaitGroup
	subscribers []chan Notification
	providers   providers
	watching    providers
	fields      fields
	options     Options
}

// New constructs a new Fido.
func New(dst interface{}, opts ...Option) (*Fido, error) {
	f := &Fido{
		providers: make(providers),
		fields:    make(fields),
		options:   DefaultOptions(),
	}

	for _, opt := range opts {
		opt.apply(&f.options)
	}

	return f, f.hydrate([]string{}, reflect.ValueOf(dst))
}

// Add adds providers to Fido.
func (f *Fido) Add(providers ...Provider) {
	f.providers.add(providers...)
}

// Fetch fetches configuration values from the given providers with a background context.
func (f *Fido) Fetch(providers ...Provider) error {
	return f.FetchWithContext(context.Background(), providers...)
}

// FetchWithContext fetches configuration values from the given providers with the provided context.
// If the AutoWatch option is enabled Fido will start watching providers that support the
// NotifyProvider optional extension interface automatically. If the AutoWatch option is disabled
// you will need to call Watch/WatchWithContext for the providers you wish Fido to watch.
func (f *Fido) FetchWithContext(ctx context.Context, providers ...Provider) error {
	f.Add(providers...)

	for provider := range f.providers {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if f.options.AutoWatch {
				if err := f.WatchWithContext(ctx); err != nil {
					return err
				}
			}

			if err := f.fetch(ctx, provider); err != nil {
				return err
			}
		}
	}

	return nil
}

// Watch starts watching providers that support the NotifyProvider optional extension interface.
func (f *Fido) Watch(providers ...Provider) error {
	return f.FetchWithContext(context.Background())
}

// WatchWithContext starts watching providers that support the NotifyProvider optional extension
// interface with the provided context.
func (f *Fido) WatchWithContext(ctx context.Context, providers ...Provider) error {
	f.Add(providers...)

	for provider := range f.providers {
		if notifier, ok := provider.(NotifyProvider); ok {
			if _, ok := f.watching[provider]; !ok {
				ch, err := notifier.Notify()
				if err != nil {
					return fmt.Errorf("%w: failed to start provider %s notifier", err, provider)
				}

				f.watching.add(provider)
				f.wg.Add(1)

				go f.watch(ctx, ch)
			}
		}
	}

	return nil
}

// Subscribe creates a subscription to Notification's.
func (f *Fido) Subscribe() <-chan Notification {
	ch := make(chan Notification)

	f.subscribers = append(f.subscribers, ch)

	return ch
}

// Close calls the close method on any providers that implement the optional CloseProvider optional
// extension interface. It will also close any subscriber channels that are currently open.
func (f *Fido) Close() error {
	for provider := range f.providers {
		if closer, ok := provider.(CloseProvider); ok {
			if err := closer.Close(); err != nil {
				return err
			}
		}
	}

	f.wg.Wait()

	for _, subscriber := range f.subscribers {
		close(subscriber)
	}

	return nil
}

// hydrate recursively populates the field map, mapping paths to struct fields.
func (f *Fido) hydrate(p Path, v reflect.Value) error {
	if len(p) == 0 {
		switch {
		case !v.IsValid():
			return ErrDestinationTypeInvalid
		case v.Type().Kind() != reflect.Ptr:
			return fmt.Errorf("%w: %s", ErrDestinationNotPtr, v.Type().Kind())
		case v.Type().Kind() == reflect.Ptr && v.IsNil():
			return ErrDestinationNil
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
			if errors.Is(err, ErrStructTagNotFound) && f.options.ErrorOnMissingTag {
				return fmt.Errorf("%w: failed parse struct tag for %s", err, ft.Name)
			}

			continue
		}

		if err := f.hydrate(append(p, tag.Name), fv); err != nil {
			return fmt.Errorf("%w: failed parse struct tag for %s", err, ft.Name)
		}
	}

	return nil
}

// publish pushes a notification to subscribers.
func (f *Fido) publish(notification Notification) {
	for _, ch := range f.subscribers {
		ch <- notification
	}
}

// watch continiously pulls values from the given channel until the context is complete or the
// channel is closed. If AutoUpdate is enabled fetch will be called for the Provider given on the
// channel reloading configuration values from that Provider.
func (f *Fido) watch(ctx context.Context, ch <-chan Provider) {
	defer f.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case provider, ok := <-ch:
			if !ok {
				return // Channel has closed
			}

			if f.options.AutoUpdate {
				if err := f.fetch(ctx, provider); err != nil {
					f.publish(&FieldUpdateError{
						Err: err,
					})
				}
			}
		}
	}
}

// fetch fetches values from the given provider. Update notifications are pumped onto an internal
// channel and passed to publish to be sent to notification subscribers.
// A named return value is used to catch and return a wrapped recover error on panic.
func (f *Fido) fetch(ctx context.Context, provider Provider) (err error) {
	defer func() {
		if r := recover(); r != nil {
			switch r := r.(type) {
			case error:
				err = fmt.Errorf("%w: recovered from panic", r)
			default:
				err = fmt.Errorf("%w: recovered from panic", NonErrPanic{Value: r})
			}
		}
	}()

	var updates FieldUpdates

	var (
		updatesCh = make(chan *FieldUpdate)
		doneCh    = make(chan struct{})
	)

	go func() {
		defer close(doneCh)

		for update := range updatesCh {
			updates = append(updates, update)
		}
	}()

	err = provider.Values(ctx, f.callback(provider, updatesCh))

	close(updatesCh)

	<-doneCh

	if err == nil {
		f.publish(updates)
	}

	return err
}

// callback returns the callback function gigven to a provider to call when it wishes to send
// a configuration value to Fido. It finds the destination struct field by the Path given and set
// that field to be the value of that of the one provided.
func (f *Fido) callback(provider Provider, updates chan<- *FieldUpdate) Callback {
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

			current := field.Value().Interface()

			if value != current {
				if field.Provider() != nil && f.options.EnforcePriority {
					if f.providers[field.Provider()] > f.providers[provider] {
						return nil
					}
				}

				if err := field.Set(value, provider); err != nil {
					return fmt.Errorf("%w: failed to set field %s value %v", err, path, value)
				}

				updates <- &FieldUpdate{
					Path:     path,
					New:      value,
					Old:      current,
					Provider: provider,
				}
			}

			return nil
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
