package query_parser

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Example of sort json input
// "sortBy":[{"value":"revenue_out", direction":"desc"}]

type Sort struct {
	Value     string `json:"value"`
	Direction string `json:"direction"`
}

func ParseSortParams(params string, allowedColumns []string) ([]string, error) {
	var sort []Sort
	err := json.Unmarshal([]byte(params), &sort)
	if err != nil {
		return nil, err
	}

	// Whitelist the columns that are allowed to sorted by.
	qp := QueryParser{
		AllowedColumns: allowedColumns,
	}

	sortString, err := qp.ParseSortClauses(sort)
	if err != nil {
		return nil, err
	}
	return sortString, nil
}

func (qp *QueryParser) ParseSort(s Sort) (r string, err error) {

	// Prevents SQL injection by only allowing whitelisted column names.
	if !qp.IsAllowed(s.Value) {
		return r,
			fmt.Errorf("sorting by %s is not allwed. Try one of: %v",
				s.Value,
				strings.Join(qp.AllowedColumns, ", "),
			)
	}

	// Prevent SQL injection by only allowing asc and desc as directions.
	if !(s.Direction == "asc" || s.Direction == "desc") {
		return r, fmt.Errorf("%s is not a valid sort direction. Should be either asc or desc", s.Direction)
	}

	return fmt.Sprintf("%s %s", s.Value, s.Direction), nil
}

func (qp *QueryParser) ParseSortClauses(s []Sort) (r []string, err error) {

	for _, sort := range s {
		os, err := qp.ParseSort(sort)
		if err != nil {
			return r, err
		}
		r = append(r, os)
	}

	return r, nil
}
