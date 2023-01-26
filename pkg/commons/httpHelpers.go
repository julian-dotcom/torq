package commons

import "time"

type ShortChannelIdHttpRequest struct {
	TransactionHash string `json:"transactionHash"`
	OutputIndex     int    `json:"outputIndex"`
	UnixTime        int64  `json:"unixTime"`
	Signature       string `json:"signature"`
	PublicKey       string `json:"publicKey"`
}

type ShortChannelIdHttpResponse struct {
	Request        ShortChannelIdHttpRequest `json:"request"`
	ShortChannelId string                    `json:"shortChannelId"`
}

type TransactionDetailsHttpRequest struct {
	TransactionHash string `json:"transactionHash"`
	UnixTime        int64  `json:"unixTime"`
	Signature       string `json:"signature"`
	PublicKey       string `json:"publicKey"`
}

type TransactionDetailsHttpResponse struct {
	Request          TransactionDetailsHttpRequest `json:"request"`
	TransactionCount int                           `json:"transactionCount"`
	TransactionIndex int                           `json:"transactionIndex"`
	BlockHash        string                        `json:"blockHash"`
	BlockTimestamp   time.Time                     `json:"blockTimestamp"`
	BlockHeight      uint32                        `json:"blockHeight"`
}
