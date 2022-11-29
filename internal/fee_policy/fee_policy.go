package fee_policy

import "time"

type FeePolicy struct {
	FeePolicyId         int       `json:"feePolicyId" db:"fee_policy_id"`
	FeePolicyType       int       `json:"feePolicyType" db:"fee_policy_type"`
	Name                string    `json:"name" db:"name"`
	IncludePendingHTLCs bool      `json:"includePendingHTLCs" db:"include_pending_htlcs"`
	AggregateOnPeer     bool      `json:"aggregateOnPeer" db:"aggregate_on_peer"`
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

type FeePolicyTarget struct {
	FeePolicyTargetId int       `json:"feePolicyTarget" db:"fee_policy_target_id"`
	FeePolicyId       int       `json:"feePolicyId" db:"fee_policy_id"`
	TagId             int       `json:"tagId" db:"tag_id"`
	CategoryId        int       `json:"categoryId" db:"category_id"`
	NodeId            int       `json:"nodeId" db:"node_id"`
	ChannelId         int       `json:"channelId" db:"channel_id"`
	CreatedOn         time.Time `json:"createdOn" db:"created_on"`
	UpdatedOn         time.Time `json:"updatedOn" db:"updated_on"`
}

type FeePolicyStep struct {
	FeePolicyStepId int       `json:"feePolicyStepId" db:"fee_policy_step_id"`
	FeePolicyId     int       `json:"feePolicyId" db:"fee_policy_id"`
	MinHTLC         *int      `json:"minHTLC" db:"min_htlc"`
	MaxHTLC         *int      `json:"maxHTLC" db:"max_htlc"`
	FeeRatePPM      int       `json:"feeRatePPM" db:"fee_rate_ppm"`
	BaseFee         int       `json:"baseFee" db:"base_fee"`
	CreatedOn       time.Time `json:"createdOn" db:"created_on"`
	UpdatedOn       time.Time `json:"updatedOn" db:"updated_on"`
}
