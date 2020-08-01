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

	Values(ctx context.Context, callback Callback) error
}

// A ReadProvider reads values from an io.Reader.
type ReadProvider interface {
	fmt.Stringer

	Values(ctx context.Context, reader io.Reader, callback Callback) error
}

// A PathProvider as an optional extension interface that if implemented by the Provider will allow
// Fido send the known key paths inferred from the destination struct tags to the provider.
type PathProvider interface {
	Paths(ch <-chan []string)
}

// A NotifyProvider is a an optional extension interface that if implemented by the Provider will
// allow Fido to send notifications of changed values whilst the application is running.
type NotifyProvider interface {
	Notify() <-chan error
}

// A CloseProvider is a an optional extension interface that if implemented by the Provider will
// allow Fido to call a close method on the Provider.
type CloseProvider interface {
	Close() error
}

// FromString constructs a new StringProvider.
func FromString(provider ReadProvider, value string) Provider {
	return &StringProvider{
		value:    value,
		provider: provider,
	}
}

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
func (s *StringProvider) Values(ctx context.Context, callback Callback) error {
	return s.provider.Values(ctx, strings.NewReader(s.value), callback)
}

// FromBytes constructs a new BytesProvider.
func FromBytes(provider ReadProvider, value []byte) Provider {
	return &BytesProvider{
		value:    value,
		provider: provider,
	}
}

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
func (s *BytesProvider) Values(ctx context.Context, callback Callback) error {
	return s.provider.Values(ctx, bytes.NewReader(s.value), callback)
}

type providers map[Provider]uint8

func (p providers) add(provider Provider) {
	p[provider] = uint8(len(p) + 1)
}

func (p providers) priority(provider Provider) uint8 {
	return p[provider]
}

func (p providers) exists(provider Provider) bool {
	_, ok := p[provider]

	return ok
}
