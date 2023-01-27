package messages

import (
	"net/http"

	"github.com/cockroachdb/errors"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"

	"github.com/lncapital/torq/pkg/server_errors"
)

type SignMessageRequest struct {
	NodeId     int    `json:"nodeId"`
	Message    string `json:"message"`
	SingleHash *bool  `json:"singleHash"`
}

type SignMessageResponse struct {
	Signature string `json:"signature"`
}

type VerifyMessageRequest struct {
	NodeId    int    `json:"nodeId"`
	Message   string `json:"message"`
	Signature string `json:"signature"`
}

type VerifyMessageResponse struct {
	Valid  bool   `json:"valid"`
	PubKey string `json:"pubKey"`
}

func signMessageHandler(c *gin.Context, db *sqlx.DB, lightningRequestChannel chan interface{}) {
	var signMsgReq SignMessageRequest

	if err := c.BindJSON(&signMsgReq); err != nil {
		server_errors.SendBadRequestFromError(c, errors.Wrap(err, server_errors.JsonParseError))
		return
	}

	response, err := signMessage(signMsgReq, lightningRequestChannel)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Sign message")
		return
	}

	c.JSON(http.StatusOK, response)
}

func verifyMessageHandler(c *gin.Context, db *sqlx.DB, lightningRequestChannel chan interface{}) {
	var verifyMsgReq VerifyMessageRequest

	if err := c.BindJSON(&verifyMsgReq); err != nil {
		server_errors.SendBadRequestFromError(c, errors.Wrap(err, server_errors.JsonParseError))
		return
	}

	response, err := verifyMessage(db, verifyMsgReq, lightningRequestChannel)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Verify message")
		return
	}

	c.JSON(http.StatusOK, response)
}

func RegisterMessagesRoutes(r *gin.RouterGroup, db *sqlx.DB, lightningRequestChannel chan interface{}) {
	r.POST("sign", func(c *gin.Context) { signMessageHandler(c, db, lightningRequestChannel) })
	r.POST("verify", func(c *gin.Context) { verifyMessageHandler(c, db, lightningRequestChannel) })
}
