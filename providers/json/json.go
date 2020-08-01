package json

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/krak3n/fido"
)

// ProviderName is the name of the Provider.
const ProviderName = "json"

// Provider provides a JSON fido.ReadProvider.
type Provider struct{}

// New constructs a new Provider.
func New() *Provider {
	return &Provider{}
}

func (p *Provider) String() string {
	return ProviderName
}

// Values reads json from the given io.Reader passing the values back to Fido for processing.
func (p *Provider) Values(ctx context.Context, reader io.Reader, callback fido.Callback) error {
	b, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}

	var dst map[string]interface{}

	if err := json.Unmarshal(b, &dst); err != nil {
		return fmt.Errorf("%w: failed to unmarshal JSON: '%s'", err, string(b))
	}

	return fido.WalkMap(ctx, dst, fido.Path{}, callback)
}
