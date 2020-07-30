package inmemory

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/krak3n/fido"
)

func TestProvider(t *testing.T) {
	type want struct {
		path  fido.Path
		value interface{}
	}

	cases := map[string]struct {
		ctx    context.Context
		values map[string]interface{}
		want   []want
		err    error
	}{
		"ErrInvalidMapKey": {
			ctx: context.Background(),
			values: map[string]interface{}{
				"foo": map[int]string{
					0: "bar",
				},
			},
			want: []want{
				{
					path: []string{"foo"},
					value: map[int]string{
						0: "bar",
					},
				},
			},
			err: ErrInvalidMapKey,
		},
		"ErrInvalidMapValue": {
			ctx: context.Background(),
			values: map[string]interface{}{
				"foo": map[string]string{
					"foo": "bar",
				},
			},
			want: []want{
				{
					path: []string{"foo"},
					value: map[string]string{
						"foo": "bar",
					},
				},
			},
			err: ErrInvalidMapValue,
		},
		"ErrContextCancelled": {
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()

				return ctx
			}(),
			values: map[string]interface{}{
				"foo": map[string]interface{}{
					"bar": "fizz",
				},
			},
			want: []want{},
			err:  context.Canceled,
		},
		"SendsValues": {
			ctx: context.Background(),
			values: map[string]interface{}{
				"foo": map[string]interface{}{
					"bar": "fizz",
				},
			},
			want: []want{
				{
					path:  []string{"foo", "bar"},
					value: "fizz",
				},
			},
		},
	}

	for name, testCase := range cases {
		tc := testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var i int
			cb := func(p fido.Path, v interface{}) error {
				if i+1 > len(tc.want) {
					t.Fatal("received more than expected values")
				}

				want := tc.want[i]

				if !reflect.DeepEqual(want.path, p) {
					t.Errorf("want %+v path, got %+v", want.path, p)
				}

				if !reflect.DeepEqual(want.value, v) {
					t.Errorf("want %+v value, got %+v", want.value, v)
				}

				i++

				return nil
			}

			err := New(tc.values).Values(tc.ctx, cb)

			if !errors.Is(err, tc.err) {
				t.Errorf("want %+v error got %+v", tc.err, err)
			}
		})
	}
}
