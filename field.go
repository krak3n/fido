package fido

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

const pathSeparator = "."

// A Field can be set to a value by a provider.
type Field interface {
	fmt.Stringer

	Path() Path
	Value() reflect.Value
	Provider() Provider
	Set(interface{}, Provider) error
}

// Path is a path to a field, e.g [foo.bar.baz] = fizz.
type Path []string

func (p Path) String() string {
	return p.key()
}

func (p Path) key() string {
	return strings.Join(p, pathSeparator)
}

func (p Path) equal(other Path) bool {
	if len(p) != len(other) {
		return false
	}

	for i, v := range p {
		if v != other[i] {
			return false
		}
	}

	return true
}

type fields map[string]Field

func (f fields) set(path Path, field Field) {
	f[path.key()] = field
}

// get finds the closest field match for the given path.
func (f fields) get(path Path) (Field, bool) {
	field, ok := f[path.key()]
	if ok {
		return field, ok
	}

	if len(path) == 0 {
		return nil, false
	}

	path = path[:(len(path) - 1)]
	if len(path) == 0 {
		return nil, false
	}

	return f.get(path)
}

type field struct {
	path     Path
	value    reflect.Value
	provider Provider
}

func (f *field) String() string {
	return f.path.key()
}

func (f *field) Path() Path {
	return f.path
}

func (f *field) Value() reflect.Value {
	return f.value
}

func (f *field) Provider() Provider {
	return f.provider
}

func (f *field) Set(to interface{}, p Provider) error {
	if err := setValue(f.value, to); err != nil {
		return fmt.Errorf("%w: cannot set %s to %+v", err, f, to)
	}

	f.provider = p

	return nil
}

type mapfield struct {
	*field

	dst reflect.Value // destination map
	idx reflect.Value // destination map index
}

func (f *mapfield) Set(to interface{}, by Provider) error {
	if err := f.field.Set(to, by); err != nil {
		return err
	}

	f.dst.SetMapIndex(f.idx, f.value)

	return nil
}

func setValue(dst reflect.Value, to interface{}) error {
	switch dst.Kind() {
	case reflect.Ptr:
		return setValue(dst.Elem(), to)
	case reflect.String:
		return setValueToString(dst, to)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return setValueToInt(dst, to)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return setValueToUint(dst, to)
	case reflect.Float32, reflect.Float64:
		return setValueToFloat(dst, to)
	case reflect.Slice, reflect.Array:
		return setValueToSlice(dst, to)
	default:
		return fmt.Errorf("%w: could not set %s to %s", ErrSetInvalidType, to, dst.Kind())
	}
}

func setValueToString(dst reflect.Value, to interface{}) error {
	if !dst.CanSet() {
		return ErrReflectValueNotSetable
	}

	var str string

	switch to := to.(type) {
	case string:
		str = to
	case bool:
		str = strconv.FormatBool(to)
	case int:
		str = strconv.FormatInt(int64(to), 10)
	case int8:
		str = strconv.FormatInt(int64(to), 10)
	case int16:
		str = strconv.FormatInt(int64(to), 10)
	case int32:
		str = strconv.FormatInt(int64(to), 10)
	case int64:
		str = strconv.FormatInt(to, 10)
	case uint:
		str = strconv.FormatUint(uint64(to), 10)
	case uint8:
		str = strconv.FormatUint(uint64(to), 10)
	case uint16:
		str = strconv.FormatUint(uint64(to), 10)
	case uint32:
		str = strconv.FormatUint(uint64(to), 10)
	case uint64:
		str = strconv.FormatUint(to, 10)
	default:
		stringer, ok := to.(fmt.Stringer)
		if !ok {
			return fmt.Errorf("%w: cannot set %T to %s", ErrSetInvalidType, to, dst.Kind())
		}

		str = stringer.String()
	}

	dst.SetString(str)

	return nil
}

func setValueToInt(dst reflect.Value, to interface{}) error {
	if !dst.CanSet() {
		return ErrReflectValueNotSetable
	}

	var i int64

	switch t := to.(type) {
	case string:
		v, err := strconv.ParseInt(t, 10, 64)
		if err != nil {
			return fmt.Errorf("%w: could not convert %s to int64", ErrSetInvalidValue, t)
		}

		i = v
	case int:
		i = int64(t)
	case int8:
		i = int64(t)
	case int16:
		i = int64(t)
	case int32:
		i = int64(t)
	case int64:
		i = t
	default:
		return fmt.Errorf("%w: cannot set %T to %s", ErrSetInvalidType, to, dst.Kind())
	}

	if dst.OverflowInt(i) {
		return fmt.Errorf("%w: %T to %s", ErrSetOverflow, to, dst.Kind())
	}

	dst.SetInt(i)

	return nil
}

func setValueToUint(dst reflect.Value, to interface{}) error {
	if !dst.CanSet() {
		return ErrReflectValueNotSetable
	}

	var i uint64

	switch t := to.(type) {
	case string:
		v, err := strconv.ParseUint(t, 10, 64)
		if err != nil {
			return fmt.Errorf("%w: could not convert %s to uint64", ErrSetInvalidValue, t)
		}

		i = v
	case uint:
		i = uint64(t)
	case uint8:
		i = uint64(t)
	case uint16:
		i = uint64(t)
	case uint32:
		i = uint64(t)
	case uint64:
		i = t
	default:
		return fmt.Errorf("%w: cannot set %T to %s", ErrSetInvalidType, to, dst.Kind())
	}

	if dst.OverflowUint(i) {
		return fmt.Errorf("%w: %T to %s", ErrSetOverflow, to, dst.Kind())
	}

	dst.SetUint(i)

	return nil
}

func setValueToFloat(dst reflect.Value, to interface{}) error {
	if !dst.CanSet() {
		return ErrReflectValueNotSetable
	}

	var fl float64

	switch t := to.(type) {
	case string:
		v, err := strconv.ParseFloat(t, 64)
		if err != nil {
			return fmt.Errorf("%w: could not convert %s to float64", ErrSetInvalidValue, t)
		}

		fl = v
	case float32:
		fl = float64(t)
	case float64:
		fl = t
	default:
		return fmt.Errorf("%w: cannot set %T to %s", ErrSetInvalidType, to, dst.Kind())
	}

	if dst.OverflowFloat(fl) {
		return fmt.Errorf("%w: %T to %s", ErrSetOverflow, to, dst.Kind())
	}

	dst.SetFloat(fl)

	return nil
}

func setValueToSlice(dst reflect.Value, to interface{}) error {
	if !dst.CanSet() {
		return ErrReflectValueNotSetable
	}

	dt := dst.Type()
	tv := reflect.ValueOf(to)

	if tv.Kind() != reflect.Array && tv.Kind() != reflect.Slice {
		return fmt.Errorf("%w: expected array or slice, got %T", ErrSetInvalidType, to)
	}

	slice := reflect.MakeSlice(reflect.SliceOf(dt.Elem()), tv.Len(), tv.Cap())

	for i := 0; i < tv.Len(); i++ {
		e := reflect.New(dt.Elem())
		if err := setValue(e, tv.Index(i).Interface()); err != nil {
			return err
		}

		slice.Index(i).Set(e.Elem())
	}

	dst.Set(slice)

	return nil
}
