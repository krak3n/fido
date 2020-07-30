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
