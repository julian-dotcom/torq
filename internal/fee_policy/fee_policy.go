package fee_policy

import "time"

type FeePolicy struct {
	FeePolicyId         int       `json:"feePolicyId" db:"fee_policy_id"`
	FeePolicyType       int       `json:"feePolicyType" db:"fee_policy_type"`
	Name                string    `json:"name" db:"name"`
	IncludePendingHTLCs bool      `json:"includePendingHTLCs" db:"include_pending_htlcs"`
	AggregateOnPeer     bool      `json:"aggregateOnPeer" db:"aggregate_on_peer"`
	MinHTLC             *int      `json:"minHTLC" db:"min_htlc"`
	MaxHTLC             *int      `json:"maxHTLC" db:"max_htlc"`
	MinRatio            *int      `json:"minRatio" db:"min_ratio"`
	MaxRatio            *int      `json:"maxRatio" db:"max_ratio"`
	MinBalance          *int      `json:"minBalance" db:"min_balance"`
	MaxBalance          *int      `json:"maxBalance" db:"max_balance"`
	Active              bool      `json:"active" db:"active"`
	Interval            int       `json:"interval" db:"interval"`
	LastRunOn           time.Time `json:"lastRunOn" db:"last_run_on"`
	CreatedOn           time.Time `json:"createdOn" db:"created_on"`
	UpdatedOn           time.Time `json:"updatedOn" db:"updated_on"`
}
