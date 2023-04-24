package on_chain_tx

import (
	"net/http"
	"strconv"

	sq "github.com/Masterminds/squirrel"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"

	ah "github.com/lncapital/torq/internal/api_helpers"
	"github.com/lncapital/torq/internal/cache"
	"github.com/lncapital/torq/internal/core"
	qp "github.com/lncapital/torq/internal/query_parser"
	"github.com/lncapital/torq/pkg/server_errors"
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
			"amount",
			"total_fees",
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
				"dest_addresses",
				"dest_addresses_count",
				"amount",
				"total_fees",
				"label",
				"lnd_tx_type_label",
				"lnd_short_chan_id",
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

	network, err := strconv.Atoi(c.Query("network"))
	if err != nil {
		server_errors.SendBadRequest(c, "Can't process network")
		return
	}

	chain := core.Bitcoin

	r, total, err := getOnChainTxs(db, cache.GetAllTorqNodeIdsByNetwork(chain, core.Network(network)), filter, sort, limit, offset)
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
