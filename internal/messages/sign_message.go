package messages

import (
	"time"

	"github.com/cockroachdb/errors"

	"github.com/lncapital/torq/pkg/commons"
)

func signMessage(req SignMessageRequest, lightningRequestChannel chan<- interface{}) (SignMessageResponse, error) {
	if req.NodeId == 0 {
		return SignMessageResponse{}, errors.New("Node Id missing")
	}
	response := commons.SignMessage(time.Now(), req.NodeId, req.Message, req.SingleHash, lightningRequestChannel)
	if response.Status != commons.Active {
		return SignMessageResponse{}, errors.New(response.Error)
	}
	return SignMessageResponse{
		Signature: response.Signature,
	}, nil
}
