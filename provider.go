package fido

import (
	"context"
	"fmt"
)

// A Provider sends values to Fido for processing.
type Provider interface {
	fmt.Stringer

	Values(ctx context.Context, callback func([]string, interface{})) error
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
