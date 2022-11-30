package fee_policy

import "time"

type FeePolicyStrategy int

const (
	policyStrategyStep FeePolicyStrategy = iota
)

type FeePolicy struct {
	FeePolicyId         int               `json:"feePolicyId" db:"fee_policy_id"`
	FeePolicyStrategy   FeePolicyStrategy `json:"feePolicyStrategy" db:"fee_policy_strategy"`
	Name                string            `json:"name" db:"name"`
	IncludePendingHTLCs bool              `json:"includePendingHTLCs" db:"include_pending_htlcs"`
	AggregateOnPeer     bool              `json:"aggregateOnPeer" db:"aggregate_on_peer"`
	Active              bool              `json:"active" db:"active"`
	Interval            int               `json:"interval" db:"interval"`
	Targets             []FeePolicyTarget `json:"targets" db:"-"`
	Steps               []FeePolicyStep   `json:"steps" db:"-"`
	LastRunOn           *time.Time        `json:"lastRunOn" db:"last_run_on"`
	CreatedOn           time.Time         `json:"createdOn" db:"created_on"`
	UpdatedOn           time.Time         `json:"updatedOn" db:"updated_on"`
}

type FeePolicyTarget struct {
	FeePolicyTargetId int       `json:"feePolicyTargetId" db:"fee_policy_target_id"`
	FeePolicyId       int       `json:"feePolicyId" db:"fee_policy_id"`
	TagId             *int      `json:"tagId" db:"tag_id"`
	CategoryId        *int      `json:"categoryId" db:"category_id"`
	NodeId            *int      `json:"nodeId" db:"node_id"`
	ChannelId         *int      `json:"channelId" db:"channel_id"`
	CreatedOn         time.Time `json:"createdOn" db:"created_on"`
	UpdatedOn         time.Time `json:"updatedOn" db:"updated_on"`
}

type FeePolicyStep struct {
	FeePolicyStepId  int       `json:"feePolicyStepId" db:"fee_policy_step_id"`
	FeePolicyId      int       `json:"feePolicyId" db:"fee_policy_id"`
	FilterMinRatio   *int      `json:"filterMinRatio" db:"filter_min_ratio"`
	FilterMaxRatio   *int      `json:"filterMaxRatio" db:"filter_max_ratio"`
	FilterMinBalance *int      `json:"filterMinBalance" db:"filter_min_balance"`
	FilterMaxBalance *int      `json:"filterMaxBalance" db:"filter_max_balance"`
	SetMinHTLC       *int      `json:"setMinHTLC" db:"set_min_htlc"`
	SetMaxHTLC       *int      `json:"setMaxHTLC" db:"set_max_htlc"`
	SetFeePPM        int       `json:"setFeePPM" db:"set_fee_ppm"`
	SetBaseFee       int       `json:"setBaseFee" db:"set_base_fee"`
	CreatedOn        time.Time `json:"createdOn" db:"created_on"`
	UpdatedOn        time.Time `json:"updatedOn" db:"updated_on"`
}
