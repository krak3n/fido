package inmemory_test

import (
	"fmt"

	"github.com/krak3n/fido"
	"github.com/krak3n/fido/providers/inmemory"
)

func Example() {
	type Config struct {
		Foo  string            `fido:"foo"`
		Fizz map[string]string `fido:"fizz"`
	}

	var cfg Config

	provider := inmemory.New()
	provider.Add("foo", "bar")
	provider.Add("fizz.buzz", "bazz")

	f, err := fido.New(&cfg)
	if err != nil {
		fmt.Println(err)
	}

	defer f.Close()

	if err := f.Fetch(provider); err != nil {
		fmt.Println(err)
	}

	fmt.Printf("%+v", cfg)
	// Output:
	// {Foo:bar Fizz:map[buzz:bazz]}
}
