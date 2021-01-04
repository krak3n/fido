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
	separator     string
	values        map[string]interface{}
	notifications []chan string
}

// New constructs a new Provider.
func New(opts ...Option) *Provider {
	p := &Provider{
		separator:     ".",
		values:        make(map[string]interface{}),
		notifications: make([]chan string, 0),
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

	for _, ch := range p.notifications {
		ch <- path
	}
}

// Values walks the in memory values map calling the callback function passing the path and value to
// Fido for processing.
func (p *Provider) Values(ctx context.Context, writer fido.Writer) error {
	for k, v := range p.values {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			path := fido.Path(strings.Split(k, p.separator))

			if err := writer.Write(path, v); err != nil {
				return err
			}
		}
	}

	return nil
}

// Notify implements the optional NotifyProvider extension interface sending notifications of
// changes to configuration values handled by this provider. This blocks until Close is called.
func (p *Provider) Notify(ctx context.Context, writer fido.Writer) error {
	ch := make(chan string)

	p.notifications = append(p.notifications, ch)

	for path := range ch {
		if err := writer.Write(fido.Path(strings.Split(path, p.separator)), p.values[path]); err != nil {
			return err
		}
	}

	return nil
}

// Close implements the optional NotifyCloser extension interface closing any notification channels
// currently sending notifications back to Fido of value changes.
func (p *Provider) Close() error {
	for _, ch := range p.notifications {
		close(ch)
	}

	return nil
}
