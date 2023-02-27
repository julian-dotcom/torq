package messages

import (
	"time"

	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"

	"github.com/lncapital/torq/pkg/commons"
)

func verifyMessage(db *sqlx.DB, req VerifyMessageRequest, lightningRequestChannel chan interface{}) (VerifyMessageResponse, error) {
	if req.NodeId == 0 {
		return VerifyMessageResponse{}, errors.New("Node Id missing")
	}
	response := commons.SignatureVerification(time.Now(), req.NodeId, req.Message, req.Signature, lightningRequestChannel)
	if response.Status != commons.Active {
		return VerifyMessageResponse{}, errors.New(response.Error)
	}
	return VerifyMessageResponse{
		Valid:  response.Valid,
		PubKey: response.PublicKey,
	}, nil
}
