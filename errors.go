package fido

import "fmt"

const (
	ErrDestinationTypeInvalid Error = iota + 1
	ErrDestinationNil
	ErrDestinationNotPtr
	ErrStructTagNotFound
	ErrInvalidType
	ErrInvalidPath
	ErrFieldNotFound
	ErrExpectedMap
	ErrInvalidMapKeyType
	ErrInvalidReflectValue
	ErrReflectValueNotAddressable
	ErrReflectValueNotSetable
	ErrSetInvalidType
	ErrSetInvalidValue
	ErrSetOverflow
	ErrDoesNotImplementNotifyProvider
)

type Error uint8

func (e Error) Error() string {
	switch e {
	case ErrDestinationTypeInvalid:
		return "invalid destination type"
	case ErrDestinationNil:
		return "destination is nil"
	case ErrDestinationNotPtr:
		return "destination is not a pointer"
	case ErrStructTagNotFound:
		return "struct tag not found on field"
	case ErrInvalidPath:
		return "invalid path"
	case ErrFieldNotFound:
		return "field not found"
	case ErrExpectedMap:
		return "expected map"
	case ErrInvalidMapKeyType:
		return "invalid map key type"
	case ErrInvalidReflectValue:
		return "invalid reflect value"
	case ErrReflectValueNotAddressable:
		return "reflect value is not addressable"
	case ErrReflectValueNotSetable:
		return "reflect value cannot be set"
	case ErrSetInvalidType:
		return "cannot set to type"
	case ErrSetInvalidValue:
		return "cannot set to value"
	case ErrSetOverflow:
		return "set overflow"
	case ErrDoesNotImplementNotifyProvider:
		return "does not implement NotifyProvider extension interface"
	}

	return "unknown error"
}

// NonErrPanic is returned if Fido recovers from a panic that was not an error.
type NonErrPanic struct {
	Value interface{}
}

func (e NonErrPanic) Error() string {
	return fmt.Sprintf("%v: non error panic", e.Value)
}
