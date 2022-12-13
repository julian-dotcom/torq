package query_parser

import (
	"github.com/cockroachdb/errors"
	"reflect"
)

type QueryParser struct {
	AllowedColumns []string
}

func GetDBKeyName(v interface{}) (string, error) {
	field, ok := reflect.TypeOf(v).Elem().FieldByName("Key")
	if !ok {
		return "", errors.New("field not found")
	}
	return string(field.Tag.Get("db")), nil
}

func NewParser(allowedColumns []string) *QueryParser {
	return &QueryParser{
		AllowedColumns: allowedColumns,
	}
}

func (qp *QueryParser) IsAllowed(c string) bool {
	for _, ac := range qp.AllowedColumns {
		if ac == c {
			return true
		}
	}
	return false
}
