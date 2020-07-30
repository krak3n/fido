package fido

import (
	"context"
	"testing"
)

type TestValue struct {
	Path  []string
	Value interface{}
	Err   error
}

type TestProvider struct {
	t      *testing.T
	values []TestValue
}

func (t *TestProvider) String() string {
	return "test"
}

func (t *TestProvider) Values(ctx context.Context, callback Callback) error {
	for _, v := range t.values {
		if v.Err != nil {
			return v.Err
		}

		if err := callback(v.Path, v.Value); err != nil {
			return err
		}
	}

	return nil
}

func (t *TestProvider) Add(path []string, value interface{}, err error) {
	t.values = append(t.values, TestValue{path, value, err})
}

func NewTestProvider(t *testing.T) *TestProvider {
	return &TestProvider{
		t:      t,
		values: make([]TestValue, 0),
	}
}

func Test_providers_priority(t *testing.T) {
	provider1 := NewTestProvider(t)
	provider2 := NewTestProvider(t)

	cases := map[string]struct {
		providers providers
		provider  Provider
		want      uint8
	}{
		"NotExistsReturns0": {
			providers: make(providers),
			provider:  provider1,
			want:      uint8(0),
		},
		"ExistsReturnsPriority": {
			providers: func() providers {
				p := make(providers)
				p.add(provider1)
				p.add(provider2)

				return p
			}(),
			provider: provider2,
			want:     uint8(2),
		},
	}

	for name, testCase := range cases {
		tc := testCase

		t.Run(name, func(t *testing.T) {
			priority := tc.providers.priority(tc.provider)

			if tc.want != priority {
				t.Errorf("want %+v tag, got %+v", tc.want, priority)
			}
		})
	}
}
