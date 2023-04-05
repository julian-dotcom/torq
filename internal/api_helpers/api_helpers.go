package api_helpers

type Pagination struct {
	Total  uint64 `json:"total"`
	Limit  uint64 `json:"limit"`
	Offset uint64 `json:"offset"`
}

type ApiResponse struct {
	Data       interface{} `json:"data"`
	Pagination Pagination  `json:"pagination"`
}
