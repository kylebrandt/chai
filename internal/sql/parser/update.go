package parser

import (
	"github.com/genjidb/genji/internal/query"
	"github.com/genjidb/genji/internal/sql/scanner"
)

// parseUpdateStatement parses a update string and returns a Statement AST object.
// This function assumes the UPDATE token has already been consumed.
func (p *Parser) parseUpdateStatement() (*query.UpdateStmt, error) {
	var stmt query.UpdateStmt
	var err error

	// Parse table name
	stmt.TableName, err = p.parseIdent()
	if err != nil {
		pErr := err.(*ParseError)
		pErr.Expected = []string{"table_name"}
		return nil, pErr
	}

	// Parse clause: SET or UNSET.
	tok, pos, lit := p.ScanIgnoreWhitespace()
	switch tok {
	case scanner.SET:
		stmt.SetPairs, err = p.parseSetClause()
	case scanner.UNSET:
		stmt.UnsetFields, err = p.parseUnsetClause()
	default:
		err = newParseError(scanner.Tokstr(tok, lit), []string{"SET", "UNSET"}, pos)
	}
	if err != nil {
		return nil, err
	}

	// Parse condition: "WHERE EXPR".
	stmt.WhereExpr, err = p.parseCondition()
	if err != nil {
		return nil, err
	}

	return &stmt, nil
}

// parseSetClause parses the "SET" clause of the query.
func (p *Parser) parseSetClause() ([]query.UpdateSetPair, error) {
	var pairs []query.UpdateSetPair

	firstPair := true
	for {
		if !firstPair {
			// Scan for a comma.
			tok, _, _ := p.ScanIgnoreWhitespace()
			if tok != scanner.COMMA {
				p.Unscan()
				break
			}
		}

		// Scan the identifier for the path name.
		path, err := p.parsePath()
		if err != nil {
			pErr := err.(*ParseError)
			pErr.Expected = []string{"path"}
			return nil, pErr
		}

		// Scan the eq sign
		if err := p.parseTokens(scanner.EQ); err != nil {
			return nil, err
		}

		// Scan the expr for the value.
		expr, err := p.ParseExpr()
		if err != nil {
			return nil, err
		}
		pairs = append(pairs, query.UpdateSetPair{Path: path, E: expr})

		firstPair = false
	}

	return pairs, nil
}

func (p *Parser) parseUnsetClause() ([]string, error) {
	var fields []string

	firstField := true
	for {
		if !firstField {
			// Scan for a comma.
			tok, _, _ := p.ScanIgnoreWhitespace()
			if tok != scanner.COMMA {
				p.Unscan()
				break
			}
		}

		// Scan the identifier for the path to unset.
		lit, err := p.parseIdent()
		if err != nil {
			return nil, err
		}
		fields = append(fields, lit)

		firstField = false
	}
	return fields, nil
}