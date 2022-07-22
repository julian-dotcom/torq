package on_chain_tx

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	qp "github.com/lncapital/torq/internal/query_parser"
	"github.com/lncapital/torq/pkg/server_errors"
	"net/http"
	"strconv"
)

func getOnChainTxsHandler(c *gin.Context, db *sqlx.DB) {

	// Filter parser with whitelisted columns
	var filter sq.Sqlizer
	filterParam := c.Query("filter")
	var err error
	if filterParam != "" {
		filter, err = qp.ParseFilterParam(filterParam, []string{
			"date",
			"dest_addresses",
			"dest_addresses_count",
			"amount_msat",
			"total_fees_msat",
			"label",
			"lnd_tx_type_label",
			"lnd_short_chan_id",
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
				"dest_addresses_count",
				"amount_msat",
				"total_fees_msat",
			})
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
			return
		}
	}

	var limit uint64 = 100
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

	var offset uint64 = 0
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

	r, err := getOnChainTxs(db, filter, sort, limit, offset)
	if err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}

	c.JSON(http.StatusOK, r)
}

// "tx_hash", "raw_tx_hex",
//func getOnChainTxHandler(c *gin.Context, db *sqlx.DB) {
//
//	r, err := getOnChainTxDetails(db, c.Param("identifier"))
//	switch err.(type) {
//	case nil:
//		break
//	case ErrOnChainTxNotFound:
//		c.JSON(http.StatusNotFound, gin.H{"Error": err.Error(), "Identifier": c.Param("identifier")})
//		return
//	default:
//		server_errors.LogAndSendServerError(c, err)
//		return
//	}
//
//	c.JSON(http.StatusOK, r)
//}

func RegisterOnChainTxsRoutes(r *gin.RouterGroup, db *sqlx.DB) {
	r.GET("", func(c *gin.Context) { getOnChainTxsHandler(c, db) })
	//r.GET(":identifier", func(c *gin.Context) { getOnChainTxHandler(c, db) })
}
