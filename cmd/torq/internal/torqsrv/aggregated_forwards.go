package torqsrv

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/lncapital/torq/torqrpc"
	"time"
)

func (s torqGrpc) GetAggrigatedForwards(ctx context.Context, req *torqrpc.AggregatedForwardsRequest) (
	*torqrpc.AggregatedForwardsResponse, error) {

	resp := torqrpc.AggregatedForwardsResponse{}
	resp.FromTs = req.FromTs
	resp.ToTs = req.ToTs

	switch x := req.Ids.(type) {
	case *torqrpc.AggregatedForwardsRequest_ChannelIds:

		r, err := getAggForwardsByChanIds(s.db, req.FromTs, req.ToTs, req.GetChannelIds().ChanIds)
		if err != nil {
			return nil, errors.Wrapf(err, "GetAggregatedForwards -> getAggForwardsByChanIds(%v, %d, %d, %v)",
				s.db, req.FromTs, req.ToTs, req.GetChannelIds().ChanIds)
		}

		resp.AggregatedForwards = r
		resp.GroupType = torqrpc.GroupType_CHANNEL

		return &resp, nil
	case *torqrpc.AggregatedForwardsRequest_PeerIds:
		// Fetch based on peer ids
		r, err := getAggForwardsByPubKeys(s.db, req.FromTs, req.ToTs, req.GetPeerIds().PubKeys)
		if err != nil {
			fmt.Println(err)
			return nil, errors.Wrapf(err, "GetAggregatedForwards -> getAggForwardsByPubKeys(%v, %d, %d, %v)",
				s.db, req.FromTs, req.ToTs, req.GetPeerIds().PubKeys)
		}

		resp.AggregatedForwards = r
		resp.GroupType = torqrpc.GroupType_PEER

		return &resp, nil
	case *torqrpc.AggregatedForwardsRequest_TagIds:
		// Fetch based on tags
		resp.GroupType = torqrpc.GroupType_TAG

		return &resp, fmt.Errorf("aggregating by tag is not yet implemented")
	case nil:
		return nil, fmt.Errorf("no aggregation type set")
	default:
		return nil, fmt.Errorf("aggregatedForwardsRequest has unexpected Id type %T", x)
	}

	return nil, nil
}

func getAggForwardsByChanIds(db *sqlx.DB, fromTs int64, toTs int64, cids []uint64) (r []*torqrpc.AggregatedForwards, err error) {

	var fromTime time.Time
	toTime := time.Now()

	if fromTs != 0 {
		fromTime = time.Unix(fromTs, 0).UTC()
	}

	if toTs != 0 {
		toTime = time.Unix(toTs, 0).UTC()
	}

	var rows *sql.Rows

	// Request given channel IDs, if specified.
	if len(cids) != 0 {

		var q string
		var args []interface{}

		query := "select * from agg_forwards_by_chan_id(?, ?, (?))"
		q, args, err = sqlx.In(query, fromTime, toTime, pq.Array(cids))
		if err != nil {
			return nil, errors.Wrapf(err, "getAggForwardsByChanIds -> sqlx.In(%s, %d, %d, %v)",
				query, fromTs, toTs, cids)
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
				fromTs, toTs)
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
			fmt.Println("Happened")
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

func getAggForwardsByPubKeys(db *sqlx.DB, fromTs int64, toTs int64, pubKeys []string) (r []*torqrpc.AggregatedForwards, err error) {

	var fromTime time.Time
	toTime := time.Now()

	if fromTs != 0 {
		fromTime = time.Unix(fromTs, 0).UTC()
	}

	if toTs != 0 {
		toTime = time.Unix(toTs, 0).UTC()
	}

	var rows *sql.Rows

	// Request given peers, if any public keys are specified.
	if len(pubKeys) != 0 {

		var q string
		var args []interface{}

		query := "select * from agg_forwards_by_pub_key(?, ?, (?))"
		q, args, err = sqlx.In(query, fromTime, toTime, pq.Array(pubKeys))
		if err != nil {
			return nil, errors.Wrapf(err, "getAggForwardsByChanIds -> sqlx.In(%s, %d, %d, %v)",
				query, fromTs, toTs, pubKeys)
		}

		qs := db.Rebind(q)
		rows, err = db.Query(qs, args...)
		if err != nil {
			return nil, errors.Wrapf(err, "getAggForwardsByChanIds -> db.Queryx(db.Rebind(qs), args...)")
		}

	} else { // Request all peers if none are given
		rows, err = db.Query("select * from agg_forwards_by_pub_key($1, $2, null)", fromTime, toTime)
		if err != nil {
			return nil, errors.Wrapf(err, "getAggForwardsByChanIds -> "+
				"db.Query(\"select * from agg_forwards_by_pub_key(?, ?, null)\", %d, %d)",
				fromTs, toTs)
		}

	}

	for rows.Next() {
		afw := &torqrpc.AggregatedForwards{}
		var groupId string
		var chanIds pq.Int64Array
		var closed pq.BoolArray
		var groupName string
		err = rows.Scan(&groupId,
			&groupName,
			&closed,
			&chanIds,
			&afw.AmountOut,
			&afw.AmountIn,
			&afw.AmountTotal,
			&afw.RevenueOut,
			&afw.RevenueIn,
			&afw.RevenueTotal,
			&afw.CountOut,
			&afw.CountIn,
			&afw.CountTotal,
			&afw.Capacity,
			&afw.Turnover,
		)
		if err != nil {
			return r, err
		}

		// List all channels, with public keys and alias
		afw.Channels = []*torqrpc.ChanInfo{}
		for i, cid := range chanIds {
			// Add the channel Info
			afw.Channels = append(afw.Channels, &torqrpc.ChanInfo{
				ChanId: uint64(cid),
				Alias:  groupName,
				PubKey: groupId,
				Closed: closed[i],
			})
		}

		afw.GroupType = torqrpc.GroupType_PEER
		afw.GroupId = groupId // group_id = remote public key
		afw.GroupName = groupName

		// Append to the result
		r = append(r, afw)

	}

	return r, nil
}
