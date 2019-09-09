package query

import (
	"errors"

	"github.com/asdine/genji"
	"github.com/asdine/genji/record"
	"github.com/asdine/genji/table"
)

// UpdateStmt is a DSL that allows creating a full Update query.
// It is typically created using the Update function.
type UpdateStmt struct {
	tableSelector TableSelector
	pairs         map[string]Expr
	whereExpr     Expr
}

// Update creates a DSL equivalent to the SQL Update command.
func Update(tableSelector TableSelector) UpdateStmt {
	return UpdateStmt{
		tableSelector: tableSelector,
		pairs:         make(map[string]Expr),
	}
}

// Set assignes the result of the evaluation of e into the field selected
// by f.
func (u UpdateStmt) Set(fieldName string, e Expr) UpdateStmt {
	u.pairs[fieldName] = e
	return u
}

// Where uses e to filter records if it evaluates to a falsy value.
// Calling this method is optional.
func (u UpdateStmt) Where(e Expr) UpdateStmt {
	u.whereExpr = e
	return u
}

// Run the Update query within tx.
// If Where was called, records will be filtered depending on the result of the
// given expression. If the Where expression implements the IndexMatcher interface,
// the MatchIndex method will be called instead of the Eval one.
func (u UpdateStmt) Run(tx *genji.Tx) error {
	if u.tableSelector == nil {
		return errors.New("missing table selector")
	}

	if len(u.pairs) == 0 {
		return errors.New("Set method not called")
	}

	t, err := u.tableSelector.SelectTable(tx)
	if err != nil {
		return err
	}

	var tr table.Reader = t

	var useIndex bool

	if im, ok := u.whereExpr.(IndexMatcher); ok {
		tree, ok, err := im.MatchIndex(t)
		if err != nil && err != genji.ErrIndexNotFound {
			return err
		}

		if ok && err == nil {
			useIndex = true
			tr = &indexResultTable{
				tree:  tree,
				table: t,
			}
		}
	}

	st := table.NewStream(tr)

	if !useIndex {
		st = st.Filter(whereClause(tx, u.whereExpr))
	}

	return st.Iterate(func(recordID []byte, r record.Record) error {
		var fb record.FieldBuffer
		err := fb.ScanRecord(r)
		if err != nil {
			return err
		}

		for fname, e := range u.pairs {
			f, err := fb.GetField(fname)
			if err != nil {
				return err
			}

			s, err := e.Eval(EvalContext{
				Tx:     tx,
				Record: r,
			})
			if err != nil {
				return err
			}

			f.Type = s.Type
			f.Data = s.Data
			err = fb.Replace(f.Name, f)
			if err != nil {
				return err
			}

			err = t.Replace(recordID, &fb)
			if err != nil {
				return err
			}
		}

		return nil
	})
}
