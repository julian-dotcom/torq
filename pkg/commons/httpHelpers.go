package commons

type ShortChannelIdHttpRequest struct {
	TransactionHash string `json:"transactionHash"`
	OutputIndex     int    `json:"outputIndex"`
	UnixTime        int64  `json:"unixTime"`
	Signature       string `json:"signature"`
}

type ShortChannelIdHttpResponse struct {
	Request        ShortChannelIdHttpRequest `json:"request"`
	ShortChannelId string                    `json:"shortChannelId"`
}
