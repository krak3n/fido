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
		want string
		err  error
	}{
		"InvalidType": {
			to:  []string{"foo"},
			err: ErrSetInvalidType,
		},
		"Stringer": {
			to:   TestStringer{"foo"},
			want: "foo",
		},
		"String": {
			to:   "foo",
			want: "foo",
		},
		"Bool": {
			to:   true,
			want: "true",
		},
		"Int": {
			to:   int(1),
			want: "1",
		},
		"Int8": {
			to:   int8(8),
			want: "8",
		},
		"Int16": {
			to:   int16(16),
			want: "16",
		},
		"Int32": {
			to:   int32(32),
			want: "32",
		},
		"Int64": {
			to:   int64(64),
			want: "64",
		},
		"Uint": {
			to:   uint(1),
			want: "1",
		},
		"Uint8": {
			to:   uint8(8),
			want: "8",
		},
		"Uint16": {
			to:   uint16(16),
			want: "16",
		},
		"Uint32": {
			to:   uint32(32),
			want: "32",
		},
		"Uint64": {
			to:   uint64(64),
			want: "64",
		},
	}

	for name, testCase := range cases {
		tc := testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var str string

			err := setValueToString(reflect.ValueOf(&str).Elem(), tc.to)

			if !errors.Is(err, tc.err) {
				t.Errorf("want %+v err, got %+v", err, tc.err)
			}

			if !reflect.DeepEqual(tc.want, str) {
				t.Errorf("want %+v value, got %+v", tc.want, str)
			}
		})
	}
}

func Test_setValueToInt(t *testing.T) {
	cases := map[string]struct {
		to   interface{}
		want int64
		err  error
	}{
		"InvalidType": {
			to:  []string{"foo"},
			err: ErrSetInvalidType,
		},
		"String": {
			to:   "1",
			want: 1,
		},
		"Int": {
			to:   int(1),
			want: 1,
		},
		"Int8": {
			to:   int8(8),
			want: 8,
		},
		"Int16": {
			to:   int16(16),
			want: 16,
		},
		"Int32": {
			to:   int32(32),
			want: 32,
		},
		"Int64": {
			to:   int64(64),
			want: 64,
		},
	}

	for name, testCase := range cases {
		tc := testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var i int64

			err := setValueToInt(reflect.ValueOf(&i).Elem(), tc.to)

			if !errors.Is(err, tc.err) {
				t.Errorf("want %+v err, got %+v", err, tc.err)
			}

			if !reflect.DeepEqual(tc.want, i) {
				t.Errorf("want %+v value, got %+v", tc.want, i)
			}
		})
	}
}

func Test_setValueToUint(t *testing.T) {
	cases := map[string]struct {
		to   interface{}
		want uint64
		err  error
	}{
		"InvalidType": {
			to:  []string{"foo"},
			err: ErrSetInvalidType,
		},
		"String": {
			to:   "1",
			want: 1,
		},
		"Uint": {
			to:   uint(1),
			want: 1,
		},
		"Uint8": {
			to:   uint8(8),
			want: 8,
		},
		"Uint16": {
			to:   uint16(16),
			want: 16,
		},
		"Uint32": {
			to:   uint32(32),
			want: 32,
		},
		"Uint64": {
			to:   uint64(64),
			want: 64,
		},
	}

	for name, testCase := range cases {
		tc := testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var ui uint64

			err := setValueToUint(reflect.ValueOf(&ui).Elem(), tc.to)

			if !errors.Is(err, tc.err) {
				t.Errorf("want %+v err, got %+v", err, tc.err)
			}

			if !reflect.DeepEqual(tc.want, ui) {
				t.Errorf("want %+v value, got %+v", tc.want, ui)
			}
		})
	}
}

func Test_setValueToFloat(t *testing.T) {
	cases := map[string]struct {
		to   interface{}
		want float64
		err  error
	}{
		"InvalidType": {
			to:  []string{"foo"},
			err: ErrSetInvalidType,
		},
		"String": {
			to:   "4.99",
			want: 4.99,
		},
		"Float32": {
			to:   float32(32),
			want: 32,
		},
		"Float64": {
			to:   float64(64),
			want: 64,
		},
	}

	for name, testCase := range cases {
		tc := testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var fl float64

			err := setValueToFloat(reflect.ValueOf(&fl).Elem(), tc.to)

			if !errors.Is(err, tc.err) {
				t.Errorf("want %+v err, got %+v", err, tc.err)
			}

			if !reflect.DeepEqual(tc.want, fl) {
				t.Errorf("want %+v value, got %+v", tc.want, fl)
			}
		})
	}
}

func Test_setValueToSlice(t *testing.T) {
	cases := map[string]struct {
		to   interface{}
		want []string
		err  error
	}{
		"InvalidType": {
			to:  "foo",
			err: ErrSetInvalidType,
		},
		"String": {
			to:   []string{"foo", "bar"},
			want: []string{"foo", "bar"},
		},
	}

	for name, testCase := range cases {
		tc := testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var slice []string

			err := setValueToSlice(reflect.ValueOf(&slice).Elem(), tc.to)

			if !errors.Is(err, tc.err) {
				t.Errorf("want %+v err, got %+v", tc.err, err)
			}

			if !reflect.DeepEqual(tc.want, slice) {
				t.Errorf("want %+v value, got %+v", tc.want, slice)
			}
		})
	}
}
