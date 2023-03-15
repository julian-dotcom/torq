package messages

import (
	"time"

	"github.com/cockroachdb/errors"

	"github.com/lncapital/torq/pkg/commons"
)

func verifyMessage(req VerifyMessageRequest, lightningRequestChannel chan<- interface{}) (VerifyMessageResponse, error) {
	if req.NodeId == 0 {
		return VerifyMessageResponse{}, errors.New("Node Id missing")
	}
	response := commons.SignatureVerification(time.Now(), req.NodeId, req.Message, req.Signature, lightningRequestChannel)
	if response.Status == commons.Active || response.Message == "Signature is not valid" {
		return VerifyMessageResponse{
			Valid:  response.Valid,
			PubKey: response.PublicKey,
		}, nil
	}
	return VerifyMessageResponse{}, errors.New(response.Error)
}
