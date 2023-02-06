package on_chain_tx

import (
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type Transaction struct {
	Date               time.Time      `json:"date" db:"date"`
	TxHash             string         `json:"txHash" db:"tx_hash"`
	DestAddresses      pq.StringArray `json:"destAddresses" db:"dest_addresses"`
	DestAddressesCount string         `json:"destAddressesCount" db:"dest_addresses_count"`
	AmountMsat         int64          `json:"amount" db:"amount"`
	TotalFeesMsat      int64          `json:"totalFees" db:"total_fees"`
	Label              *string        `json:"label" db:"label"`
	LndTxTypeLabel     *string        `json:"lndTxTypeLabel" db:"lnd_tx_type_label"`
	LndShortChannelId  *string        `json:"lndShortChanId" db:"lnd_short_chan_id"`
}

func getOnChainTxs(db *sqlx.DB, nodeIds []int, filter sq.Sqlizer, order []string, limit uint64, offset uint64) (r []*Transaction,
	total uint64, err error) {

	//language=PostgreSQL
	qb := sq.Select("*").
		PlaceholderFormat(sq.Dollar).
		FromSelect(
			sq.Select(`
			   timestamp as date,
			   tx_hash,
			   --block_hash,
			   --block_height,
			   --raw_tx_hex,
			   dest_addresses,
			   array_length(dest_addresses, 1) as dest_addresses_count,
			   amount,
			   total_fees,
			   label,
			   (regexp_matches(label, '\d{1,}:(openchannel|closechannel|sweep)|$'))[1] as lnd_tx_type_label,
       		   (regexp_matches(label, '\d{1,}:(openchannel|closechannel):shortchanid-(\d{18,18})|$') )[2] as lnd_short_chan_id
			`).
				PlaceholderFormat(sq.Dollar).
				From("tx").
				Where(sq.Eq{"tx.node_id": nodeIds}),
			"subquery").
		Where(filter).
		OrderBy(order...)

	if limit > 0 {
		qb = qb.Limit(limit).Offset(offset)
	}

	// Compile the query
	qs, args, err := qb.ToSql()
	if err != nil {
		return nil, 0, errors.Wrap(err, "SQL compile statement")
	}

	rows, err := db.Queryx(qs, args...)
	if err != nil {
		return nil, 0, errors.Wrap(err, "Run SQL Query")
	}
	defer rows.Close()
	for rows.Next() {
		var tx Transaction
		err = rows.Scan(
			&tx.Date,
			&tx.TxHash,
			&tx.DestAddresses,
			&tx.DestAddressesCount,
			&tx.AmountMsat,
			&tx.TotalFeesMsat,
			&tx.Label,
			&tx.LndTxTypeLabel,
			&tx.LndShortChannelId,
		)

		if err != nil {
			return nil, 0, errors.Wrap(err, "SQL row scan")
		}

		r = append(r, &tx)
	}

	totalQb := sq.Select("count(*) as total").
		PlaceholderFormat(sq.Dollar).
		FromSelect(
			sq.Select(`
			   timestamp as date,
			   tx_hash,
			   --block_hash,
			   --block_height,
			   --raw_tx_hex,
			   dest_addresses,
			   array_length(dest_addresses, 1) as dest_addresses_count,
			   amount,
			   total_fees,
			   label,
			   (regexp_matches(label, '\d{1,}:(openchannel|closechannel|sweep)|$'))[1] as lnd_tx_type_label,
       		   (regexp_matches(label, '\d{1,}:(openchannel|closechannel):shortchanid-(\d{18,18})|$') )[2] as lnd_short_chan_id
			`).
				PlaceholderFormat(sq.Dollar).
				From("tx").
				Where(sq.Eq{"tx.node_id": nodeIds}),
			"subquery").
		Where(filter)

	totalQs, args, err := totalQb.ToSql()
	if err != nil {
		return nil, 0, errors.Wrap(err, "SQL compile query")
	}

	err = db.QueryRowx(totalQs, args...).Scan(&total)
	if err != nil {
		return nil, 0, errors.Wrap(err, "SQL run query")
	}

	return r, total, nil
}
