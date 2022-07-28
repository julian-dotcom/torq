package query_parser

type QueryParser struct {
	AllowedColumns []string
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
