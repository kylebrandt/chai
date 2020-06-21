package expr_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/genjidb/genji/database"
	"github.com/genjidb/genji/document"
	"github.com/genjidb/genji/sql/parser"
	"github.com/genjidb/genji/sql/query/expr"
	"github.com/stretchr/testify/require"
)

var doc document.Document = func() document.Document {
	d, _ := document.NewFromJSON([]byte(`{
		"a": 1,
		"b": {"foo bar": [1, 2]},
		"c": [1, {"foo": "bar"}, [1, 2]]
	}`))

	return d
}()

var stackWithDoc = expr.EvalStack{
	Document: doc,
}

var fakeTableConfig = &database.TableConfig{
	FieldConstraints: []database.FieldConstraint{
		{Path: []string{"c", "0"}, IsPrimaryKey: true},
		{Path: []string{"c", "1"}},
	},
}
var stackWithDocAndConfig = expr.EvalStack{
	Document: doc,
	Cfg:      fakeTableConfig,
}

var nullLitteral = document.NewNullValue()

func testExpr(t testing.TB, exprStr string, stack expr.EvalStack, want document.Value, fails bool) {
	e, _, err := parser.NewParser(strings.NewReader(exprStr)).ParseExpr()
	require.NoError(t, err)
	res, err := e.Eval(stack)
	if fails {
		require.Error(t, err)
	} else {
		require.NoError(t, err)
		require.Equal(t, want, res)
	}
}

func TestString(t *testing.T) {
	var operands = []string{
		`10.4`,
		"true",
		"500",
		`foo.bar.1`,
		`"hello"`,
		`[1, 2, "foo"]`,
		`{"a": "foo", "b": 10}`,
		"pk()",
		"CAST(10 AS int64)",
	}

	var operators = []string{
		"=", ">", ">=", "<", "<=",
		"+", "-", "*", "/", "%", "&", "|", "^",
		"AND", "OR",
	}

	testFn := func(s string, want string) {
		e, _, err := parser.NewParser(strings.NewReader(s)).ParseExpr()
		require.NoError(t, err)
		require.Equal(t, want, fmt.Sprintf("%v", e))
	}

	for _, op := range operands {
		testFn(op, op)
	}

	for _, op := range operators {
		want := fmt.Sprintf("10.4 %s foo.bar.1", op)
		testFn(want, want)
	}
}