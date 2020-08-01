package fido_test

import (
	"context"
	"fmt"

	"github.com/krak3n/fido"
	"github.com/krak3n/fido/providers/inmemory"
)

func ExampleFido_Fetch() {
	type Config struct {
		Foo string `fido:"foo"`
	}

	var cfg Config

	provider := inmemory.New()
	provider.Add("foo", "bar")

	f, err := fido.New(&cfg)
	if err != nil {
		fmt.Println(err)
	}

	if err := f.Fetch(provider); err != nil {
		fmt.Println(err)
	}

	fmt.Println(cfg.Foo)
	// Output:
	// bar
}

func ExampleFido_FetchWithContext_canceled() {
	type Config struct {
		Foo string `fido:"foo"`
	}

	var cfg Config

	provider := inmemory.New()
	provider.Add("foo", "bar")

	f, err := fido.New(&cfg)
	if err != nil {
		fmt.Println(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := f.FetchWithContext(ctx, provider); err != nil {
		fmt.Println(err)
	}
	// Output:
	// context canceled
}
