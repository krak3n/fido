package fido

import (
	"errors"
	"reflect"
	"testing"
)

type TestStringer struct {
	Value string
}

func (t TestStringer) String() string {
	return t.Value
}

func Test_setValueToString(t *testing.T) {
	cases := map[string]struct {
		to   interface{}
		dst  reflect.Value
		want string
		err  error
	}{
		"NotSetable": {
			to: []string{"foo"},
			dst: func() reflect.Value {
				var v string

				return reflect.ValueOf(v)
			}(),
			err: ErrReflectValueNotSetable,
		},
		"InvalidType": {
			to: []string{"foo"},
			dst: func() reflect.Value {
				var v string

				return reflect.ValueOf(&v).Elem()
			}(),
			err: ErrSetInvalidType,
		},
		"Stringer": {
			to: TestStringer{"foo"},
			dst: func() reflect.Value {
				var v string

				return reflect.ValueOf(&v).Elem()
			}(),
			want: "foo",
		},
		"String": {
			to: "foo",
			dst: func() reflect.Value {
				var v string

				return reflect.ValueOf(&v).Elem()
			}(),
			want: "foo",
		},
		"Bool": {
			to: true,
			dst: func() reflect.Value {
				var v string

				return reflect.ValueOf(&v).Elem()
			}(),
			want: "true",
		},
		"Int": {
			to: int(1),
			dst: func() reflect.Value {
				var v string

				return reflect.ValueOf(&v).Elem()
			}(),
			want: "1",
		},
		"Int8": {
			to: int8(8),
			dst: func() reflect.Value {
				var v string

				return reflect.ValueOf(&v).Elem()
			}(),
			want: "8",
		},
		"Int16": {
			to: int16(16),
			dst: func() reflect.Value {
				var v string

				return reflect.ValueOf(&v).Elem()
			}(),
			want: "16",
		},
		"Int32": {
			to: int32(32),
			dst: func() reflect.Value {
				var v string

				return reflect.ValueOf(&v).Elem()
			}(),
			want: "32",
		},
		"Int64": {
			to: int64(64),
			dst: func() reflect.Value {
				var v string

				return reflect.ValueOf(&v).Elem()
			}(),
			want: "64",
		},
		"Uint": {
			to: uint(1),
			dst: func() reflect.Value {
				var v string

				return reflect.ValueOf(&v).Elem()
			}(),
			want: "1",
		},
		"Uint8": {
			to: uint8(8),
			dst: func() reflect.Value {
				var v string

				return reflect.ValueOf(&v).Elem()
			}(),
			want: "8",
		},
		"Uint16": {
			to: uint16(16),
			dst: func() reflect.Value {
				var v string

				return reflect.ValueOf(&v).Elem()
			}(),
			want: "16",
		},
		"Uint32": {
			to: uint32(32),
			dst: func() reflect.Value {
				var v string

				return reflect.ValueOf(&v).Elem()
			}(),
			want: "32",
		},
		"Uint64": {
			to: uint64(64),
			dst: func() reflect.Value {
				var v string

				return reflect.ValueOf(&v).Elem()
			}(),
			want: "64",
		},
	}

	for name, testCase := range cases {
		tc := testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := setValueToString(tc.dst, tc.to)

			if !errors.Is(err, tc.err) {
				t.Errorf("want %+v err, got %+v", err, tc.err)
			}

			if !reflect.DeepEqual(tc.want, tc.dst.String()) {
				t.Errorf("want %+v value, got %+v", tc.want, tc.dst.String())
			}
		})
	}
}

func Test_setValueToInt(t *testing.T) {
	cases := map[string]struct {
		to   interface{}
		dst  reflect.Value
		want int64
		err  error
	}{
		"NotSetable": {
			to: []string{"foo"},
			dst: func() reflect.Value {
				var v int64

				return reflect.ValueOf(v)
			}(),
			err: ErrReflectValueNotSetable,
		},
		"InvalidType": {
			to: []string{"foo"},
			dst: func() reflect.Value {
				var v int64

				return reflect.ValueOf(&v).Elem()
			}(),
			err: ErrSetInvalidType,
		},
		"InvalidSyntax": {
			to: "foo",
			dst: func() reflect.Value {
				var v int64

				return reflect.ValueOf(&v).Elem()
			}(),
			err: ErrSetInvalidValue,
		},
		"Overflow": {
			to: int64(1 << 31),
			dst: func() reflect.Value {
				var v int32

				return reflect.ValueOf(&v).Elem()
			}(),
			err:  ErrSetOverflow,
			want: 0,
		},
		"String": {
			to: "1",
			dst: func() reflect.Value {
				var v int64

				return reflect.ValueOf(&v).Elem()
			}(),
			want: 1,
		},
		"Int": {
			to: int(1),
			dst: func() reflect.Value {
				var v int64

				return reflect.ValueOf(&v).Elem()
			}(),
			want: 1,
		},
		"Int8": {
			to: int8(8),
			dst: func() reflect.Value {
				var v int64

				return reflect.ValueOf(&v).Elem()
			}(),
			want: 8,
		},
		"Int16": {
			to: int16(16),
			dst: func() reflect.Value {
				var v int64

				return reflect.ValueOf(&v).Elem()
			}(),
			want: 16,
		},
		"Int32": {
			to: int32(32),
			dst: func() reflect.Value {
				var v int64

				return reflect.ValueOf(&v).Elem()
			}(),
			want: 32,
		},
		"Int64": {
			to: int64(64),
			dst: func() reflect.Value {
				var v int64

				return reflect.ValueOf(&v).Elem()
			}(),
			want: 64,
		},
	}

	for name, testCase := range cases {
		tc := testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := setValueToInt(tc.dst, tc.to)

			if !errors.Is(err, tc.err) {
				t.Errorf("want %+v err, got %+v", tc.err, err)
			}

			if !reflect.DeepEqual(tc.want, tc.dst.Int()) {
				t.Errorf("want %+v int64, got %+v", tc.want, tc.dst.Int())
			}
		})
	}
}

func Test_setValueToUint(t *testing.T) {
	cases := map[string]struct {
		to   interface{}
		dst  reflect.Value
		want uint64
		err  error
	}{
		"NotSetable": {
			to: []string{"foo"},
			dst: func() reflect.Value {
				var v uint64

				return reflect.ValueOf(v)
			}(),
			err: ErrReflectValueNotSetable,
		},
		"InvalidType": {
			to: []string{"foo"},
			dst: func() reflect.Value {
				var v uint64

				return reflect.ValueOf(&v).Elem()
			}(),
			err: ErrSetInvalidType,
		},
		"InvalidSyntax": {
			to: "foo",
			dst: func() reflect.Value {
				var v uint64

				return reflect.ValueOf(&v).Elem()
			}(),
			err: ErrSetInvalidValue,
		},
		"Overflow": {
			to: uint64(1 << 32),
			dst: func() reflect.Value {
				var v uint32

				return reflect.ValueOf(&v).Elem()
			}(),
			err:  ErrSetOverflow,
			want: 0,
		},
		"String": {
			dst: func() reflect.Value {
				var v uint64

				return reflect.ValueOf(&v).Elem()
			}(),
			to:   "1",
			want: 1,
		},
		"Uint": {
			dst: func() reflect.Value {
				var v uint64

				return reflect.ValueOf(&v).Elem()
			}(),
			to:   uint(1),
			want: 1,
		},
		"Uint8": {
			dst: func() reflect.Value {
				var v uint64

				return reflect.ValueOf(&v).Elem()
			}(),
			to:   uint8(8),
			want: 8,
		},
		"Uint16": {
			dst: func() reflect.Value {
				var v uint64

				return reflect.ValueOf(&v).Elem()
			}(),
			to:   uint16(16),
			want: 16,
		},
		"Uint32": {
			dst: func() reflect.Value {
				var v uint64

				return reflect.ValueOf(&v).Elem()
			}(),
			to:   uint32(32),
			want: 32,
		},
		"Uint64": {
			dst: func() reflect.Value {
				var v uint64

				return reflect.ValueOf(&v).Elem()
			}(),
			to:   uint64(64),
			want: 64,
		},
	}

	for name, testCase := range cases {
		tc := testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := setValueToUint(tc.dst, tc.to)

			if !errors.Is(err, tc.err) {
				t.Errorf("want %+v err, got %+v", err, tc.err)
			}

			if !reflect.DeepEqual(tc.want, tc.dst.Uint()) {
				t.Errorf("want %+v uint64, got %+v", tc.want, tc.dst.Uint())
			}
		})
	}
}

func Test_setValueToFloat(t *testing.T) {
	cases := map[string]struct {
		to   interface{}
		dst  reflect.Value
		want float64
		err  error
	}{
		"NotSetable": {
			to: []string{"foo"},
			dst: func() reflect.Value {
				var v float64

				return reflect.ValueOf(v)
			}(),
			err: ErrReflectValueNotSetable,
		},
		"InvalidType": {
			to: []string{"foo"},
			dst: func() reflect.Value {
				var v float64

				return reflect.ValueOf(&v).Elem()
			}(),
			err: ErrSetInvalidType,
		},
		"Overflow": {
			to: float64((1<<24-1)<<(127-23) + 1<<(127-52)),
			dst: func() reflect.Value {
				var v float32

				return reflect.ValueOf(&v).Elem()
			}(),
			err: ErrSetOverflow,
		},
		"InvalidSyntax": {
			to: "foo",
			dst: func() reflect.Value {
				var v float64

				return reflect.ValueOf(&v).Elem()
			}(),
			err: ErrSetInvalidValue,
		},
		"String": {
			to: "4.99",
			dst: func() reflect.Value {
				var v float64

				return reflect.ValueOf(&v).Elem()
			}(),
			want: 4.99,
		},
		"Float32": {
			to: float32(32),
			dst: func() reflect.Value {
				var v float64

				return reflect.ValueOf(&v).Elem()
			}(),
			want: 32,
		},
		"Float64": {
			to: float64(64),
			dst: func() reflect.Value {
				var v float64

				return reflect.ValueOf(&v).Elem()
			}(),
			want: 64,
		},
	}

	for name, testCase := range cases {
		tc := testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := setValueToFloat(tc.dst, tc.to)

			if !errors.Is(err, tc.err) {
				t.Errorf("want %+v err, got %+v", err, tc.err)
			}

			if !reflect.DeepEqual(tc.want, tc.dst.Float()) {
				t.Errorf("want %+v float64, got %+v", tc.want, tc.dst.Float())
			}
		})
	}
}

func Test_setValueToSlice(t *testing.T) {
	cases := map[string]struct {
		to   interface{}
		dst  reflect.Value
		want []string
		err  error
	}{
		"NotSetable": {
			to: "foo",
			dst: func() reflect.Value {
				var v []string

				return reflect.ValueOf(v)
			}(),
			err: ErrReflectValueNotSetable,
		},
		"InvalidType": {
			to: "foo",
			dst: func() reflect.Value {
				var v []string

				return reflect.ValueOf(&v).Elem()
			}(),
			err: ErrSetInvalidType,
		},
		"InvalidValue": {
			to: []interface{}{struct{}{}},
			dst: func() reflect.Value {
				var v []string

				return reflect.ValueOf(&v).Elem()
			}(),
			err: ErrSetInvalidType,
		},
		"String": {
			to: []string{"foo", "bar"},
			dst: func() reflect.Value {
				var v []string

				return reflect.ValueOf(&v).Elem()
			}(),
			want: []string{"foo", "bar"},
		},
	}

	for name, testCase := range cases {
		tc := testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := setValueToSlice(tc.dst, tc.to)

			if !errors.Is(err, tc.err) {
				t.Errorf("want %+v err, got %+v", tc.err, err)
			}

			if !reflect.DeepEqual(tc.want, tc.dst.Interface()) {
				t.Errorf("want %+v value, got %+v", tc.want, tc.dst.Interface())
			}
		})
	}
}
