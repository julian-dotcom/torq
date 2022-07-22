package api_helpers

type Pagination struct {
	Total  uint64 `json:"total"`
	Limit  uint64 `json:"Limit"`
	Offset uint64 `json:"Offset"`
}

type ApiResponse struct {
	Data       interface{}
	Pagination Pagination
}
