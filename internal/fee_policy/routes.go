package fee_policy

import (
	"net/http"

	"github.com/cockroachdb/errors"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/lncapital/torq/pkg/server_errors"
)

func RegisterFeePolicyRoutes(r *gin.RouterGroup, db *sqlx.DB) {
	r.POST("", func(c *gin.Context) { addFeePolicyHandler(c, db) })
	r.GET("", func(c *gin.Context) { getFeePolicyListHandler(c, db) })
}

func addFeePolicyHandler(c *gin.Context, db *sqlx.DB) {
	var requestBody FeePolicy
	if err := c.BindJSON(&requestBody); err != nil {
		server_errors.SendBadRequestFromError(c, errors.Wrap(err, server_errors.JsonParseError))
		return
	}

	feePolicy, err := addFeePolicy(db, requestBody)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Inserting fee policy")
		return
	}

	c.JSON(http.StatusOK, feePolicy)
	return
}

func getFeePolicyListHandler(c *gin.Context, db *sqlx.DB) {
	feePolicies, err := getFeePolices(db)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Getting fee policies")
		return
	}
	c.JSON(http.StatusOK, feePolicies)
	return
}
