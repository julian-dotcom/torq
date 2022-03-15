package channels

import (
	"database/sql"

	"github.com/gin-gonic/gin"

	// "gopkg.in/guregu/null.v4"
	"github.com/lncapital/torq/pkg/server_errors"
	"net/http"
	"time"

	"fmt"
	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/lncapital/torq/torqrpc"
)

func getChannelsHandler(c *gin.Context, db *sqlx.DB) {
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
	r, err := getAggForwardsByChanIds(db, from, to, nil)
	if err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}
	c.JSON(http.StatusOK, r)
}

func getAggForwardsByChanIds(db *sqlx.DB, fromTime time.Time, toTime time.Time, cids []uint64) (r []*torqrpc.AggregatedForwards, err error) {

	var rows *sql.Rows

	// Request given channel IDs, if specified.
	if len(cids) != 0 {

		var q string
		var args []interface{}

		query := "select * from agg_forwards_by_chan_id(?, ?, (?))"
		q, args, err = sqlx.In(query, fromTime, toTime, pq.Array(cids))
		if err != nil {
			return nil, errors.Wrapf(err, "getAggForwardsByChanIds -> sqlx.In(%s, %d, %d, %v)",
				query, fromTime, toTime, cids)
		}

		qs := db.Rebind(q)
		rows, err = db.Query(qs, args...)
		if err != nil {
			return nil, errors.Wrapf(err, "getAggForwardsByChanIds -> db.Query(db.Rebind(qs), args...)")
		}

	} else { // Request all channel IDs if none are given
		rows, err = db.Query("select * from agg_forwards_by_chan_id($1, $2, null)", fromTime, toTime)
		if err != nil {
			return nil, errors.Wrapf(err, "getAggForwardsByChanIds -> "+
				"db.Queryx(\"select * from agg_forwards_by_chan_id(?, ?, null)\", %d, %d)",
				fromTime, toTime)
		}

	}

	for rows.Next() {
		afw := &torqrpc.AggregatedForwards{}
		var chanId uint64
		var alias string
		var pubKey string
		err = rows.Scan(&chanId,
			&alias,
			&afw.AmountIn,
			&afw.RevenueIn,
			&afw.CountIn,
			&afw.AmountOut,
			&afw.RevenueOut,
			&afw.CountOut,
			&pubKey)
		if err != nil {
			return r, err
		}

		// Add the channel Info
		afw.Channels = []*torqrpc.ChanInfo{{
			ChanId: chanId,
			Alias:  alias,
			PubKey: pubKey,
		}}
		afw.GroupType = torqrpc.GroupType_CHANNEL
		afw.GroupId = fmt.Sprintf("%d", chanId)
		afw.GroupName = alias

		// Append to the result
		r = append(r, afw)

	}

	return r, nil
}
