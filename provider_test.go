package fido

import (
	"context"
	"errors"
	"io"
	"io/ioutil"
	"reflect"
	"strings"
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
	fn     func(context.Context, Callback) error
}

func (t *TestProvider) String() string {
	return "TestProvider"
}

func (t *TestProvider) Values(ctx context.Context, callback Callback) error {
	if t.fn == nil {
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

	return t.fn(ctx, callback)
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

func NewTestProviderWithFunc(t *testing.T, fn func(context.Context, Callback) error) *TestProvider {
	return &TestProvider{
		t:      t,
		values: make([]TestValue, 0),
		fn:     fn,
	}
}

type TestReadProvider struct {
	t  *testing.T
	fn func(context.Context, io.Reader, Callback) error
}

func (t *TestReadProvider) String() string {
	return "TestReadProvider"
}

func (t *TestReadProvider) Values(ctx context.Context, reader io.Reader, callback Callback) error {
	if t.fn == nil {
		t.t.Error("values function defined")
		return nil
	}

	return t.fn(ctx, reader, callback)
}

func NewTestReadProvider(t *testing.T, fn func(context.Context, io.Reader, Callback) error) *TestReadProvider {
	return &TestReadProvider{
		t:  t,
		fn: fn,
	}
}

func TestStringProvider(t *testing.T) {
	cases := map[string]struct {
		provider func(*testing.T) ReadProvider
		callback func(*testing.T) Callback
		err      error
	}{
		"PassesValues": {
			provider: func(t *testing.T) ReadProvider {
				return NewTestReadProvider(t, func(ctx context.Context, reader io.Reader, callback Callback) error {
					b, err := ioutil.ReadAll(reader)
					if err != nil {
						return err
					}

					parts := strings.Split(string(b), ":")

					return callback(Path{parts[0]}, parts[1])
				})
			},
			callback: func(t *testing.T) Callback {
				return func(path Path, value interface{}) error {
					{
						want := Path{"foo"}
						if !reflect.DeepEqual(path, want) {
							t.Errorf("want %+v, got %+v", want, path)
						}
					}

					{
						want := "bar"
						if !reflect.DeepEqual(value, want) {
							t.Errorf("want %+v, got %+v", want, value)
						}
					}

					return nil
				}
			},
		},
	}

	for name, testCase := range cases {
		tc := testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			rp := tc.provider(t)

			provider := FromString(rp, "foo:bar")
			if provider.String() != JoinProviderNames(rp.String(), StringProviderName) {
				t.Errorf("invalid provider name, got %+v", provider.String())
			}

			err := provider.Values(context.Background(), tc.callback(t))
			if !errors.Is(err, tc.err) {
				t.Errorf("want %+v, got %+v", tc.err, err)
			}
		})
	}
}

func TestBytesProvider(t *testing.T) {
	cases := map[string]struct {
		provider func(*testing.T) ReadProvider
		callback func(*testing.T) Callback
		err      error
	}{
		"PassesValues": {
			provider: func(t *testing.T) ReadProvider {
				return NewTestReadProvider(t, func(ctx context.Context, reader io.Reader, callback Callback) error {
					b, err := ioutil.ReadAll(reader)
					if err != nil {
						return err
					}

					parts := strings.Split(string(b), ":")

					return callback(Path{parts[0]}, parts[1])
				})
			},
			callback: func(t *testing.T) Callback {
				return func(path Path, value interface{}) error {
					{
						want := Path{"foo"}
						if !reflect.DeepEqual(path, want) {
							t.Errorf("want %+v, got %+v", want, path)
						}
					}

					{
						want := "bar"
						if !reflect.DeepEqual(value, want) {
							t.Errorf("want %+v, got %+v", want, value)
						}
					}

					return nil
				}
			},
		},
	}

	for name, testCase := range cases {
		tc := testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			rp := tc.provider(t)

			provider := FromBytes(rp, []byte("foo:bar"))
			if provider.String() != JoinProviderNames(rp.String(), BytesProviderName) {
				t.Errorf("invalid provider name, got %+v", provider.String())
			}

			err := provider.Values(context.Background(), tc.callback(t))
			if !errors.Is(err, tc.err) {
				t.Errorf("want %+v, got %+v", tc.err, err)
			}
		})
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
