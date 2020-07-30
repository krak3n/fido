package fido

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
	ErrSetOverflow
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
	case ErrSetOverflow:
		return "set overflow"
	}

	return "unknown error"
}
