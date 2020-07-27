package fido

import (
	"context"
	"fmt"
)

type Fido struct {
}

func New(dst interface{}) *Fido {
	return nil
}

func (f *Fido) Fetch(providers ...Provider) error {
	return f.FetchWithContext(context.Background(), providers...)
}

func (f *Fido) FetchWithContext(ctx context.Context, providers ...Provider) error {
	for _, provider := range providers {

		if _, ok := provider.(PathProvider); ok {
			fmt.Println("I support receiving paths")
		}

		if _, ok := provider.(NotifyProvider); ok {
			fmt.Println("I support notifications")
		}

		closer, ok := provider.(CloseProvider)
		if ok {
			defer closer.Close()
		}

		fn := func(path []string, value interface{}) {
			fmt.Println(path, value)
		}

		if err := provider.Values(ctx, fn); err != nil {
			return err
		}
	}

	return nil
}
