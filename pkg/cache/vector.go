package cache

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/lncapital/torq/build"
	"github.com/lncapital/torq/pkg/core"
)

const VectorUrl = "https://vector.ln.capital/"

const vectorShortchannelidUrlSuffix = "api/bitcoin/shortChannelId"
const vectorTransactiondetailsUrlSuffix = "api/bitcoin/transactionDetails"

func GetVectorUrl(suffix string) string {
	return GetVectorUrlBase() + suffix
}

func GetShortChannelIdFromVector(fundingTransactionHash string, fundingOutputIndex int,
	nodeSettings NodeSettingsCache) string {

	unixTime := time.Now()
	requestObject := core.ShortChannelIdHttpRequest{
		TransactionHash: fundingTransactionHash,
		OutputIndex:     fundingOutputIndex,
		UnixTime:        unixTime.Unix(),
		PublicKey:       nodeSettings.PublicKey,
	}
	requestObjectBytes, err := json.Marshal(requestObject)
	if err != nil {
		log.Error().Msgf("Failed (Marshal) to obtain shortChannelId for closed channel with channel point %v:%v",
			fundingTransactionHash, fundingOutputIndex)
		return ""
	}
	req, err := http.NewRequest("GET", GetVectorUrl(vectorShortchannelidUrlSuffix), bytes.NewBuffer(requestObjectBytes))
	if err != nil {
		log.Error().Msgf("Failed (http.NewRequest) to obtain shortChannelId for closed channel with channel point %v:%v",
			fundingTransactionHash, fundingOutputIndex)
		return ""
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Torq-Version", build.ExtendedVersion())
	req.Header.Set("Torq-UUID", GetSettings().TorqUuid)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Error().Msgf("Failed (http.Get) to obtain shortChannelId for closed channel with channel point %v:%v",
			fundingTransactionHash, fundingOutputIndex)
		return ""
	}
	var vectorResponse core.ShortChannelIdHttpResponse
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
	log.Debug().Msgf("Obtained short channel id from vector for channel point %v:%v",
		fundingTransactionHash, fundingOutputIndex)
	return vectorResponse.ShortChannelId
}

func GetTransactionDetailsFromVector(transactionHash string,
	nodeSettings NodeSettingsCache) core.TransactionDetailsHttpResponse {

	unixTime := time.Now()
	requestObject := core.TransactionDetailsHttpRequest{
		TransactionHash: transactionHash,
		UnixTime:        unixTime.Unix(),
		PublicKey:       nodeSettings.PublicKey,
	}
	requestObjectBytes, err := json.Marshal(requestObject)
	if err != nil {
		log.Error().Msgf("Failed (Marshal) to obtain transaction details for transaction hash %v", transactionHash)
		return core.TransactionDetailsHttpResponse{}
	}
	req, err := http.NewRequest("GET", GetVectorUrl(vectorTransactiondetailsUrlSuffix), bytes.NewBuffer(requestObjectBytes))
	if err != nil {
		log.Error().Msgf("Failed (http.NewRequest) to obtain transaction details for transaction hash %v", transactionHash)
		return core.TransactionDetailsHttpResponse{}
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Torq-Version", build.ExtendedVersion())
	req.Header.Set("Torq-UUID", GetSettings().TorqUuid)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Error().Msgf("Failed (http.Get) to obtain transaction details for transaction hash %v", transactionHash)
		return core.TransactionDetailsHttpResponse{}
	}
	var vectorResponse core.TransactionDetailsHttpResponse
	err = json.NewDecoder(resp.Body).Decode(&vectorResponse)
	if err != nil {
		log.Error().Msgf("Failed (Decode) to obtain transaction details for transaction hash %v", transactionHash)
		return core.TransactionDetailsHttpResponse{}
	}
	err = resp.Body.Close()
	if err != nil {
		log.Error().Msgf("Failed (Body.Close) to obtain transaction details for transaction hash %v", transactionHash)
		return core.TransactionDetailsHttpResponse{}
	}
	log.Debug().Msgf("Obtained block height from vector for transaction hash %v", transactionHash)
	return vectorResponse
}
