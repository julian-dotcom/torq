package messages

import (
	"net/http"

	"github.com/cockroachdb/errors"
	"github.com/gin-gonic/gin"

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

func signMessageHandler(c *gin.Context) {
	var signMsgReq SignMessageRequest

	if err := c.BindJSON(&signMsgReq); err != nil {
		server_errors.SendBadRequestFromError(c, errors.Wrap(err, server_errors.JsonParseError))
		return
	}

	response, err := signMessage(signMsgReq)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Sign message")
		return
	}

	c.JSON(http.StatusOK, response)
}

func verifyMessageHandler(c *gin.Context) {
	var verifyMsgReq VerifyMessageRequest

	if err := c.BindJSON(&verifyMsgReq); err != nil {
		server_errors.SendBadRequestFromError(c, errors.Wrap(err, server_errors.JsonParseError))
		return
	}

	response, err := verifyMessage(verifyMsgReq)
	if err != nil {
		serr := server_errors.ServerError{}
		// TODO: Replace with error codes
		serr.AddServerError("Torq could not verify message signature.")
		server_errors.SendBadRequestFieldError(c, &serr)
		return
	}

	c.JSON(http.StatusOK, response)
}

func RegisterMessagesRoutes(r *gin.RouterGroup) {
	r.POST("sign", func(c *gin.Context) { signMessageHandler(c) })
	r.POST("verify", func(c *gin.Context) { verifyMessageHandler(c) })
}
