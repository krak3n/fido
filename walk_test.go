package fido

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

func TestWalkMap(t *testing.T) {
	src := map[string]interface{}{
		"foo": "bar",
		"fizz": map[string]interface{}{
			"buzz": "fuzz",
		},
	}

	cases := map[string]struct {
		ctx      context.Context
		callback func(*testing.T, map[string]interface{}) Callback
		want     map[string]interface{}
		err      error
	}{
		"ContextCanceled": {
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				return ctx
			}(),
			callback: func(t *testing.T, _ map[string]interface{}) Callback {
				return func(Path, interface{}) error {
					return nil
				}
			},
			want: map[string]interface{}{},
			err:  context.Canceled,
		},
		"CallbackError": {
			ctx: context.Background(),
			callback: func(t *testing.T, _ map[string]interface{}) Callback {
				var i int
				return func(Path, interface{}) error {
					defer func() {
						i++
					}()

					switch i {
					case 0:
						return nil
					case 1:
						return ErrSetInvalidValue
					}

					return nil
				}
			},
			want: map[string]interface{}{},
			err:  ErrSetInvalidValue,
		},
		"PassesValuesToCallback": {
			ctx: context.Background(),
			callback: func(t *testing.T, m map[string]interface{}) Callback {
				return func(path Path, value interface{}) error {
					m[path.key()] = value

					return nil
				}
			},
			want: map[string]interface{}{
				"foo":       "bar",
				"fizz.buzz": "fuzz",
			},
			err: nil,
		},
	}

	for name, testCase := range cases {
		tc := testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			values := map[string]interface{}{}

			err := WalkMap(tc.ctx, src, Path{}, tc.callback(t, values))
			if !errors.Is(err, tc.err) {
				t.Errorf("want %+v, got %+v", tc.err, err)
			}

			for k, v := range tc.want {
				value, ok := values[k]
				if !ok {
					t.Errorf("%s key not found in values", k)
					continue
				}

				if !reflect.DeepEqual(v, value) {
					t.Errorf("want %+v, got %+v", v, value)
				}
			}
		})
	}
}
