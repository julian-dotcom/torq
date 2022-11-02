package channel_history

import (
	"time"

	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"gopkg.in/guregu/null.v4"

	"github.com/lncapital/torq/internal/channels"
)

type channel struct {
	FirstNodeId  null.String `json:"firstNodeId" db:"first_node_id"`
	SecondNodeId null.String `json:"secondNodeId" db:"second_node_id"`
	// Database primary key of channel
	ChannelID              null.Int `json:"channelId" db:"channel_id"`
	FundingTransactionHash string   `json:"fundingTransactionHash" db:"funding_transaction_hash"`
	FundingOutputIndex     string   `json:"fundingOutputIndex" db:"funding_output_index"`
	ClosingTransactionHash *string  `json:"closingTransactionHash" db:"closing_transaction_hash"`
	// Short channel id in c-lightning / BOLT format
	ShortChannelID null.String `json:"shortChannelId" db:"short_channel_id"`
	// The channel ID
	LNDShortChannelId null.String     `json:"chanId" db:"lnd_short_channel_id"`
	Status            channels.Status `json:"status" db:"status_id"`
	CreatedOn         time.Time       `json:"createdOn" db:"created_on"`
	UpdatedOn         *time.Time      `json:"updatedOn" db:"updated_on"`

	// TODO FIXME Capacity shouldn't require a JOIN and should be included in channel table
	// Capacity The channels total capacity (as created). Obtained via join from channel_event
	Capacity *uint64 `json:"capacity" db:"capacity"`
}

func getChannels(db *sqlx.DB, all bool, channelIds []int) (r []*channel, err error) {

	sql := `
		select ce.capacity, c.*
		from (
			select
			    channel_id,
			    last(event->'capacity', time) as capacity
			from channel_event
			where event_type in (0,1) and ($1 or channel_id = ANY ($2))
			group by channel_id
		) as ce
		join channel as c on c.channel_id = ce.channel_id;
	`

	rows, err := db.Queryx(sql, all, pq.Array(channelIds))
	if err != nil {
		return nil, errors.Wrapf(err, "Running getChannelsByPubkey query")
	}

	for rows.Next() {
		c := &channel{}
		err = rows.StructScan(&c)
		if err != nil {
			return r, err
		}

		// Append to the result
		r = append(r, c)
	}
	return r, nil
}
