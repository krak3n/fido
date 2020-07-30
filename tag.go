package fido

import (
	"fmt"
	"reflect"
	"strings"
)

type Tag struct {
	RawTag    string
	Name      string
	FieldName string
}

func LookupTag(tag string, f reflect.StructField) (Tag, error) {
	t := Tag{
		FieldName: f.Name,
	}

	v, ok := f.Tag.Lookup(tag)
	if !ok {
		return t, fmt.Errorf("%w: %s", ErrStructTagNotFound, tag)
	}

	t.RawTag = v

	for i, v := range strings.Split(v, ",") {
		switch i {
		case 0:
			t.Name = v
		}
	}

	return t, nil
}
