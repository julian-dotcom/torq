package channels

import (
	"database/sql"

	"github.com/gin-gonic/gin"

	// "gopkg.in/guregu/null.v4"
	"net/http"
	"time"

	"github.com/lncapital/torq/pkg/server_errors"

	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
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

type channel struct {
	// The channel ID
	ChanId uint64 `json:"chan_id"`
	// Alias of remote peer
	Alias string `json:"alias"`
	// The remote public key
	PubKey string `json:"pub_key"`
	// The inbound amount in sats (Satoshis)
	AmountIn uint64 `json:"amount_in"`
	// The inbound revenue in sats. This is what the channel has indirectly produced.
	// This revenue are not really earned by this channel/peer/group, but represents
	// the channel/peer/group contribution to revenue earned by other channels.
	RevenueIn uint64 `json:"revenue_in"`
	// Number of inbound forwards.
	CountIn uint64 `json:"count_in"`
	// The  utbound amount in sats (Satoshis)
	AmountOut uint64 `json:"amount_out"`
	// The outbound revenue in sats. This is what the channel has directly produced.
	RevenueOut uint64 `json:"revenue_out"`
	// Number of outbound forwards.
	CountOut uint64 `json:"count_out"`
}

func getAggForwardsByChanIds(db *sqlx.DB, fromTime time.Time, toTime time.Time, cids []uint64) (r []*channel, err error) {

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
		c := &channel{}
		err = rows.Scan(&c.ChanId,
			&c.Alias,
			&c.AmountIn,
			&c.RevenueIn,
			&c.CountIn,
			&c.AmountOut,
			&c.RevenueOut,
			&c.CountOut,
			&c.PubKey)
		if err != nil {
			return r, err
		}

		// Append to the result
		r = append(r, c)

	}

	return r, nil
}
