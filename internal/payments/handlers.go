package payments

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/lncapital/torq/pkg/server_errors"
	"net/http"
)

type FilterClauses struct {
	And    []FilterClauses `json:"$and"`
	Or     []FilterClauses `json:"$or"`
	Filter Filter          `json:"$filter"`
}

type Filter struct {
	FuncName  string `json:"funcName"`
	Category  string `json:"category"`
	Key       string `json:"key"`
	Parameter string `json:"parameter"`
}

func getPaymentsHandler(c *gin.Context, db *sqlx.DB) {

	qpp := QueryPaymentsParams{}

	birdJson := `{"$and":[{"$filter":{"funcName":"gte","category":"number","key":"capacity","parameter":"100000"}},{"$or":[{"$filter":{"funcName":"gte","category":"number","key":"amount_out","parameter":"200"}},{"$filter":{"funcName":"gte","category":"number","key":"amount_total","parameter":"10000"}}]}]}`

	var filter FilterClauses //map[string]interface{}
	err := json.Unmarshal([]byte(birdJson), &filter)
	if err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}

	//pf, err := ParseFiltersParams(filter)
	//if err != nil {
	//	server_errors.LogAndSendServerError(c, err)
	//	return
	//}
	fmt.Println(filter)
	//from, err := time.Parse("2006-01-02", c.Query("from"))
	//if err != nil {
	//	server_errors.LogAndSendServerError(c, err)
	//	return
	//}
	//
	//to, err := time.Parse("2006-01-02", c.Query("to"))
	//if err != nil {
	//	server_errors.LogAndSendServerError(c, err)
	//	return
	//}
	//
	//_, isRebalance := c.Params.Get("is_rebalance")
	//
	//status := strings.Split(c.Param("status"), ",")

	r, err := getPayments(db, qpp, 200, 0)
	if err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}

	c.JSON(http.StatusOK, r)
}

func RegisterPaymentsRoutes(r *gin.RouterGroup, db *sqlx.DB) {
	r.GET("", func(c *gin.Context) { getPaymentsHandler(c, db) })
}
