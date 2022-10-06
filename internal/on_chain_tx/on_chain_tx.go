package on_chain_tx

import (
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/rs/zerolog/log"
)

type Transaction struct {
	Date               time.Time      `json:"date" db:"date"`
	TxHash             string         `json:"transactionHash" db:"tx_hash"`
	DestAddresses      pq.StringArray `json:"destinationAddresses" db:"dest_addresses"`
	DestAddressesCount string         `json:"destinationAddressesCount" db:"dest_addresses_count"`
	AmountMsat         int64          `json:"amount" db:"amount"`
	TotalFeesMsat      int64          `json:"totalFees" db:"total_fees"`
	Label              *string        `json:"label" db:"label"`
	LndTxTypeLabel     *string        `json:"lndTransactionTypeLabel" db:"lnd_tx_type_label"`
	LndShortChannelId  *string        `json:"lndShortChannelId" db:"lnd_short_chan_id"`
	//BlockHash        *string   `json:"block_hash" db:"block_hash"`
	//BlockHeight      uint64    `json:"block_height" db:"block_height"`
	//RawTxHex         string    `json:"raw_tx_hex" db:"raw_tx_hex"`
}

func getOnChainTxs(db *sqlx.DB, filter sq.Sqlizer, order []string, limit uint64, offset uint64) (r []*Transaction,
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
				From("tx"),
			"subquery").
		Where(filter).
		OrderBy(order...)

	if limit > 0 {
		qb = qb.Limit(limit).Offset(offset)
	}

	// Compile the query
	qs, args, err := qb.ToSql()
	if err != nil {
		return nil, 0, err
	}

	// Log for debugging
	log.Debug().Msgf("Query: %s, \n Args: %v", qs, args)

	rows, err := db.Queryx(qs, args...)
	if err != nil {
		return nil, 0, err
	}
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
			return nil, 0, err
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
				From("tx"),
			"subquery").
		Where(filter)

	totalQs, args, err := totalQb.ToSql()
	if err != nil {
		return nil, 0, err
	}

	err = db.QueryRowx(totalQs, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	return r, total, nil
}
