package fido

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
)

// Generic provider names
const (
	StringProviderName = "String"
	BytesProviderName  = "Bytes"
)

// JoinProviderNames joins multiple provider names into one.
func JoinProviderNames(names ...string) string {
	return strings.Join(names, ".")
}

// A Provider sends values to Fido for processing.
type Provider interface {
	fmt.Stringer

	Values(ctx context.Context, writer Writer) error
}

// A CloseProvider is a an optional extension interface that if implemented by the Provider will
// allow Fido to call a close method on the Provider.
type CloseProvider interface {
	Close() error
}

// A NotifyProvider is a an optional extension interface that if implemented by the Provider will
// allow Fido watch the Provider for changes to configuration and pass those changes onto
// Subscribers. If configured to do so Fido can also reload configuration when a notified of
// changes.
type NotifyProvider interface {
	CloseProvider

	Notify(ctx context.Context, writer Writer) error
}

// A ReadProvider reads values from an io.Reader.
type ReadProvider interface {
	fmt.Stringer

	Values(ctx context.Context, reader io.Reader, writer Writer) error
}

// A PathProvider as an optional extension interface that if implemented by the Provider will allow
// Fido send the known key paths inferred from the destination struct tags to the provider.
type PathProvider interface {
	Paths(ch <-chan []string)
}

// FromString constructs a new StringProvider.
func FromString(provider ReadProvider, value string) *StringProvider {
	return &StringProvider{
		value:    value,
		provider: provider,
	}
}

// Ensure StringProvider implements the Provider interface.
var _ Provider = (*StringProvider)(nil)

// StringProvider wraps a ReadProvider implementing the standard Provider interface.
type StringProvider struct {
	value    string
	provider ReadProvider
}

func (s *StringProvider) String() string {
	return JoinProviderNames(s.provider.String(), StringProviderName)
}

// Values calls the Values function on the wrapped provider passing it the context, the string value as
// an io.Reader and the callback function.
func (s *StringProvider) Values(ctx context.Context, writer Writer) error {
	return s.provider.Values(ctx, strings.NewReader(s.value), writer)
}

// FromBytes constructs a new BytesProvider.
func FromBytes(provider ReadProvider, value []byte) *BytesProvider {
	return &BytesProvider{
		value:    value,
		provider: provider,
	}
}

// Ensure BytesProvider implements the Provider interface.
var _ Provider = (*BytesProvider)(nil)

// BytesProvider wraps a ReadProvider implementing the standard Provider interface.
type BytesProvider struct {
	value    []byte
	provider ReadProvider
}

func (s *BytesProvider) String() string {
	return JoinProviderNames(s.provider.String(), BytesProviderName)
}

// Values calls the Values function on the wrapped provider passing it the context, the string value as
// an io.Reader and the callback function.
func (s *BytesProvider) Values(ctx context.Context, writer Writer) error {
	return s.provider.Values(ctx, bytes.NewReader(s.value), writer)
}

type providers map[Provider]uint8

func (p providers) add(items ...Provider) {
	for _, provider := range items {
		if _, ok := p[provider]; !ok {
			p[provider] = uint8(len(p) + 1)
		}
	}
}

func (p providers) priority(provider Provider) uint8 {
	return p[provider]
}
