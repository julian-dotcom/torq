package payments

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	qp "github.com/lncapital/torq/internal/query_parser"
	"github.com/lncapital/torq/pkg/server_errors"
	"net/http"
	"strconv"
)

func getPaymentsHandler(c *gin.Context, db *sqlx.DB) {

	// Filter parser with whitelisted columns
	var filter sq.Sqlizer
	filterParam := c.Query("filters")
	var err error
	if filterParam != "" {
		filter, err = qp.ParseFilterParam(filterParam, []string{
			"date",
			"destination_pub_key",
			"status",
			"value_msat",
			"fee_msat",
			"failure_reason",
			"is_rebalance",
			"is_mpp",
			"count_successful_attempts",
			"count_failed_attempts",
			"seconds_in_flight",
		})
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
			return
		}
	}

	var sort []string
	sortParam := c.Query("sortBy")
	if sortParam != "" {
		// Sort parser with whitelisted columns
		sort, err = qp.ParseSortParams(
			sortParam,
			[]string{
				"date",
				"status",
				"value_msat",
				"fee_msat",
				"failure_reason",
				"count_successful_attempts",
				"count_failed_attempts",
				"seconds_in_flight",
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

	r, err := getPayments(db, filter, sort, limit, offset)
	if err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}

	c.JSON(http.StatusOK, r)
}

func RegisterPaymentsRoutes(r *gin.RouterGroup, db *sqlx.DB) {
	r.GET("", func(c *gin.Context) { getPaymentsHandler(c, db) })
}
