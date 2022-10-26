package channel_history

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/lncapital/torq/pkg/server_errors"
	"net/http"
	"strings"
	"time"
)

type ChannelHistory struct {
	// The Label of the requested channels group,
	// this is either an alias in the case where a single channel or a single node is requested.
	// In the case where a group of channels is requested the Label will be based on the common name,
	// such as a tag.
	Label string `json:"label"`

	OnChainCost     *uint64 `json:"on_chain_cost"`
	RebalancingCost *uint64 `json:"rebalancing_cost"`

	// The  outbound amount in sats (Satoshis)
	AmountOut *uint64 `json:"amount_out"`
	// The inbound amount in sats (Satoshis)
	AmountIn *uint64 `json:"amount_in"`
	// The total amount in sats (Satoshis) forwarded
	AmountTotal *uint64 `json:"amount_total"`

	// The outbound revenue in sats. This is what the channel has directly produced.
	RevenueOut *uint64 `json:"revenue_out"`
	// The inbound revenue in sats. This is what the channel has indirectly produced.
	// This revenue are not really earned by this channel/peer/group, but represents
	// the channel/peer/group contribution to revenue earned by other channels.
	RevenueIn *uint64 `json:"revenue_in"`
	// The total revenue in sats. This is what the channel has directly and indirectly produced.
	RevenueTotal *uint64 `json:"revenue_total"`

	// Number of outbound forwards.
	CountOut *uint64 `json:"count_out"`
	// Number of inbound forwards.
	CountIn *uint64 `json:"count_in"`
	// Number of total forwards.
	CountTotal *uint64 `json:"count_total"`

	// Aggregated details about successful rebalancing (i.g. amount, cost, counts)
	RebalancingDetails RebalancingDetails `json:"rebalancing"`

	// Channel balances over time
	ChannelBalances []*ChannelBalance `json:"channel_balance"`

	// A list of channels included in this response
	Channels []*channel               `json:"channels"`
	History  []*ChannelHistoryRecords `json:"history"`
	Events   []*ChannelEvent          `json:"events"`
}

func getChannelHistoryHandler(c *gin.Context, db *sqlx.DB) {
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

	chanIds := strings.Split(c.Param("chanIds"), ",")

	// Get the total values for the whole requested time range (from - to)
	r, err := getChannelTotal(db, chanIds, from, to)
	if err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}

	// Get the details for the requested channels
	channels, err := getChannels(db, chanIds)
	if err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}
	r.Channels = channels

	// Get the daily values
	chanHistory, err := getChannelHistory(db, chanIds, from, to)
	if err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}
	r.History = chanHistory

	chanEventHistory, err := getChannelEventHistory(db, chanIds, from, to)
	if err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}
	r.Events = chanEventHistory

	if chanIds[0] == "1" {
		r.OnChainCost, err = getTotalOnChainCost(db, from, to)
	} else {
		r.OnChainCost, err = getChannelOnChainCost(db, chanIds)
	}
	if err != nil {
		server_errors.LogAndSendServerError(c, err)
		return
	}

	if chanIds[0] == "1" {
		reb, err := getRebalancingCost(db, from, to)
		r.RebalancingCost = &reb.TotalCostMsat
		r.RebalancingDetails = reb
		if err != nil {
			server_errors.LogAndSendServerError(c, err)
			return
		}
	} else {
		r.OnChainCost, err = getChannelOnChainCost(db, chanIds)
		if err != nil {
			server_errors.LogAndSendServerError(c, err)
			return
		}
		reb, err := getChannelRebalancing(db, chanIds, from, to)
		r.RebalancingCost = &reb.SplitCostMsat
		r.RebalancingDetails = reb
		if err != nil {
			server_errors.LogAndSendServerError(c, err)
			return
		}
	}

	if chanIds[0] != "1" {

		for _, chanId := range chanIds {
			cb, err := getChannelBalance(db, string(chanId), from, to)
			if err != nil {
				server_errors.LogAndSendServerError(c, err)
				return
			}

			if len(r.ChannelBalances) == 0 {
				r.ChannelBalances = []*ChannelBalance{&cb}
			} else {
				r.ChannelBalances = append(r.ChannelBalances, &cb)
			}

		}

	}

	c.JSON(http.StatusOK, r)
}
