package communications

import (
	"database/sql"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/rs/zerolog/log"

	"github.com/lncapital/torq/internal/database"
)

type Communication struct {
	CommunicationId int `json:"communicationId" db:"communication_id"`
	// CommunicationType bitshifted value use the Add/Has/Remove methods...
	ActivationFlagNodeDetails bool                    `json:"activationFlagNodeDetails" db:"activation_flag_node_details"`
	TargetType                CommunicationTargetType `json:"targetType" db:"target_type"`
	TargetName                string                  `json:"targetName" db:"target_name"`
	TargetText                string                  `json:"targetText" db:"target_text"`
	TargetNumber              int64                   `json:"targetNumber" db:"target_number"`
	NodeId                    int                     `json:"nodeId" db:"node_id"`
	ChannelId                 *int                    `json:"channelId" db:"channel_id"`
	CreatedOn                 time.Time               `json:"createdOn" db:"created_on"`
	UpdatedOn                 time.Time               `json:"updatedOn" db:"updated_on"`
}

func (communication *Communication) AddCommunicationType(communicationType CommunicationType) {
	switch communicationType {
	case NodeDetailsChanged:
		communication.ActivationFlagNodeDetails = true
	}
}
func (communication *Communication) HasCommunicationType(communicationType CommunicationType) bool {
	switch communicationType {
	case NodeDetailsChanged:
		return communication.ActivationFlagNodeDetails
	}
	return false
}
func (communication *Communication) RemoveCommunicationType(communicationType CommunicationType) {
	switch communicationType {
	case NodeDetailsChanged:
		communication.ActivationFlagNodeDetails = false
	}
}

func GetNodeIdsByCommunication(db *sqlx.DB,
	communicationTargetType CommunicationTargetType) ([]int, error) {

	var nodeIds []int
	err := db.Select(&nodeIds, `SELECT DISTINCT node_id FROM communication WHERE target_type=$1;`,
		communicationTargetType)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, errors.Wrap(err, database.SqlExecutionError)
		}
	}
	return nodeIds, nil
}

func GetCommunicationIdsByNodeId(db *sqlx.DB,
	nodeId int,
	communicationTargetType CommunicationTargetType) ([]int, error) {

	var communicationIds []int
	err := db.Select(&communicationIds,
		`SELECT communication_id FROM communication WHERE node_id=$1 AND target_type=$2;`,
		nodeId, communicationTargetType)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, errors.Wrap(err, database.SqlExecutionError)
		}
	}
	return communicationIds, nil
}

func GetCommunicationIdsByCommunicationTargetType(db *sqlx.DB,
	communicationTargetType CommunicationTargetType) ([]int, error) {

	var communicationIds []int
	err := db.Select(&communicationIds, `SELECT communication_id FROM communication WHERE target_type=$1;`,
		communicationTargetType)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, errors.Wrap(err, database.SqlExecutionError)
		}
	}
	return communicationIds, nil
}

func GetCommunicationSettings(db *sqlx.DB, communicationId int) (map[CommunicationType]bool, error) {
	var activationFlagNodeDetails bool
	err := db.Get(&activationFlagNodeDetails, `SELECT activation_flag_node_details FROM communication WHERE communication_id=$1;`,
		communicationId)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, errors.Wrap(err, database.SqlExecutionError)
		}
	}
	result := make(map[CommunicationType]bool)
	for _, ct := range GetCommunicationTypes() {
		switch ct {
		case NodeDetailsChanged:
			result[ct] = activationFlagNodeDetails
		}
	}
	return result, nil
}

func AddCommunication(db *sqlx.DB, communication Communication) (int, error) {
	communication.CreatedOn = time.Now().UTC()
	communication.UpdatedOn = communication.CreatedOn
	err := db.QueryRowx(`INSERT INTO communication
    	(activation_flag_node_details, target_type, target_name, target_text, target_number,
    	 node_id, channel_id, created_on, updated_on)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING communication_id;`,
		communication.ActivationFlagNodeDetails, communication.TargetType, communication.TargetName,
		communication.TargetText, communication.TargetNumber, communication.NodeId, communication.ChannelId,
		communication.CreatedOn, communication.UpdatedOn).Scan(&communication.CommunicationId)
	if err != nil {
		return 0, errors.Wrap(err, database.SqlExecutionError)
	}
	return communication.CommunicationId, nil
}

func RemoveCommunicationByTargetText(db *sqlx.DB,
	nodeId int,
	targetType CommunicationTargetType,
	targetText string) (int64, error) {

	res, err := db.Exec(`DELETE FROM communication WHERE node_id=$1 AND target_type=$2 AND target_text=$3;`,
		nodeId, targetType, targetText)
	if err != nil {
		return 0, errors.Wrap(err, database.SqlExecutionError)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, database.SqlAffectedRowsCheckError)
	}
	return rowsAffected, nil
}

func RemoveCommunicationByTargetNumber(db *sqlx.DB,
	nodeId int,
	targetType CommunicationTargetType,
	targetNumber int64) (int64, error) {

	res, err := db.Exec(`DELETE FROM communication WHERE node_id=$1 AND target_type=$2 AND target_number=$3;`,
		nodeId, targetType, targetNumber)
	if err != nil {
		return 0, errors.Wrap(err, database.SqlExecutionError)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, database.SqlAffectedRowsCheckError)
	}
	return rowsAffected, nil
}

func GetCommunicationsForNodeDetails(db *sqlx.DB, nodeId int,
	communicationTargetTypes ...CommunicationTargetType) ([]Communication, error) {
	var communications []Communication
	err := db.Select(&communications,
		`SELECT * FROM communication WHERE node_id=$1 AND target_type=ANY($2) AND activation_flag_node_details=$3 AND channel_id IS NULL;`,
		nodeId, pq.Array(communicationTargetTypes), true)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, errors.Wrap(err, database.SqlExecutionError)
		}
	}
	return communications, nil
}

func GetCommunicationsByNodeIdAndTargetTypes(db *sqlx.DB, nodeId int,
	communicationTargetTypes ...CommunicationTargetType) ([]Communication, error) {
	var communications []Communication
	err := db.Select(&communications,
		`SELECT * FROM communication WHERE node_id=$1 AND target_type=ANY($2) AND channel_id IS NULL;`,
		nodeId, pq.Array(communicationTargetTypes))
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, errors.Wrap(err, database.SqlExecutionError)
		}
	}
	return communications, nil
}

func GetCommunicationsByChannelIdAndTargetTypes(db *sqlx.DB, nodeId int, channelId int,
	communicationTargetTypes ...CommunicationTargetType) ([]Communication, error) {
	var communications []Communication
	err := db.Select(&communications, `
		SELECT *
		FROM communication
		WHERE channel_id=$1 AND target_type=ANY($2) AND node_id=$3;`,
		channelId, pq.Array(communicationTargetTypes), nodeId)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, errors.Wrap(err, database.SqlExecutionError)
		}
	}
	return communications, nil
}

func SetCommunication(db *sqlx.DB, communication Communication) (Communication, error) {
	updatedOn := time.Now().UTC()
	res, err := db.Exec(`
		UPDATE communication
		SET activation_flag_node_details=$3, target_name=$4, target_type=$5, target_text=$6, target_number=$7,
		    node_id=$8, channel_id=$9, updated_on=$10
		WHERE communication_id=$1 AND updated_on=$2;`,
		communication.CommunicationId, communication.UpdatedOn,
		communication.ActivationFlagNodeDetails, communication.TargetName, communication.TargetType,
		communication.TargetText, communication.TargetNumber,
		communication.NodeId, communication.ChannelId, updatedOn)
	if err != nil {
		return Communication{}, errors.Wrap(err, database.SqlExecutionError)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		log.Error().Err(err).Msgf(database.SqlAffectedRowsCheckError+" %v", err)
	}
	if rowsAffected != 1 {
		return Communication{}, errors.Wrap(err, database.SqlUpdateOneExecutionError)
	}
	communication.UpdatedOn = updatedOn
	return communication, nil
}

func GetCommunicationIdsByTargetNumber(db *sqlx.DB, chatId int64) ([]int, error) {
	var communicationIds []int
	err := db.Select(&communicationIds, `SELECT communication_id FROM communication WHERE target_number=$1;`, chatId)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, errors.Wrap(err, database.SqlExecutionError)
		}
	}
	return communicationIds, nil

}

func GetCommunicationIdsByTargetText(db *sqlx.DB, targetText string) ([]int, error) {
	var communicationIds []int
	err := db.Select(&communicationIds, `SELECT communication_id FROM communication WHERE target_text=$1;`,
		targetText)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, errors.Wrap(err, database.SqlExecutionError)
		}
	}
	return communicationIds, nil

}
