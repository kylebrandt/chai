package query

import (
	"errors"

	"github.com/asdine/genji"
	"github.com/asdine/genji/record"
	"github.com/asdine/genji/table"
)

// A Query can execute actions against the database. It can read or write data
// from any table, or even alter the structure of the database.
// Results are returned as a stream.
type Query interface {
	Run(*genji.DB) Result
}

// A Statement represents a unique action that can be executed against the database.
type Statement interface {
	Run(*TxManager) (*table.Stream, error)
}

// TxManager is used by statements to automatically open transactions.
// If the Tx field is nil, it will automatically create a new transaction.
// If the Tx field is not nil, it will be passed to View and Update.
type TxManager struct {
	DB *genji.DB
	Tx *genji.Tx
}

// View runs fn in a read-only transaction if the Tx field is nil.
// If not, it will pass it to fn regardless of it being a read-only or read-write transaction.
func (tx TxManager) View(fn func(tx *genji.Tx) error) error {
	if tx.Tx != nil {
		return fn(tx.Tx)
	}

	return tx.DB.View(fn)
}

// Update runs fn in a read-write transaction if the Tx field is nil.
// If not, it will pass it to fn regardless of it being a read-only or read-write transaction.
func (tx TxManager) Update(fn func(tx *genji.Tx) error) error {
	if tx.Tx != nil {
		return fn(tx.Tx)
	}

	return tx.DB.Update(fn)
}

// Result of a query.
type Result struct {
	*table.Stream
	err error
}

// Err returns a non nil error if an error occured during the query.
func (r Result) Err() error {
	return r.err
}

// Scan takes a table scanner and passes it the result table.
func (r Result) Scan(s table.Scanner) error {
	if r.err != nil {
		return r.err
	}

	return s.ScanTable(r.Stream)
}

var errStop = errors.New("stop")

func whereClause(tx *genji.Tx, e Expr) func(recordID []byte, r record.Record) (bool, error) {
	if e == nil {
		return func(recordID []byte, r record.Record) (bool, error) {
			return true, nil
		}
	}

	return func(recordID []byte, r record.Record) (bool, error) {
		sc, err := e.Eval(EvalContext{Tx: tx, Record: r})
		if err != nil {
			return false, err
		}

		return sc.Truthy(), nil
	}
}
