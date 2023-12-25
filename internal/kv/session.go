package kv

import (
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/pebble"
)

// Common errors returned by the engine.
var (
	// ErrKeyNotFound is returned when the targeted key doesn't exist.
	ErrKeyNotFound = errors.New("key not found")

	// ErrKeyAlreadyExists is returned when the targeted key already exists.
	ErrKeyAlreadyExists = errors.New("key already exists")
)

type Session interface {
	Commit() error
	Close() error
	// Insert inserts a key-value pair. If it already exists, it returns ErrKeyAlreadyExists.
	Insert(k, v []byte) error
	// Put stores a key-value pair. If it already exists, it overrides it.
	Put(k, v []byte) error
	// Get returns a value associated with the given key. If not found, returns ErrKeyNotFound.
	Get(k []byte) ([]byte, error)
	// Exists returns whether a key exists and is visible by the current session.
	Exists(k []byte) (bool, error)
	// Delete a record by key. If not found, returns ErrKeyNotFound.
	Delete(k []byte) error
	DeleteRange(start []byte, end []byte) error
	Iterator(opts *IterOptions) (Iterator, error)
}

type Iterator interface {
	Close() error
	First() bool
	Last() bool
	Valid() bool
	Next() bool
	Prev() bool
	Error() error
	Key() []byte
	Value() ([]byte, error)
}

type IterOptions struct {
	// LowerBound specifies the smallest key (inclusive) that the iterator will
	// return during iteration. If the iterator is seeked or iterated past this
	// boundary the iterator will return Valid()==false. Setting LowerBound
	// effectively truncates the key space visible to the iterator.
	LowerBound []byte
	// UpperBound specifies the largest key (exclusive) that the iterator will
	// return during iteration. If the iterator is seeked or iterated past this
	// boundary the iterator will return Valid()==false. Setting UpperBound
	// effectively truncates the key space visible to the iterator.
	UpperBound []byte
}

type iterator struct {
	*pebble.Iterator
}

func (i *iterator) Value() ([]byte, error) {
	return i.Iterator.ValueAndErr()
}

// Get returns a value associated with the given key. If not found, returns ErrKeyNotFound.
func get(r pebble.Reader, k []byte) ([]byte, error) {
	value, closer, err := r.Get(k)
	if err != nil {
		if errors.Is(err, pebble.ErrNotFound) {
			return nil, errors.WithStack(ErrKeyNotFound)
		}

		return nil, err
	}

	cp := make([]byte, len(value))
	copy(cp, value)

	err = closer.Close()
	if err != nil {
		return nil, err
	}

	return cp, nil
}

// Exists returns whether a key exists and is visible by the current session.
func exists(r pebble.Reader, k []byte) (bool, error) {
	_, closer, err := r.Get(k)
	if err != nil {
		if errors.Is(err, pebble.ErrNotFound) {
			return false, nil
		}

		return false, err
	}
	err = closer.Close()
	if err != nil {
		return false, err
	}
	return true, nil
}
