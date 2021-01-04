package main

import (
	"log"

	"github.com/krak3n/fido"
	"github.com/krak3n/fido/providers/json"
)

type Config struct {
	Foo  string            `fido:"foo"`
	Fizz map[string]string `fido:"fizz"`
}

func main() {
	var cfg Config

	f, err := fido.New(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	f.Add(fido.FromFiles(json.New(), "cfg.json", "*.json"))

	if err := f.Fetch(); err != nil {
		log.Println(err)
	}

	log.Println(cfg)
}
