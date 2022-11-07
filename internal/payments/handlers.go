package payments

import (
	"net/http"
	"strconv"

	sq "github.com/Masterminds/squirrel"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"

	qp "github.com/lncapital/torq/internal/query_parser"
	ah "github.com/lncapital/torq/pkg/api_helpers"
	"github.com/lncapital/torq/pkg/commons"
	"github.com/lncapital/torq/pkg/server_errors"
)

func getPaymentsHandler(c *gin.Context, db *sqlx.DB) {

	// Filter parser with whitelisted columns
	var filter sq.Sqlizer
	filterParam := c.Query("filter")
	var err error
	if filterParam != "" {
		filter, err = qp.ParseFilterParam(filterParam, []string{
			"date",
			"destinationPubKey",
			"status",
			"value",
			"fee",
			"ppm",
			"failureReason",
			"isRebalance",
			"isMpp",
			"countSuccessfulAttempts",
			"countFailedAttempts",
			"secondsInFlight",
			"paymentHash",
			"paymentPreimage",
		})
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
			return
		}
	}

	var sort []string
	sortParam := c.Query("order")
	if sortParam != "" {
		// Order parser with whitelisted columns
		sort, err = qp.ParseOrderParams(
			sortParam,
			[]string{
				"date",
				"status",
				"value",
				"fee",
				"ppm",
				"failureReason",
				"countSuccessfulAttempts",
				"countFailedAttempts",
				"secondsInFlight",
			})
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
			return
		}
	}

	var limit uint64
	if c.Query("limit") != "" {
		limit, err = strconv.ParseUint(c.Query("limit"), 10, 64)
		switch err.(type) {
		case nil:
			break
		case *strconv.NumError:
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Limit must be a positive number"})
			return
		default:
			server_errors.LogAndSendServerError(c, err)
		}
		if limit == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Limit must be a at least 1"})
			return
		}
	}

	var offset uint64
	if c.Query("offset") != "" {
		offset, err = strconv.ParseUint(c.Query("offset"), 10, 64)
		switch err.(type) {
		case nil:
			break
		case *strconv.NumError:
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Offset must be a positive number"})
			return
		default:
			server_errors.LogAndSendServerError(c, err)
		}
	}

	network := c.Query("network")
	chain := c.Query("chain")

	r, total, err := getPayments(db, commons.GetAllTorqNodeIds(commons.GetChain(chain), commons.GetNetwork(network)), filter, sort, limit, offset)
	if err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}

	c.JSON(http.StatusOK, ah.ApiResponse{
		Data: r, Pagination: ah.Pagination{
			Total:  total,
			Limit:  limit,
			Offset: offset,
		}})
}

func getPaymentHandler(c *gin.Context, db *sqlx.DB) {
	network := c.Query("network")
	chain := c.Query("chain")

	r, err := getPaymentDetails(db, commons.GetAllTorqNodeIds(commons.GetChain(chain), commons.GetNetwork(network)), c.Param("identifier"))
	switch err.(type) {
	case nil:
		break
	case ErrPaymentNotFound:
		c.JSON(http.StatusNotFound, gin.H{"Error": err.Error(), "Identifier": c.Param("identifier")})
		return
	default:
		server_errors.LogAndSendServerError(c, err)
		return
	}

	c.JSON(http.StatusOK, r)
}
