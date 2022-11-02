package graph_events

import (
	"time"
)

type NodeEventFromGraph struct {
	EventTime     time.Time `json:"eventTime" db:"timestamp"`
	EventNodeId   int       `json:"eventNodeId" db:"event_node_id"`
	PublicKey     string    `json:"publicKey" db:"public_key"`
	Alias         string    `json:"alias" db:"alias"`
	Color         string    `json:"color" db:"color"`
	NodeAddresses string    `json:"node_addresses" db:"node_addresses"`
	Features      string    `json:"features" db:"features"`
	NodeId        int       `json:"nodeId" db:"node_id"`
}

type ChannelEventFromGraph struct {
	EventTime        time.Time `json:"eventTime" db:"ts"`
	ChannelId        int       `json:"channelId" db:"channel_id"`
	AnnouncingNodeId int       `json:"announcingNodeId" db:"announcing_node_id"`
	ConnectingNodeId int       `json:"connectingNodeId" db:"connecting_node_id"`
	Disabled         bool      `json:"disabled" db:"disabled"`
	Outbound         bool      `json:"outbound" db:"outbound"`
	TimeLockDelta    uint32    `json:"timeLockDelta" db:"time_lock_delta"`
	MinHtlc          int64     `json:"minHtlc" db:"min_htlc"`
	MaxHtlcMsat      uint64    `json:"maxHtlcMsat" db:"max_htlc_msat"`
	FeeBaseMsat      int64     `json:"feeBaseMsat" db:"fee_base_msat"`
	FeeRateMilliMsat int64     `json:"feeRateMilliMsat" db:"fee_rate_mill_msat"`
	NodeId           int       `json:"nodeId" db:"node_id"`
}
