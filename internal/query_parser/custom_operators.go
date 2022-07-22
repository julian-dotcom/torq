package query_parser

import (
	"fmt"
	sq "github.com/Masterminds/squirrel"
)

// Overlap is used to check if two arrays overlap (SQL operator &&). But can also be used as the ANY
// operator in SQL to check if a string/number is in an array.
func Overlap(param interface{}, key string, notEq bool) (r sq.Sqlizer, err error) {

	var equalOpr = "="

	if notEq {
		equalOpr = "!="
	}

	switch param.(type) {
	case nil:
		return sq.Eq{key: nil}, nil
	case []float64:
		sa := sq.Or{}
		for _, value := range param.([]float64) {
			sa = append(sa, sq.Expr(fmt.Sprintf("? %s ANY(%d)", equalOpr, key), value))
		}
		return sa, nil
	case []string:
		sa := sq.Or{}
		for _, value := range param.([]string) {
			sa = append(sa, sq.Expr(fmt.Sprintf("? %s ANY(%s)", equalOpr, key), value))
		}
		return sa, nil
	case float64:
		return sq.Expr(fmt.Sprintf("? %s ANY(%s)", equalOpr, key), param), nil
	case string:
		return sq.Expr(fmt.Sprintf("? %s ANY(%s)", equalOpr, key), param), nil
	default:
		return r, fmt.Errorf("unsupported parameter type: %T. Any only supports string, "+
			"float or an array of string or float", param)
	}
}
