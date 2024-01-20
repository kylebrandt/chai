package types

import (
	"github.com/cockroachdb/errors"
)

var (
	// ErrFieldNotFound must be returned by object implementations, when calling the GetByField method and
	// the field wasn't found in the object.
	ErrFieldNotFound = errors.New("field not found")
	// ErrValueNotFound must be returned by Array implementations, when calling the GetByIndex method and
	// the index wasn't found in the array.
	ErrValueNotFound = errors.New("value not found")

	errStop = errors.New("stop")
)

// ValueType represents a value type supported by the database.
type ValueType uint8

// List of supported value types.
const (
	// TypeAny denotes the absence of type
	TypeAny ValueType = iota
	TypeNull
	TypeBoolean
	TypeInteger
	TypeDouble
	TypeTimestamp
	TypeText
	TypeBlob
	TypeArray
	TypeObject
)

func (t ValueType) String() string {
	switch t {
	case TypeNull:
		return "null"
	case TypeBoolean:
		return "boolean"
	case TypeInteger:
		return "integer"
	case TypeDouble:
		return "double"
	case TypeTimestamp:
		return "timestamp"
	case TypeBlob:
		return "blob"
	case TypeText:
		return "text"
	case TypeArray:
		return "array"
	case TypeObject:
		return "object"
	}

	return "any"
}

// IsNumber returns true if t is either an integer or a float.
func (t ValueType) IsNumber() bool {
	return t == TypeInteger || t == TypeDouble
}

// IsTimestampCompatible returns true if t is either a timestamp, an integer, or a text.
func (t ValueType) IsTimestampCompatible() bool {
	return t == TypeTimestamp || t == TypeText
}

func (t ValueType) IsComparableWith(other ValueType) bool {
	if t == other {
		return true
	}

	if t.IsNumber() && other.IsNumber() {
		return true
	}

	if t.IsTimestampCompatible() && other.IsTimestampCompatible() {
		return true
	}

	return false
}

// IsAny returns whether this is type is Any or a real type
func (t ValueType) IsAny() bool {
	return t == TypeAny
}

type Value interface {
	Type() ValueType
	V() any
	IsZero() (bool, error)
	String() string
	MarshalJSON() ([]byte, error)
	MarshalText() ([]byte, error)
	EQ(other Value) (bool, error)
	GT(other Value) (bool, error)
	GTE(other Value) (bool, error)
	LT(other Value) (bool, error)
	LTE(other Value) (bool, error)
	Between(a, b Value) (bool, error)
	// Add u to v and return the result.
	// Only numeric values and booleans can be added together.
	Add(other Value) (Value, error)
	// Sub calculates v - u and returns the result.
	// Only numeric values and booleans can be calculated together.
	Sub(other Value) (Value, error)
	// Mul calculates v * u and returns the result.
	// Only numeric values and booleans can be calculated together.
	Mul(other Value) (Value, error)
	// Div calculates v / u and returns the result.
	// Only numeric values and booleans can be calculated together.
	// If both v and u are integers, the result will be an integer.
	Div(other Value) (Value, error)
	// Mod calculates v / u and returns the result.
	// Only numeric values and booleans can be calculated together.
	// If both v and u are integers, the result will be an integer.
	Mod(other Value) (Value, error)
	// BitwiseAnd calculates v & u and returns the result.
	// Only numeric values and booleans can be calculated together.
	// If both v and u are integers, the result will be an integer.
	BitwiseAnd(other Value) (Value, error)
	// BitwiseOr calculates v | u and returns the result.
	// Only numeric values and booleans can be calculated together.
	// If both v and u are integers, the result will be an integer.
	BitwiseOr(other Value) (Value, error)
	// BitwiseXor calculates v ^ u and returns the result.
	// Only numeric values and booleans can be calculated together.
	// If both v and u are integers, the result will be an integer.
	BitwiseXor(other Value) (Value, error)
}

// A Object represents a group of key value pairs.
type Object interface {
	// Iterate goes through all the fields of the object and calls the given function by passing each one of them.
	// If the given function returns an error, the iteration stops.
	Iterate(fn func(field string, value Value) error) error
	// GetByField returns a value by field name.
	// Must return ErrFieldNotFound if the field doesn't exist.
	GetByField(field string) (Value, error)

	// MarshalJSON implements the json.Marshaler interface.
	// It returns a JSON representation of the object.
	MarshalJSON() ([]byte, error)
}

// An Array contains a set of values.
type Array interface {
	// Iterate goes through all the values of the array and calls the given function by passing each one of them.
	// If the given function returns an error, the iteration stops.
	Iterate(fn func(i int, value Value) error) error
	// GetByIndex returns a value by index of the array.
	GetByIndex(i int) (Value, error)

	// MarshalJSON implements the json.Marshaler interface.
	// It returns a JSON representation of the array.
	MarshalJSON() ([]byte, error)
}
