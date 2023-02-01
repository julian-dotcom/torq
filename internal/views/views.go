package views

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx/types"
	"golang.org/x/exp/slices"
)

type TableViewPage string

const (
	PageForwards = TableViewPage("forwards")
	PageChannel  = TableViewPage("channel")
	PagePayments = TableViewPage("payments")
	PageInvoices = TableViewPage("invoices")
	PageOnChain  = TableViewPage("onChain")
	PageWorkflow = TableViewPage("workflow")
	PageTag      = TableViewPage("tag")
)

type NewTableView struct {
	View types.JSONText `json:"view" db:"view"`
	Page string         `json:"page" db:"page"`
}

type UpdateTableView struct {
	Id      int            `json:"id" db:"id"`
	View    types.JSONText `json:"view" db:"view"`
	Version string         `json:"version" db:"version"`
}

type TableViewOrder struct {
	Id        int `json:"id" db:"id"`
	ViewOrder int `json:"viewOrder" db:"view_order"`
}

type TableViewLayout struct {
	Id        int            `json:"id" db:"id"`
	View      types.JSONText `json:"view" db:"view"`
	Page      string         `json:"page" db:"page"`
	ViewOrder int            `json:"viewOrder" db:"view_order"`
	Version   string         `json:"version" db:"version"`
}

type TableViewResponses struct {
	Forwards []TableViewLayout `json:"forwards"`
	Channel  []TableViewLayout `json:"channel"`
	Payments []TableViewLayout `json:"payments"`
	Invoices []TableViewLayout `json:"invoices"`
	OnChain  []TableViewLayout `json:"onChain"`
}

type TableViewStructured struct {
	TableViewId int                `json:"tableViewId"`
	Page        string             `json:"page"`
	Title       string             `json:"title"`
	Order       int                `json:"order"`
	UpdateOn    time.Time          `json:"updatedOn"`
	Columns     []TableViewColumn  `json:"columns"`
	Filters     []TableViewFilter  `json:"filters"`
	Sortings    []TableViewSorting `json:"sortings"`
}

type TableView struct {
	TableViewId int       `json:"tableViewId" db:"table_view_id"`
	Page        string    `json:"page" db:"page"`
	Title       string    `json:"title" db:"title"`
	Order       int       `json:"order" db:"order"`
	CreatedOn   time.Time `json:"createdOn" db:"created_on"`
	UpdateOn    time.Time `json:"updatedOn" db:"updated_on"`
}

type TableViewColumn struct {
	TableViewColumnId int       `json:"TableViewColumnId" db:"table_view_column_id"`
	Key               string    `json:"key" db:"key"`
	KeySecond         *string   `json:"keySecond" db:"key_second"`
	Order             int       `json:"order" db:"order"`
	Type              string    `json:"type" db:"type"`
	TableViewId       int       `json:"tableViewId" db:"table_view_id"`
	CreatedOn         time.Time `json:"createdOn" db:"created_on"`
	UpdateOn          time.Time `json:"updatedOn" db:"updated_on"`
}

type TableViewFilter struct {
	TableViewFilterId int            `json:"TableViewFilterId" db:"table_view_filter_id"`
	Filter            types.JSONText `json:"filter" db:"filter"`
	TableViewId       int            `json:"tableViewId" db:"table_view_id"`
	CreatedOn         time.Time      `json:"createdOn" db:"created_on"`
	UpdateOn          time.Time      `json:"updatedOn" db:"updated_on"`
}

type TableViewSorting struct {
	TableViewSortingId int       `json:"TableViewSortingId" db:"table_view_sorting_id"`
	Key                string    `json:"key" db:"key"`
	Order              int       `json:"order" db:"order"`
	Ascending          bool      `json:"ascending" db:"ascending"`
	TableViewId        int       `json:"tableViewId" db:"table_view_id"`
	CreatedOn          time.Time `json:"createdOn" db:"created_on"`
	UpdateOn           time.Time `json:"updatedOn" db:"updated_on"`
}

type tableViewColumnDefinition struct {
	key           string
	locked        bool
	sortable      bool
	heading       string
	visualType    string
	keySecond     string
	valueType     string
	suffix        string
	selectOptions []tableViewSelectOptions
	pages         []TableViewPage
}

type tableViewSelectOptions struct {
	label string
	value string
}

func getTableViewColumnDefinition(key string) tableViewColumnDefinition {
	return getTableViewColumnDefinitions()[key]
}

//go:generate go run ../../cmd/torq/gen.go
func GetTableViewColumnDefinitionsForPage(page TableViewPage) string {
	result := "import { channel } from \"features/channels/channelsTypes\";\nimport { ColumnMetaData } from \"features/table/types\";\n\nexport const AllChannelsColumns: ColumnMetaData<channel>[] = ["
	for _, definition := range getTableViewColumnDefinitions() {
		if slices.Contains(definition.pages, page) {
			result = result + fmt.Sprintf(
				"\n\t{\n\t\theading: \"%v\",\n\t\ttype: \"%v\",\n\t\tkey: \"%v\",\n\t\tvalueType: \"%v\",",
				definition.heading, definition.visualType, definition.key, definition.valueType)
			if definition.locked {
				result = result + "\n\t\tlocked: true,"
			}
			result = result + "\n\t},"
		}
	}
	result = result + "\n];"

	result = result + "\nexport const SortableColumns: Array<keyof channel> = ["
	for _, definition := range getTableViewColumnDefinitions() {
		if slices.Contains(definition.pages, page) {
			if definition.sortable {
				result = result + fmt.Sprintf("\n\t\"%v\",", definition.key)
			}
		}
	}
	result = result + "\n];"

	return result
}

func getTableViewColumnDefinitions() map[string]tableViewColumnDefinition {
	return map[string]tableViewColumnDefinition{
		"peerAlias": {
			key:        "peerAlias",
			locked:     true,
			sortable:   true,
			heading:    "Peer Alias",
			visualType: "AliasCell",
			valueType:  "string",
			pages: []TableViewPage{
				PageChannel,
			},
		},
		"active": {
			key:        "active",
			sortable:   true,
			heading:    "Active",
			visualType: "BooleanCell",
			valueType:  "boolean",
			pages: []TableViewPage{
				PageChannel,
			},
		},
		"balance": {
			key:        "balance",
			heading:    "Balance",
			visualType: "BalanceCell",
			valueType:  "number",
			pages: []TableViewPage{
				PageChannel,
			},
		},
		"tags": {
			key:        "tags",
			heading:    "Tags",
			visualType: "TagsCell",
			valueType:  "tag",
			pages: []TableViewPage{
				PageChannel,
			},
		},
		"shortChannelId": {
			key:        "shortChannelId",
			sortable:   true,
			heading:    "Short Channel ID",
			visualType: "LongTextCell",
			valueType:  "string",
			pages: []TableViewPage{
				PageChannel,
			},
		},
		"gauge": {
			key:        "gauge",
			sortable:   true,
			heading:    "Channel Balance (%)",
			visualType: "BarCell",
			valueType:  "number",
			suffix:     "%",
			pages: []TableViewPage{
				PageChannel,
			},
		},
		"remoteBalance": {
			key:        "remoteBalance",
			sortable:   true,
			heading:    "Remote Balance",
			visualType: "NumericCell",
			valueType:  "number",
			pages: []TableViewPage{
				PageChannel,
			},
		},
		"localBalance": {
			key:        "localBalance",
			sortable:   true,
			heading:    "Local Balance",
			visualType: "NumericCell",
			valueType:  "number",
			pages: []TableViewPage{
				PageChannel,
			},
		},
		"capacity": {
			key:        "capacity",
			sortable:   true,
			heading:    "Capacity",
			visualType: "NumericCell",
			valueType:  "number",
			pages: []TableViewPage{
				PageChannel,
			},
		},
		"feeRateMilliMsat": {
			key:        "feeRateMilliMsat",
			sortable:   true,
			heading:    "Fee rate (PPM)",
			visualType: "NumericDoubleCell",
			keySecond:  "remoteFeeRateMilliMsat",
			valueType:  "number",
			suffix:     "ppm",
			pages: []TableViewPage{
				PageChannel,
			},
		},
		"feeBase": {
			key:        "feeBase",
			sortable:   true,
			heading:    "Base Fee",
			visualType: "NumericDoubleCell",
			keySecond:  "remoteFeeBase",
			valueType:  "number",
			suffix:     "sat",
			pages: []TableViewPage{
				PageChannel,
			},
		},
		"remoteFeeRateMilliMsat": {
			key:        "remoteFeeRateMilliMsat",
			sortable:   true,
			heading:    "Remote Fee rate (PPM)",
			visualType: "NumericCell",
			valueType:  "number",
			suffix:     "ppm",
			pages: []TableViewPage{
				PageChannel,
			},
		},
		"remoteFeeBase": {
			key:        "remoteFeeBase",
			sortable:   true,
			heading:    "Remote Base Fee",
			visualType: "NumericCell",
			valueType:  "number",
			suffix:     "sat",
			pages: []TableViewPage{
				PageChannel,
			},
		},
		"minHtlc": {
			key:        "minHtlc",
			sortable:   true,
			heading:    "Minimum HTLC",
			visualType: "NumericDoubleCell",
			keySecond:  "remoteMinHtlc",
			valueType:  "number",
			suffix:     "sat",
			pages: []TableViewPage{
				PageChannel,
			},
		},
		"maxHtlc": {
			key:        "maxHtlc",
			sortable:   true,
			heading:    "Maximum HTLC",
			visualType: "NumericDoubleCell",
			keySecond:  "remoteMaxHtlc",
			valueType:  "number",
			suffix:     "sat",
			pages: []TableViewPage{
				PageChannel,
			},
		},
		"remoteMinHtlc": {
			key:        "remoteMinHtlc",
			sortable:   true,
			heading:    "Remote Minimum HTLC",
			visualType: "NumericCell",
			valueType:  "number",
			suffix:     "sat",
			pages: []TableViewPage{
				PageChannel,
			},
		},
		"remoteMaxHtlc": {
			key:        "remoteMaxHtlc",
			sortable:   true,
			heading:    "Remote Maximum HTLC",
			visualType: "NumericCell",
			valueType:  "number",
			suffix:     "sat",
			pages: []TableViewPage{
				PageChannel,
			},
		},
		"timeLockDelta": {
			key:        "timeLockDelta",
			heading:    "Time Lock Delta",
			visualType: "NumericDoubleCell",
			keySecond:  "remoteTimeLockDelta",
			valueType:  "number",
			pages: []TableViewPage{
				PageChannel,
			},
		},
		"remoteTimeLockDelta": {
			key:        "remoteTimeLockDelta",
			heading:    "Remote Time Lock Delta",
			visualType: "NumericCell",
			valueType:  "number",
			pages: []TableViewPage{
				PageChannel,
			},
		},
		"lndShortChannelId": {
			key:        "lndShortChannelId",
			heading:    "LND Short Channel ID",
			visualType: "LongTextCell",
			valueType:  "string",
			pages: []TableViewPage{
				PageChannel,
			},
		},
		"channelPoint": {
			key:        "channelPoint",
			heading:    "Channel Point",
			visualType: "LongTextCell",
			valueType:  "string",
			pages: []TableViewPage{
				PageChannel,
			},
		},
		"currentBlockHeight": {
			key:        "currentBlockHeight",
			sortable:   true,
			heading:    "Current BlockHeight",
			visualType: "NumericCell",
			valueType:  "number",
			pages: []TableViewPage{
				PageChannel,
			},
		},
		"fundingTransactionHash": {
			key:        "fundingTransactionHash",
			heading:    "Funding Transaction",
			visualType: "LongTextCell",
			valueType:  "string",
			pages: []TableViewPage{
				PageChannel,
			},
		},
		"fundingBlockHeight": {
			key:        "fundingBlockHeight",
			sortable:   true,
			heading:    "Funding BlockHeight",
			visualType: "NumericCell",
			valueType:  "number",
			pages: []TableViewPage{
				PageChannel,
			},
		},
		"fundingBlockHeightDelta": {
			key:        "fundingBlockHeightDelta",
			sortable:   true,
			heading:    "Funding BlockHeight Delta",
			visualType: "NumericCell",
			valueType:  "number",
			pages: []TableViewPage{
				PageChannel,
			},
		},
		"fundedOn": {
			key:        "fundedOn",
			heading:    "Funding Date",
			visualType: "DateCell",
			valueType:  "date",
			pages: []TableViewPage{
				PageChannel,
			},
		},
		"fundedOnSecondsDelta": {
			key:        "fundedOnSecondsDelta",
			sortable:   true,
			heading:    "Funding Date Delta (Seconds)",
			visualType: "DurationCell",
			valueType:  "duration",
			pages: []TableViewPage{
				PageChannel,
			},
		},
		"closingTransactionHash": {
			key:        "closingTransactionHash",
			heading:    "Closing Transaction",
			visualType: "LongTextCell",
			valueType:  "string",
			pages: []TableViewPage{
				PageChannel,
			},
		},
		"closingBlockHeight": {
			key:        "closingBlockHeight",
			sortable:   true,
			heading:    "Closing BlockHeight",
			visualType: "NumericCell",
			valueType:  "number",
			pages: []TableViewPage{
				PageChannel,
			},
		},
		"closingBlockHeightDelta": {
			key:        "closingBlockHeightDelta",
			sortable:   true,
			heading:    "Closing BlockHeight Delta",
			visualType: "NumericCell",
			valueType:  "number",
			pages: []TableViewPage{
				PageChannel,
			},
		},
		"closedOn": {
			key:        "closedOn",
			heading:    "Closing Date",
			visualType: "DateCell",
			valueType:  "date",
			pages: []TableViewPage{
				PageChannel,
			},
		},
		"closedOnSecondsDelta": {
			key:        "closedOnSecondsDelta",
			sortable:   true,
			heading:    "Closing Date Delta (Seconds)",
			visualType: "DurationCell",
			valueType:  "duration",
			pages: []TableViewPage{
				PageChannel,
			},
		},
		"unsettledBalance": {
			key:        "unsettledBalance",
			sortable:   true,
			heading:    "Unsettled Balance",
			visualType: "NumericCell",
			valueType:  "number",
			pages: []TableViewPage{
				PageChannel,
			},
		},
		"totalSatoshisSent": {
			key:        "totalSatoshisSent",
			sortable:   true,
			heading:    "Satoshis Sent",
			visualType: "NumericCell",
			valueType:  "number",
			pages: []TableViewPage{
				PageChannel,
			},
		},
		"totalSatoshisReceived": {
			key:        "totalSatoshisReceived",
			sortable:   true,
			heading:    "Satoshis Received",
			visualType: "NumericCell",
			valueType:  "number",
			pages: []TableViewPage{
				PageChannel,
			},
		},
		"pendingForwardingHTLCsCount": {
			key:        "pendingForwardingHTLCsCount",
			heading:    "Pending Forwarding HTLCs count",
			visualType: "NumericCell",
			valueType:  "number",
			pages: []TableViewPage{
				PageChannel,
			},
		},
		"pendingForwardingHTLCsAmount": {
			key:        "pendingForwardingHTLCsAmount",
			heading:    "Pending Forwarding HTLCs",
			visualType: "NumericCell",
			valueType:  "number",
			pages: []TableViewPage{
				PageChannel,
			},
		},
		"pendingLocalHTLCsCount": {
			key:        "pendingLocalHTLCsCount",
			heading:    "Pending Forwarding HTLCs count",
			visualType: "NumericCell",
			valueType:  "number",
			pages: []TableViewPage{
				PageChannel,
			},
		},
		"pendingLocalHTLCsAmount": {
			key:        "pendingLocalHTLCsAmount",
			heading:    "Pending Forwarding HTLCs",
			visualType: "NumericCell",
			valueType:  "number",
			pages: []TableViewPage{
				PageChannel,
			},
		},
		"pendingTotalHTLCsCount": {
			key:        "pendingTotalHTLCsCount",
			heading:    "Total Pending Forwarding HTLCs count",
			visualType: "NumericCell",
			valueType:  "number",
			pages: []TableViewPage{
				PageChannel,
			},
		},
		"pendingTotalHTLCsAmount": {
			key:        "pendingTotalHTLCsAmount",
			heading:    "Total Pending Forwarding HTLCs",
			visualType: "NumericCell",
			valueType:  "number",
			pages: []TableViewPage{
				PageChannel,
			},
		},
		"localChanReserveSat": {
			key:        "localChanReserveSat",
			sortable:   true,
			heading:    "Local Channel Reserve",
			visualType: "NumericDoubleCell",
			keySecond:  "remoteChanReserveSat",
			valueType:  "number",
			suffix:     "sat",
			pages: []TableViewPage{
				PageChannel,
			},
		},
		"remoteChanReserveSat": {
			key:        "remoteChanReserveSat",
			sortable:   true,
			heading:    "Remote Channel Reserve",
			visualType: "NumericCell",
			valueType:  "number",
			suffix:     "sat",
			pages: []TableViewPage{
				PageChannel,
			},
		},
		"commitFee": {
			key:        "commitFee",
			sortable:   true,
			heading:    "Commit Fee",
			visualType: "NumericCell",
			valueType:  "number",
			pages: []TableViewPage{
				PageChannel,
			},
		},
		"nodeName": {
			key:        "nodeName",
			sortable:   true,
			heading:    "Node Name",
			visualType: "AliasCell",
			valueType:  "string",
			pages: []TableViewPage{
				PageChannel,
			},
		},
		"mempoolSpace": {
			key:        "mempoolSpace",
			heading:    "Mempool",
			visualType: "LinkCell",
			valueType:  "link",
			pages: []TableViewPage{
				PageChannel,
			},
		},
		"ambossSpace": {
			key:        "ambossSpace",
			heading:    "Amboss",
			visualType: "LinkCell",
			valueType:  "link",
			pages: []TableViewPage{
				PageChannel,
			},
		},
		"oneMl": {
			key:        "oneMl",
			heading:    "1ML",
			visualType: "LinkCell",
			valueType:  "link",
			pages: []TableViewPage{
				PageChannel,
			},
		},
		"date": {
			key:       "date",
			heading:   "Date",
			valueType: "date",
			pages: []TableViewPage{
				PageOnChain,
			},
		},
		"amount": {
			key:       "amount",
			heading:   "Amount",
			valueType: "number",
			pages: []TableViewPage{
				PageOnChain,
			},
		},
		"totalFees": {
			key:       "totalFees",
			heading:   "Fees",
			valueType: "number",
			pages: []TableViewPage{
				PageOnChain,
			},
		},
		"txHash": {
			key:       "txHash",
			heading:   "Tx Hash",
			valueType: "string",
			pages: []TableViewPage{
				PageOnChain,
			},
		},
		"lndShortChanId": {
			key:       "lndShortChanId",
			heading:   "LND Short Channel ID",
			valueType: "string",
			pages: []TableViewPage{
				PageOnChain,
			},
		},
		"lndTxTypeLabel": {
			key:       "lndTxTypeLabel",
			heading:   "LND Tx type label",
			valueType: "string",
			pages: []TableViewPage{
				PageOnChain,
			},
		},
		"destAddressesCount": {
			key:       "destAddressesCount",
			heading:   "Destination Addresses Count",
			valueType: "number",
			pages: []TableViewPage{
				PageOnChain,
			},
		},
		"label": {
			key:       "label",
			heading:   "Label",
			valueType: "string",
			pages: []TableViewPage{
				PageOnChain,
			},
		},
		"alias": {
			key:       "alias",
			locked:    true,
			heading:   "Name",
			valueType: "string",
			pages: []TableViewPage{
				PageForwards,
			},
		},
		"revenueOut": {
			key:       "revenueOut",
			heading:   "Revenue",
			valueType: "number",
			pages: []TableViewPage{
				PageForwards,
			},
		},
		"countTotal": {
			key:       "countTotal",
			heading:   "Total Forwards",
			valueType: "number",
			pages: []TableViewPage{
				PageForwards,
			},
		},
		"amountOut": {
			key:       "amountOut",
			heading:   "Outbound Amount",
			valueType: "number",
			pages: []TableViewPage{
				PageForwards,
			},
		},
		"amountIn": {
			key:       "amountIn",
			heading:   "Inbound Amount",
			valueType: "number",
			pages: []TableViewPage{
				PageForwards,
			},
		},
		"amountTotal": {
			key:       "amountTotal",
			heading:   "Total Amount",
			valueType: "number",
			pages: []TableViewPage{
				PageForwards,
			},
		},
		"turnoverOut": {
			key:       "turnoverOut",
			heading:   "Turnover Outbound",
			valueType: "number",
			pages: []TableViewPage{
				PageForwards,
			},
		},
		"turnoverIn": {
			key:       "turnoverIn",
			heading:   "Turnover Inbound",
			valueType: "number",
			pages: []TableViewPage{
				PageForwards,
			},
		},
		"turnoverTotal": {
			key:       "turnoverTotal",
			heading:   "Total Turnover",
			valueType: "number",
			pages: []TableViewPage{
				PageForwards,
			},
		},
		"countOut": {
			key:       "countOut",
			heading:   "Outbound Forwards",
			valueType: "number",
			pages: []TableViewPage{
				PageForwards,
			},
		},
		"countIn": {
			key:       "countIn",
			heading:   "Inbound Forwards",
			valueType: "number",
			pages: []TableViewPage{
				PageForwards,
			},
		},
		"revenueIn": {
			key:       "revenueIn",
			heading:   "Revenue inbound",
			valueType: "number",
			pages: []TableViewPage{
				PageForwards,
			},
		},
		"revenueTotal": {
			key:       "revenueTotal",
			heading:   "Revenue total",
			valueType: "number",
			pages: []TableViewPage{
				PageForwards,
			},
		},
		"pubKey": {
			key:       "pubKey",
			heading:   "Public key",
			valueType: "string",
			pages: []TableViewPage{
				PageForwards,
			},
		},
		"fundingOutputIndex": {
			key:       "fundingOutputIndex",
			heading:   "Funding Tx Output Index",
			valueType: "string",
			pages: []TableViewPage{
				PageForwards,
			},
		},
		"open": {
			key:        "open",
			heading:    "Open",
			visualType: "BooleanCell",
			valueType:  "boolean",
			pages: []TableViewPage{
				PageForwards,
			},
		},
		"creationDate": {
			key:       "creationDate",
			heading:   "Creation Date",
			valueType: "date",
			pages: []TableViewPage{
				PageInvoices,
			},
		},
		"settleDate": {
			key:       "settleDate",
			heading:   "Settle Date",
			valueType: "date",
			pages: []TableViewPage{
				PageInvoices,
			},
		},
		"invoiceState": {
			key:       "invoiceState",
			heading:   "Settle Date",
			valueType: "enum",
			selectOptions: []tableViewSelectOptions{
				{label: "Open", value: "OPEN"},
				{label: "Settled", value: "SETTLED"},
				{label: "Canceled", value: "CANCELED"},
			},
			pages: []TableViewPage{
				PageInvoices,
			},
		},
		"amtPaid": {
			key:       "amtPaid",
			heading:   "Paid Amount",
			valueType: "number",
			pages: []TableViewPage{
				PageInvoices,
			},
		},
		"memo": {
			key:       "memo",
			heading:   "Memo",
			valueType: "string",
			pages: []TableViewPage{
				PageInvoices,
			},
		},
		"value": {
			key:       "value",
			heading:   "Value",
			valueType: "number",
			pages: []TableViewPage{
				PageInvoices,
			},
		},
		"isRebalance": {
			key:        "isRebalance",
			heading:    "Rebalance",
			visualType: "BooleanCell",
			valueType:  "boolean",
			pages: []TableViewPage{
				PageInvoices,
			},
		},
		"isKeysend": {
			key:        "isKeysend",
			heading:    "Keysend",
			visualType: "BooleanCell",
			valueType:  "boolean",
			pages: []TableViewPage{
				PageInvoices,
			},
		},
		"destinationPubKey": {
			key:       "destinationPubKey",
			heading:   "Destination",
			valueType: "string",
			pages: []TableViewPage{
				PageInvoices,
			},
		},
		"isAmp": {
			key:        "isAmp",
			heading:    "AMP",
			visualType: "BooleanCell",
			valueType:  "boolean",
			pages: []TableViewPage{
				PageInvoices,
			},
		},
		"fallbackAddr": {
			key:       "fallbackAddr",
			heading:   "Fallback Address",
			valueType: "string",
			pages: []TableViewPage{
				PageInvoices,
			},
		},
		"paymentAddr": {
			key:       "paymentAddr",
			heading:   "Payment Address",
			valueType: "string",
			pages: []TableViewPage{
				PageInvoices,
			},
		},
		"paymentRequest": {
			key:       "paymentRequest",
			heading:   "Payment Request",
			valueType: "string",
			pages: []TableViewPage{
				PageInvoices,
			},
		},
		"private": {
			key:        "private",
			heading:    "Private",
			visualType: "BooleanCell",
			valueType:  "boolean",
			pages: []TableViewPage{
				PageInvoices,
			},
		},
		"rHash": {
			key:       "rHash",
			heading:   "Hash",
			valueType: "string",
			pages: []TableViewPage{
				PageInvoices,
			},
		},
		"rPreimage": {
			key:       "rPreimage",
			heading:   "Preimage",
			valueType: "string",
			pages: []TableViewPage{
				PageInvoices,
			},
		},
		"expiry": {
			key:       "expiry",
			heading:   "Expiry",
			valueType: "number",
			pages: []TableViewPage{
				PageInvoices,
			},
		},
		"cltvExpiry": {
			key:       "cltvExpiry",
			heading:   "CLTV Expiry",
			valueType: "number",
			pages: []TableViewPage{
				PageInvoices,
			},
		},
		"updatedOn": {
			key:       "updatedOn",
			heading:   "Updated On",
			valueType: "date",
			pages: []TableViewPage{
				PageInvoices,
			},
		},
		"status": {
			key:       "status",
			heading:   "Status",
			valueType: "array",
			selectOptions: []tableViewSelectOptions{
				{label: "Succeeded", value: "SUCCEEDED"},
				{label: "In Flight", value: "IN_FLIGHT"},
				{label: "Failed", value: "FAILED"},
			},
			pages: []TableViewPage{
				PagePayments,
			},
		},
		"fee": {
			key:       "fee",
			heading:   "Fee",
			valueType: "number",
			pages: []TableViewPage{
				PagePayments,
			},
		},
		"ppm": {
			key:       "ppm",
			heading:   "PPM",
			valueType: "number",
			pages: []TableViewPage{
				PagePayments,
			},
		},
		"secondsInFlight": {
			key:       "secondsInFlight",
			heading:   "Seconds In Flight",
			valueType: "duration",
			pages: []TableViewPage{
				PagePayments,
			},
		},
		"failureReason": {
			key:       "failureReason",
			heading:   "Failure Reason",
			valueType: "array",
			selectOptions: []tableViewSelectOptions{
				{value: "FAILURE_REASON_NONE", label: "None"},
				{value: "FAILURE_REASON_TIMEOUT", label: "Timeout"},
				{value: "FAILURE_REASON_NO_ROUTE", label: "No Route"},
				{value: "FAILURE_REASON_ERROR", label: "Error"},
				{value: "FAILURE_REASON_INCORRECT_PAYMENT_DETAILS", label: "Incorrect Payment Details"},
				{value: "FAILURE_REASON_INCORRECT_PAYMENT_AMOUNT", label: "Incorrect Payment Amount"},
				{value: "FAILURE_REASON_PAYMENT_HASH_MISMATCH", label: "Payment Hash Mismatch"},
				{value: "FAILURE_REASON_INCORRECT_PAYMENT_REQUEST", label: "Incorrect Payment Request"},
				{value: "FAILURE_REASON_UNKNOWN", label: "Unknown"},
			},
			pages: []TableViewPage{
				PagePayments,
			},
		},
		"isMpp": {
			key:        "isMpp",
			heading:    "MPP",
			visualType: "BooleanCell",
			valueType:  "boolean",
			pages: []TableViewPage{
				PagePayments,
			},
		},
		"countFailedAttempts": {
			key:       "countFailedAttempts",
			heading:   "Failed Attempts",
			valueType: "number",
			pages: []TableViewPage{
				PagePayments,
			},
		},
		"countSuccessfulAttempts": {
			key:       "countSuccessfulAttempts",
			heading:   "Successful Attempts",
			valueType: "number",
			pages: []TableViewPage{
				PagePayments,
			},
		},
		"paymentHash": {
			key:       "paymentHash",
			heading:   "Payment Hash",
			valueType: "string",
			pages: []TableViewPage{
				PagePayments,
			},
		},
		"paymentPreimage": {
			key:       "paymentPreimage",
			heading:   "Payment Preimage",
			valueType: "string",
			pages: []TableViewPage{
				PagePayments,
			},
		},
		"workflowName": {
			key:       "workflowName",
			heading:   "Name",
			valueType: "string",
			pages: []TableViewPage{
				PageWorkflow,
			},
		},
		"workflowStatus": {
			key:        "workflowStatus",
			heading:    "Active",
			visualType: "BooleanCell",
			valueType:  "boolean",
			pages: []TableViewPage{
				PageWorkflow,
			},
		},
		"latestVersionName": {
			key:       "latestVersionName",
			heading:   "Latest Draft",
			valueType: "string",
			pages: []TableViewPage{
				PageWorkflow,
			},
		},
		"activeVersionName": {
			key:       "activeVersionName",
			heading:   "Active Version",
			valueType: "string",
			pages: []TableViewPage{
				PageWorkflow,
			},
		},
		"name": {
			key:       "name",
			heading:   "Name",
			valueType: "string",
			pages: []TableViewPage{
				PageTag,
			},
		},
		"categoryId": {
			key:       "categoryId",
			heading:   "Category",
			valueType: "string",
			pages: []TableViewPage{
				PageTag,
			},
		},
		"channels": {
			key:       "channels",
			heading:   "Applied to",
			valueType: "string",
			pages: []TableViewPage{
				PageTag,
			},
		},
		"edit": {
			key:       "edit",
			heading:   "Edit",
			valueType: "string",
			pages: []TableViewPage{
				PageTag,
			},
		},
		"delete": {
			key:       "delete",
			heading:   "Delete",
			valueType: "string",
			pages: []TableViewPage{
				PageTag,
			},
		},
	}
}
