package fido

import (
	"fmt"
	"reflect"
	"strings"
)

// Tag holds struct field specific configuration from a struct tag.
type Tag struct {
	RawTag    string
	Name      string
	FieldName string
}

func (t Tag) String() string {
	return fmt.Sprintf("%s: %s", t.FieldName, t.RawTag)
}

// LookupTag looks for the given struct tag on a struct field returning the decoded tag.
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
