package fido

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

func TestFido_Options(t *testing.T) {
	cases := map[string]struct {
		opts []Option
		want Options
	}{
		"Defaults": {
			want: DefaultOptions(),
		},
		"WithStructTag": {
			opts: []Option{
				WithStructTag("foo"),
			},
			want: Options{
				StructTag:         "foo",
				EnforcePriority:   true,
				ErrorOnMissingTag: true,
			},
		},
		"SetPriorityEnforcement": {
			opts: []Option{
				SetPriorityEnforcement(false),
			},
			want: Options{
				StructTag:         DefaultStructTag,
				EnforcePriority:   false,
				ErrorOnMissingTag: true,
			},
		},
		"SetErrorOnMissingTag": {
			opts: []Option{
				SetErrorOnMissingTag(false),
			},
			want: Options{
				StructTag:         DefaultStructTag,
				EnforcePriority:   true,
				ErrorOnMissingTag: false,
			},
		},
	}

	for name, testCase := range cases {
		tc := testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			cfg := struct{}{}

			f, err := New(&cfg, tc.opts...)
			if err != nil {
				t.Errorf("want nil error, got %+v", err)
			}

			if !reflect.DeepEqual(tc.want, f.options) {
				t.Errorf("want %+v, got %+v", tc.want, f.options)
			}
		})
	}
}

func TestFido_Fetch(t *testing.T) {
	var err = errors.New("boom")

	cases := map[string]struct {
		provider Provider
		err      error
	}{
		"ErrorPanicRecovery": {
			provider: NewTestProviderWithFunc(t, func(ctx context.Context, cb Callback) error {
				panic(err)
			}),
			err: err,
		},
		"NonErrorPanicRecovery": {
			provider: NewTestProviderWithFunc(t, func(ctx context.Context, cb Callback) error {
				panic("foo")
			}),
			err: NonErrPanic{Value: "foo"},
		},
	}

	for name, testCase := range cases {
		tc := testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			f := &Fido{
				providers: make(providers),
			}

			err := f.Fetch(tc.provider)

			if !errors.Is(err, tc.err) {
				t.Errorf("want %+v, got %+v", err, tc.err)
			}
		})
	}
}

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

func TestFido_hydrate(t *testing.T) {
	type Config struct {
		Foo string `fido:"foo"`
		Bar string
	}

	cases := map[string]struct {
		value   reflect.Value
		options Options
		err     error
	}{
		"MustBeValid": {
			value: reflect.ValueOf(nil),
			err:   ErrDestinationTypeInvalid,
		},
		"MustBePointer": {
			value: func() reflect.Value {
				var v Config

				return reflect.ValueOf(v)
			}(),
			err: ErrDestinationNotPtr,
		},
		"MustBeStruct": {
			value: func() reflect.Value {
				var v map[string]string

				return reflect.ValueOf(&v)
			}(),
			err: ErrDestinationTypeInvalid,
		},
		"MustNotBeNil": {
			value: func() reflect.Value {
				var v interface{} = (*Config)(nil)

				return reflect.ValueOf(v)
			}(),
			err: ErrDestinationNil,
		},
		"TagNotFoundError": {
			value: reflect.ValueOf(&Config{}),
			options: Options{
				StructTag:         DefaultStructTag,
				ErrorOnMissingTag: true,
			},
			err: ErrStructTagNotFound,
		},
		"NoTagNotFoundError": {
			value: reflect.ValueOf(&Config{}),
			options: Options{
				StructTag:         DefaultStructTag,
				ErrorOnMissingTag: false,
			},
		},
	}

	for name, testCase := range cases {
		tc := testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			f := &Fido{
				fields:  make(fields),
				options: tc.options,
			}

			err := f.hydrate(Path{}, tc.value)

			if !errors.Is(err, tc.err) {
				t.Errorf("want %+v, got %+v", tc.err, err)
			}
		})
	}
}

func TestFido_callback(t *testing.T) {
	p1 := NewTestProvider(t)
	p2 := NewTestProvider(t)

	cases := map[string]struct {
		path      Path
		value     interface{}
		options   Options
		field     *field
		providers providers
		provider  Provider
		want      interface{}
		err       error
	}{
		"FieldNotFound": {
			path:  Path{"bar"},
			value: "foo",
			field: &field{
				path:     Path{"foo"},
				value:    reflect.New(reflect.TypeOf("")).Elem(),
				provider: p2,
			},
			options: Options{
				ErrorOnFieldNotFound: true,
			},
			want: "",
			err:  ErrFieldNotFound,
		},
		"IgnoreFieldNotFound": {
			path:  Path{"bar"},
			value: "foo",
			field: &field{
				path:     Path{"foo"},
				value:    reflect.New(reflect.TypeOf("")).Elem(),
				provider: p2,
			},
			options: Options{
				ErrorOnFieldNotFound: false,
			},
			want: "",
		},
		"EnforceProviderPriority": {
			path:  Path{"foo"},
			value: "bar",
			field: &field{
				path:     Path{"foo"},
				value:    reflect.New(reflect.TypeOf("")).Elem(),
				provider: p2,
			},
			providers: providers{
				p1: 1,
				p2: 2,
			},
			provider: p1,
			options: Options{
				EnforcePriority: true,
			},
			want: "",
		},
		"InitMapFieldError": {
			path:  Path{"foo", "bar"},
			value: "baz",
			field: &field{
				path: Path{"foo"},
				value: func() reflect.Value {
					var m map[int]string

					return reflect.New(reflect.TypeOf(m)).Elem()
				}(),
			},
			providers: providers{
				p1: 1,
			},
			provider: p1,
			err:      ErrInvalidMapKeyType,
			want: func() interface{} {
				var m map[int]string

				return m
			}(),
		},
		"InitMapField": {
			path:  Path{"foo", "bar"},
			value: "baz",
			field: &field{
				path: Path{"foo"},
				value: func() reflect.Value {
					var m map[string]string

					return reflect.New(reflect.TypeOf(m)).Elem()
				}(),
			},
			providers: providers{
				p1: 1,
			},
			provider: p1,
			want: map[string]string{
				"bar": "baz",
			},
		},
		"SetValue": {
			path:  Path{"foo"},
			value: "bar",
			field: &field{
				path:     Path{"foo"},
				value:    reflect.New(reflect.TypeOf("")).Elem(),
				provider: p1,
			},
			providers: providers{
				p1: 1,
				p2: 2,
			},
			provider: p2,
			options: Options{
				EnforcePriority: true,
			},
			want: "bar",
		},
	}

	for name, testCase := range cases {
		tc := testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			f := &Fido{
				options: tc.options,
				fields: fields{
					tc.field.path.key(): tc.field,
				},
				providers: tc.providers,
			}

			err := f.callback(tc.provider)(tc.path, tc.value)

			if !errors.Is(err, tc.err) {
				t.Errorf("want %+v, got %+v", tc.err, err)
			}

			if !reflect.DeepEqual(tc.want, tc.field.value.Interface()) {
				t.Errorf("want %+v, got %+v", tc.want, tc.field.value.Interface())
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
