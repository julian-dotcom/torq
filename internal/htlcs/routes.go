package htlcs

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"

	"github.com/lncapital/torq/pkg/commons"
	"github.com/lncapital/torq/pkg/server_errors"
)

func RegisterHtlcRoutes(r *gin.RouterGroup, db *sqlx.DB) {
	r.GET("", func(c *gin.Context) { getHtlcsHandler(c, db) })
}

func getHtlcsHandler(c *gin.Context, db *sqlx.DB) {
	from, err := time.Parse("2006-01-02", c.Query("from"))
	if err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}
	to, err := time.Parse("2006-01-02", c.Query("to"))
	if err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}
	if c.Query("network") == "" {
		server_errors.SendBadRequest(c, "Network missing")
		return
	}
	network, err := strconv.Atoi(c.Query("network"))
	if err != nil {
		server_errors.SendBadRequest(c, "Can't process network")
		return
	}

	chain := commons.Bitcoin

	r, err := getHtlcs(db, commons.GetAllTorqNodeIds(chain, commons.Network(network)), from, to)
	if err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}
	c.JSON(http.StatusOK, r)
}
