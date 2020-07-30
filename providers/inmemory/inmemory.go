package inmemory

import (
	"context"
	"strings"

	"github.com/krak3n/fido"
)

// ProviderName is the name of the provider.
const ProviderName = "inmemory"

// An Option configures provider behaviour.
type Option interface {
	apply(*Provider)
}

// OptionFunc is an adapter allowing regular methods to act as Option's.
type OptionFunc func(*Provider)

func (fn OptionFunc) apply(p *Provider) {
	fn(p)
}

// WithSeparator configures the path separator used when sending paths to Fido.
func WithSeparator(d string) Option {
	return OptionFunc(func(p *Provider) {
		p.separator = d
	})
}

// Provider implements a inmemory fido.Provider.
type Provider struct {
	values    map[string]interface{}
	separator string
}

// New constructs a new Provider.
func New(opts ...Option) *Provider {
	p := &Provider{
		values:    make(map[string]interface{}),
		separator: ".",
	}

	for _, opt := range opts {
		opt.apply(p)
	}

	return p
}

func (p *Provider) String() string {
	return ProviderName
}

// Add adds a value to the inmemory provider for the given path.
func (p *Provider) Add(path string, value interface{}) {
	p.values[path] = value
}

// Values walks the in memory values map calling the callback function passing the path and value to
// Fido for processing.
func (p *Provider) Values(ctx context.Context, callback fido.Callback) error {
	for k, v := range p.values {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			path := fido.Path(strings.Split(k, p.separator))

			if err := callback(path, v); err != nil {
				return err
			}
		}
	}

	return nil
}
