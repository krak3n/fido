package fido

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

func TestFido_fetch(t *testing.T) {
	cases := map[string]struct {
		fido     *Fido
		provider Provider
		err      error
	}{
		"ErrFieldNotFound": {
			fido: &Fido{
				options: Options{
					ErrorOnFieldNotFound: true,
				},
				providers: make(providers),
				fields:    make(fields),
			},
			provider: func() Provider {
				p := NewTestProvider(t)
				p.Add([]string{"foo"}, "bar", nil)

				return p
			}(),
			err: ErrFieldNotFound,
		},
		"NoErrFieldNotFound": {
			fido: &Fido{
				options:   Options{},
				providers: make(providers),
				fields:    make(fields),
			},
			provider: func() Provider {
				p := NewTestProvider(t)
				p.Add([]string{"foo"}, "bar", nil)

				return p
			}(),
		},
	}

	for name, testCase := range cases {
		tc := testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := tc.fido.fetch(context.Background(), tc.provider)

			if !errors.Is(err, tc.err) {
				t.Errorf("want %+v err, got %+v", err, tc.err)
			}
		})
	}
}

func TestFido_initMap(t *testing.T) {
	cases := map[string]struct {
		path  []string
		value reflect.Value
		want  reflect.Value
		err   error
	}{
		"ErrExpectedMap": {
			value: reflect.ValueOf("foo"),
			want:  reflect.ValueOf("foo"),
			err:   ErrExpectedMap,
		},
		"InitialiseNilMapInvalidKeyType": {
			value: func() reflect.Value {
				var m map[int]string

				return reflect.New(reflect.TypeOf(m)).Elem()
			}(),
			want: func() reflect.Value {
				var m map[int]string

				return reflect.ValueOf(m)
			}(),
			err: ErrInvalidMapKeyType,
		},
		"InitialiseNilMapNotAddressableValue": {
			value: func() reflect.Value {
				var m map[string]string

				return reflect.ValueOf(m)
			}(),
			want: func() reflect.Value {
				var m map[string]string

				return reflect.ValueOf(m)
			}(),
			err: ErrReflectValueNotAddressable,
		},
		"InitialisesNilMap": {
			value: func() reflect.Value {
				var m map[string]string

				return reflect.New(reflect.TypeOf(m)).Elem()
			}(),
			want: reflect.ValueOf(map[string]string{}),
		},
		"InitialisedMap": {
			value: reflect.ValueOf(map[string]string{}),
			want:  reflect.ValueOf(map[string]string{}),
		},
		"InitialiseNestedMapInvalidPath": {
			path:  []string{"foo"},
			value: reflect.ValueOf(map[string]map[string]map[string]string{}),
			want:  reflect.ValueOf(map[string]map[string]map[string]string{}),
			err:   ErrInvalidPath,
		},
		"InitialisedNestedMap": {
			path: []string{"foo", "bar", "baz"},
			value: func() reflect.Value {
				var m map[string]map[string]map[string]string

				return reflect.New(reflect.TypeOf(m)).Elem()
			}(),
			want: reflect.ValueOf(map[string]string{}),
		},
	}

	for name, testCase := range cases {
		tc := testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			f := &Fido{}
			value, err := f.initMap(tc.path, tc.value)

			if !errors.Is(err, tc.err) {
				t.Errorf("want %+v err, got %+v", err, tc.err)
			}

			if !reflect.DeepEqual(tc.want.Interface(), value.Interface()) {
				t.Errorf("want %+v value, got %+v", tc.want.Interface(), value.Interface())
			}
		})
	}
}
