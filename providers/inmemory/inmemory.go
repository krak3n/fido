package inmemory

import (
	"context"
	"fmt"
	"reflect"
)

// ProviderName is the name of the provider.
const ProviderName = "inmemory"

// Errors that this Provider provides.
const (
	ErrInvalidMapKey Error = iota + 1
	ErrInvalidMapValue
)

// A Error is a sentinel error.
type Error uint8

func (e Error) Error() string {
	switch e {
	case ErrInvalidMapKey:
		return "invalid map key"
	case ErrInvalidMapValue:
		return "invalid map value"
	}

	return "unknown error"
}

// Provider implements a inmemory fido.Provider.
type Provider struct {
	values map[string]interface{}
}

// New constructs a new Provider.
func New(values map[string]interface{}) *Provider {
	return &Provider{
		values: values,
	}
}

func (p *Provider) String() string {
	return ProviderName
}

// Values walks the in memory values map calling the callback function passing the path and value to
// Fido for processing.
func (p *Provider) Values(ctx context.Context, cb func([]string, interface{})) error {
	return p.walk(ctx, []string{}, p.values, cb)
}

func (p *Provider) walk(ctx context.Context, path []string, values map[string]interface{}, cb func([]string, interface{})) error {
	for k, v := range values {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if rv := reflect.ValueOf(v); rv.Kind() == reflect.Map {
				kk := rv.Type().Key().Kind()
				vk := rv.Type().Elem().Kind()

				switch {
				case kk != reflect.String:
					return fmt.Errorf("%w: invalid map key type %s", ErrInvalidMapKey, kk)
				case vk != reflect.Interface:
					return fmt.Errorf("%w: invalid map value type %s", ErrInvalidMapValue, vk)
				default:
					if err := p.walk(ctx, append(path, k), v.(map[string]interface{}), cb); err != nil {
						return err
					}

					continue
				}
			}

			cb(append(path, k), v)
		}
	}

	return nil
}
