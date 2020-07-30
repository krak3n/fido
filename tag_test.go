package fido

import (
	"errors"
	"reflect"
	"testing"
)

func TestLookupTag(t *testing.T) {
	cases := map[string]struct {
		tag   string
		field reflect.StructField
		want  Tag
		err   error
	}{
		"NoTagOnSructField": {
			tag: DefaultStructTag,
			field: reflect.StructField{
				Name: "Foo",
			},
			want: Tag{
				FieldName: "Foo",
			},
			err: ErrStructTagNotFound,
		},
		"ExtractsTagValues": {
			tag: DefaultStructTag,
			field: reflect.StructField{
				Name: "Foo",
				Tag:  reflect.StructTag(`fido:"foo"`),
			},
			want: Tag{
				FieldName: "Foo",
				RawTag:    `foo`,
				Name:      "foo",
			},
		},
	}

	for name, testCase := range cases {
		tc := testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tag, err := LookupTag(tc.tag, tc.field)

			if !errors.Is(err, tc.err) {
				t.Errorf("want %+v err, got %+v", tc.err, err)
			}

			if !reflect.DeepEqual(tc.want, tag) {
				t.Errorf("want %+v tag, got %+v", tc.want, tag)
			}
		})
	}
}
