package commons

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/lncapital/torq/build"
)

func GetShortChannelIdFromVector(fundingTransactionHash string, fundingOutputIndex int, nodeId int,
	lightningRequestChannel chan interface{}) string {

	unixTime := time.Now()
	message := fmt.Sprintf("%v/%v/%v", fundingTransactionHash, fundingOutputIndex, unixTime.Unix())

	responseChannel := make(chan SignMessageResponse)
	lightningRequestChannel <- SignMessageRequest{
		CommunicationRequest: CommunicationRequest{
			RequestId:   fmt.Sprintf("%v", unixTime.Unix()),
			RequestTime: &unixTime,
			NodeId:      nodeId,
		},
		ResponseChannel: responseChannel,
		Message:         message,
	}
	response := <-responseChannel

	requestObject := ShortChannelIdHttpRequest{
		TransactionHash: fundingTransactionHash,
		OutputIndex:     fundingOutputIndex,
		UnixTime:        unixTime.Unix(),
		Signature:       response.Signature,
	}
	requestObjectBytes, err := json.Marshal(requestObject)
	if err != nil {
		log.Error().Msgf("Failed (Marshal) to obtain shortChannelId for closed channel with channel point %v:%v",
			fundingTransactionHash, fundingOutputIndex)
		return ""
	}
	req, err := http.NewRequest("GET", VECTOR_SHORTCHANNELID_URL, bytes.NewBuffer(requestObjectBytes))
	if err != nil {
		log.Error().Msgf("Failed (http.NewRequest) to obtain shortChannelId for closed channel with channel point %v:%v",
			fundingTransactionHash, fundingOutputIndex)
		return ""
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Error().Msgf("Failed (http.Get) to obtain shortChannelId for closed channel with channel point %v:%v",
			fundingTransactionHash, fundingOutputIndex)
		return ""
	}
	var vectorResponse ShortChannelIdHttpResponse
	err = json.NewDecoder(resp.Body).Decode(&vectorResponse)
	if err != nil {
		log.Error().Msgf("Failed (Decode) to obtain shortChannelId for closed channel with channel point %v:%v",
			fundingTransactionHash, fundingOutputIndex)
		return ""
	}
	err = resp.Body.Close()
	if err != nil {
		log.Error().Msgf("Failed (Body.Close) to obtain shortChannelId for closed channel with channel point %v:%v",
			fundingTransactionHash, fundingOutputIndex)
		return ""
	}
	return vectorResponse.ShortChannelId
}

func GetTransactionDetailsFromVector(transactionHash string, nodeId int,
	lightningRequestChannel chan interface{}) TransactionDetailsHttpResponse {

	unixTime := time.Now()
	message := fmt.Sprintf("%v/%v", transactionHash, unixTime.Unix())

	responseChannel := make(chan SignMessageResponse)
	lightningRequestChannel <- SignMessageRequest{
		CommunicationRequest: CommunicationRequest{
			RequestId:   fmt.Sprintf("%v", unixTime.Unix()),
			RequestTime: &unixTime,
			NodeId:      nodeId,
		},
		ResponseChannel: responseChannel,
		Message:         message,
	}
	response := <-responseChannel

	requestObject := TransactionDetailsHttpRequest{
		TransactionHash: transactionHash,
		UnixTime:        unixTime.Unix(),
		Signature:       response.Signature,
	}
	requestObjectBytes, err := json.Marshal(requestObject)
	if err != nil {
		log.Error().Msgf("Failed (Marshal) to obtain transaction details for transaction hash %v", transactionHash)
		return TransactionDetailsHttpResponse{}
	}
	req, err := http.NewRequest("GET", VECTOR_TRANSACTIONDETAILS_URL, bytes.NewBuffer(requestObjectBytes))
	if err != nil {
		log.Error().Msgf("Failed (http.NewRequest) to obtain transaction details for transaction hash %v", transactionHash)
		return TransactionDetailsHttpResponse{}
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Torq-Version", build.Version())
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Error().Msgf("Failed (http.Get) to obtain transaction details for transaction hash %v", transactionHash)
		return TransactionDetailsHttpResponse{}
	}
	var vectorResponse TransactionDetailsHttpResponse
	err = json.NewDecoder(resp.Body).Decode(&vectorResponse)
	if err != nil {
		log.Error().Msgf("Failed (Decode) to obtain transaction details for transaction hash %v", transactionHash)
		return TransactionDetailsHttpResponse{}
	}
	err = resp.Body.Close()
	if err != nil {
		log.Error().Msgf("Failed (Body.Close) to obtain transaction details for transaction hash %v", transactionHash)
		return TransactionDetailsHttpResponse{}
	}
	return vectorResponse
}
