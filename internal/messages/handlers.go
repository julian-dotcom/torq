package messages

import (
	"github.com/cockroachdb/errors"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/lncapital/torq/pkg/server_errors"
	"net/http"
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

func signMessageHandler(c *gin.Context, db *sqlx.DB) {
	var signMsgReq SignMessageRequest

	if err := c.BindJSON(&signMsgReq); err != nil {
		server_errors.SendBadRequestFromError(c, errors.Wrap(err, server_errors.JsonParseError))
		return
	}

	response, err := signMessage(db, signMsgReq)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Sign message")
		return
	}

	c.JSON(http.StatusOK, response)
}

func verifyMessageHandler(c *gin.Context, db *sqlx.DB) {
	var verifyMsgReq VerifyMessageRequest

	if err := c.BindJSON(&verifyMsgReq); err != nil {
		server_errors.SendBadRequestFromError(c, errors.Wrap(err, server_errors.JsonParseError))
		return
	}

	response, err := verifyMessage(db, verifyMsgReq)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Verify message")
		return
	}

	c.JSON(http.StatusOK, response)
}

func RegisterMessagesRoutes(r *gin.RouterGroup, db *sqlx.DB) {
	r.POST("sign", func(c *gin.Context) { signMessageHandler(c, db) })
	r.POST("verify", func(c *gin.Context) { verifyMessageHandler(c, db) })
}
