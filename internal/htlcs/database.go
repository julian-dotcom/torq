package htlcs

import (
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/pkg/errors"

	"github.com/lncapital/torq/internal/tags"
	"github.com/lncapital/torq/pkg/commons"
)

type htlcTableRow struct {
	EventTime   time.Time `json:"eventTime" db:"event_time"`
	EventType   string    `json:"eventType" db:"event_type"`
	EventOrigin string    `json:"eventOrigin" db:"event_origin"`

	IncomingChannelId         *int                   `json:"incomingChannelId" db:"incoming_channel_id"`
	IncomingShortChannelId    *string                `json:"incomingShortChannelId" db:"incoming_short_channel_id"`
	IncomingLndShortChannelId *string                `json:"incomingLndShortChannelId" db:"incoming_lnd_short_channel_id"`
	IncomingChannelTags       *[]tags.Tag            `json:"incomingChannelTags" db:""`
	IncomingAlias             *string                `json:"incomingAlias" db:"incoming_alias"`
	IncomingPublicKey         *string                `json:"incomingPublicKey" db:"incoming_public_key"`
	IncomingNodeId            *int                   `json:"incomingNodeId" db:"incoming_node_id"`
	IncomingChannelStatus     *commons.ChannelStatus `json:"incomingChannelStatus" db:"incoming_channel_status_id"`
	IncomingChannelCapacity   *uint64                `json:"incomingChannelCapacity" db:"incoming_channel_capacity"`
	IncomingTimeLock          *uint32                `json:"incomingTimeLock" db:"incoming_timelock"`
	IncomingAmountMsat        *uint64                `json:"incomingAmountMsat" db:"incoming_amount_msat"`

	OutgoingChannelId         *int                   `json:"outgoingChannelId" db:"outgoing_channel_id"`
	OutgoingShortChannelId    *string                `json:"outgoingShortChannelId" db:"outgoing_short_channel_id"`
	OutgoingLndShortChannelId *string                `json:"outgoingLndShortChannelId" db:"outgoing_lnd_short_channel_id"`
	OutgoingChannelTags       *[]tags.Tag            `json:"outgoingChannelTags" db:""`
	OutgoingAlias             *string                `json:"outgoingAlias" db:"outgoing_alias"`
	OutgoingPublicKey         *string                `json:"outgoingPublicKey" db:"outgoing_public_key"`
	OutgoingNodeId            *int                   `json:"outgoingNodeId" db:"outgoing_node_id"`
	OutgoingChannelStatus     *commons.ChannelStatus `json:"outgoingChannelStatus" db:"outgoing_channel_status_id"`
	OutgoingChannelCapacity   *uint64                `json:"outgoingChannelCapacity" db:"outgoing_channel_capacity"`
	OutgoingTimeLock          *uint32                `json:"outgoingTimeLock" db:"outgoing_timelock"`
	OutgoingAmountMsat        *uint64                `json:"outgoingAmountMsat" db:"outgoing_amount_msat"`

	BoltFailureCode   *string `json:"boltFailureCode" db:"bolt_failure_code"`
	BoltFailureString *string `json:"boltFailureString" db:"bolt_failure_string"`
	LndFailureDetail  *string `json:"lndFailureDetail" db:"lnd_failure_detail"`

	NodeId int `json:"nodeId" db:"node_id"`
}

func getHtlcs(db *sqlx.DB, nodeIds []int, fromTime time.Time, toTime time.Time) ([]htlcTableRow, error) {

	var sqlString = `
			SELECT h.event_time, h.event_type, h.event_origin,
				h.incoming_timelock,
				h.incoming_amount_msat,
				h.incoming_channel_id,
				h.incoming_short_channel_id,
				h.incoming_lnd_short_channel_id,
				h.incoming_channel_status_id,
				h.incoming_channel_capacity,
				h.incoming_node_id,
				h.outgoing_timelock,
				h.outgoing_amount_msat,
				h.outgoing_channel_id,
				h.outgoing_short_channel_id,
				h.outgoing_lnd_short_channel_id,
				h.outgoing_channel_status_id,
				h.outgoing_channel_capacity,
				h.outgoing_node_id,
				h.bolt_failure_code,
				h.bolt_failure_string,
				h.lnd_failure_detail,
				h.node_id,
				ino.public_key incoming_public_key, ono.public_key outgoing_public_key,
				LAST(ine.alias, one.timestamp) incoming_alias, LAST(one.alias, one.timestamp) outgoing_alias
			FROM (
				SELECT he.time event_time, he.event_type, he.event_origin,
					he.incoming_timelock,
					he.incoming_amt_msat incoming_amount_msat,
					ic.channel_id incoming_channel_id,
					ic.short_channel_id incoming_short_channel_id,
					ic.lnd_short_channel_id incoming_lnd_short_channel_id,
					ic.status_id incoming_channel_status_id,
					ic.capacity incoming_channel_capacity,
					CASE
					  WHEN ic.first_node_id=he.node_id THEN ic.second_node_id
					  ELSE ic.first_node_id
					END incoming_node_id,

					he.outgoing_timelock,
					he.outgoing_amt_msat outgoing_amount_msat,
					oc.channel_id outgoing_channel_id,
					oc.short_channel_id outgoing_short_channel_id,
					oc.lnd_short_channel_id outgoing_lnd_short_channel_id,
					oc.status_id outgoing_channel_status_id,
					oc.capacity outgoing_channel_capacity,
					CASE
					  WHEN oc.first_node_id=he.node_id THEN oc.second_node_id
					  ELSE oc.first_node_id
					END outgoing_node_id,

					he.bolt_failure_code,
					CASE
      				  WHEN LENGTH(he.bolt_failure_string) > 50 THEN CONCAT(LEFT(he.bolt_failure_string, 50), '...')
      				  ELSE he.bolt_failure_string
      				END bolt_failure_string,
					he.lnd_failure_detail,
					he.node_id
				FROM htlc_event he
				LEFT JOIN channel ic ON ic.channel_id = he.incoming_channel_id
				LEFT JOIN channel oc ON oc.channel_id = he.outgoing_channel_id
				WHERE he.time::timestamp AT TIME ZONE ($3) >= $1::timestamp AND
					  he.time::timestamp AT TIME ZONE ($3) <= $2::timestamp AND
					  he.node_id = ANY($4) AND
					  he.bolt_failure_code IS NOT NULL AND
					  he.bolt_failure_code <> ''
				ORDER BY he.time DESC
				LIMIT 100
			) AS h
			JOIN node ino ON ino.node_id=h.incoming_node_id
			JOIN node_event ine ON ine.event_node_id=h.incoming_node_id
			JOIN node ono ON ono.node_id=h.outgoing_node_id
			JOIN node_event one ON one.event_node_id=h.outgoing_node_id
			GROUP BY h.event_time, h.event_type, h.event_origin,
				h.incoming_timelock,
				h.incoming_amount_msat,
				h.incoming_channel_id,
				h.incoming_short_channel_id,
				h.incoming_lnd_short_channel_id,
				h.incoming_channel_status_id,
				h.incoming_channel_capacity,
				h.incoming_node_id,
				h.outgoing_timelock,
				h.outgoing_amount_msat,
				h.outgoing_channel_id,
				h.outgoing_short_channel_id,
				h.outgoing_lnd_short_channel_id,
				h.outgoing_channel_status_id,
				h.outgoing_channel_capacity,
				h.outgoing_node_id,
				h.bolt_failure_code,
				h.bolt_failure_string,
				h.lnd_failure_detail,
				h.node_id,
				ino.public_key, ono.public_key
			ORDER BY h.event_time DESC;
`

	rows, err := db.Queryx(sqlString, fromTime, toTime, commons.GetSettings().PreferredTimeZone, pq.Array(nodeIds))
	if err != nil {
		return nil, errors.Wrapf(err, "Running aggregated forwards query")
	}

	var results []htlcTableRow
	for rows.Next() {
		var h htlcTableRow
		err = rows.StructScan(&h)
		if err != nil {
			return nil, errors.Wrap(err, "SQL row scan")
		}
		results = append(results, h)
	}

	return results, nil
}
