package payments

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/lncapital/torq/pkg/server_errors"
	"net/http"
)

func getPaymentsHandler(c *gin.Context, db *sqlx.DB) {

	/*
		birdJson := `{"$and":[{"$filter":{"funcName":"eq","category":"number","key":"status","parameter":"SUCCEEDED"}},{"$filter":{"funcName":"eq","category":"number","key":"status","parameter":"SUCCEEDED"}},{"$or":[{"$filter":{"funcName":"lte","category":"number","key":"seconds_in_flight","parameter":"2"}},{"$filter":{"funcName":"lte","category":"number","key":"count_successful_attempts","parameter":"10"}}]}]}`
	*/
	filterQuery := c.Query("filter")
	var filter FilterClauses
	err := json.Unmarshal([]byte(filterQuery), &filter)
	if err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}

	// "sortBy":[{"value":"revenue_out","label":"Revenue","direction":"desc"}]
	//sortByQuery := c.Query("sortBy")

	f := ParseFiltersParams(filter)

	r, err := getPayments(db, f, 200, 0)
	if err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}

	c.JSON(http.StatusOK, r)
}

func RegisterPaymentsRoutes(r *gin.RouterGroup, db *sqlx.DB) {
	r.GET("", func(c *gin.Context) { getPaymentsHandler(c, db) })
}
