package workflows

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/pkg/errors"

	"github.com/lncapital/torq/internal/core"
	"github.com/lncapital/torq/internal/database"
)

//func GetWorkflowByWorkflowVersionId(db *sqlx.DB, workflowVersionId int) (Workflow, error) {
//	var wf Workflow
//	err := db.Get(&wf, `
//		SELECT wf.*
//		FROM workflow_version wfv
//		JOIN workflow wf ON wf.workflow_id=wfv.workflow_id
//		WHERE wfv.workflow_version_id=$1;`, workflowVersionId)
//	if err != nil {
//		if errors.Is(err, sql.ErrNoRows) {
//			return Workflow{}, nil
//		}
//		return Workflow{}, errors.Wrap(err, database.SqlExecutionError)
//	}
//	return wf, nil
//}

//func GetWorkflow(db *sqlx.DB, workflowId int) (WorkflowTableRow, error) {
//	var wfs WorkflowTableRow
//	err := db.Select(&wfs, `SELECT
//			w.workflow_id, w.name as workflow_name, w.status as workflow_status,
//			wv.name as latest_version_name, wv.version as latest_version, wv.workflow_version_id as latest_workflow_version_id, wv.status as latest_version_status,
//			awv.name as active_version_name, awv.version as active_version, awv.workflow_version_id as active_workflow_version_id, awv.status as active_version_status
//		FROM workflow as w
//		left join (
//			select * from (select *,
//							ROW_NUMBER() OVER (PARTITION BY workflow_id ORDER BY workflow_version_id DESC) as rank
//							from workflow_version order by rank) as wv where rank = 1) as wv on w.workflow_id = wv.workflow_id
//		left join (select * from (select *,
//								 	ROW_NUMBER() OVER (PARTITION BY workflow_id ORDER BY workflow_version_id DESC) as rank
//						    from workflow_version where status = 1 order by rank) as wv where rank = 1) as awv on w.workflow_id = awv.workflow_id
//		WHERE w.workflow_id=$1;`, workflowId)
//	if err != nil {
//		if errors.Is(err, sql.ErrNoRows) {
//			return WorkflowTableRow{}, nil
//		}
//		return WorkflowTableRow{}, errors.Wrap(err, database.SqlExecutionError)
//	}
//	return wfs, nil
//}

func GetWorkflow(db *sqlx.DB, workflowId int) (Workflow, error) {
	var wf Workflow
	err := db.Get(&wf, `SELECT * FROM workflow WHERE workflow_id=$1;`, workflowId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Workflow{}, nil
		}
		return Workflow{}, errors.Wrap(err, database.SqlExecutionError)
	}
	return wf, nil
}

func GetWorkflows(db *sqlx.DB) ([]WorkflowTableRow, error) {
	var wfs []WorkflowTableRow
	err := db.Select(&wfs, `SELECT
			w.workflow_id, w.name as workflow_name, w.status as workflow_status,
			wv.name as latest_version_name, wv.version as latest_version, wv.workflow_version_id as latest_workflow_version_id, wv.status as latest_version_status,
			awv.name as active_version_name, awv.version as active_version, awv.workflow_version_id as active_workflow_version_id, awv.status as active_version_status
		FROM workflow as w
		left join (
			select * from (select *,
				ROW_NUMBER() OVER (PARTITION BY workflow_id ORDER BY workflow_version_id DESC) as rank
				from workflow_version order by rank) as wv where rank = 1) as wv on w.workflow_id = wv.workflow_id
		left join (
			select * from (select *,
				ROW_NUMBER() OVER (PARTITION BY workflow_id ORDER BY workflow_version_id DESC) as rank
				from workflow_version where status = 1 order by rank) as wv where rank = 1) as awv on w.workflow_id = awv.workflow_id
		WHERE w.status != $1
		ORDER BY w.name;`, core.Archived)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []WorkflowTableRow{}, nil
		}
		return []WorkflowTableRow{}, errors.Wrap(err, database.SqlExecutionError)
	}
	return wfs, nil
}

func createWorkflow(db *sqlx.DB) (Workflow, error) {
	createdTime := time.Now().UTC()
	workflow := Workflow{
		Name:      "New Workflow",
		CreatedOn: createdTime,
		UpdateOn:  createdTime,
	}

	err := db.QueryRowx(`
			INSERT INTO workflow (name, status, created_on, updated_on)
			VALUES(
				(SELECT(
					SELECT
						CASE WHEN count(*) = 0 THEN $1
							 ELSE $1 || ' ' || coalesce((
								 SELECT max(coalesce(regexp_replace(name, ($1 || ' (\d+)'), '\1'), '0')::integer) + 1
								 FROM workflow
								 WHERE name ~* ($1 || ' (\d+)')
							 ), '1')
						END
					FROM workflow
					WHERE name = $1)
			    ),
				$2,
				$3,
				$4)
			RETURNING workflow_id;`,
		workflow.Name, workflow.Status, workflow.CreatedOn, workflow.UpdateOn).Scan(&workflow.WorkflowId)

	if err != nil {
		if err, ok := err.(*pq.Error); ok {
			if err.Code == "23505" {
				return Workflow{}, database.SqlUniqueConstraintError
			}
		}
		return Workflow{}, errors.Wrap(err, database.SqlExecutionError)
	}
	return workflow, nil
}

func updateWorkflow(db *sqlx.DB, req UpdateWorkflow) (UpdateWorkflow, error) {
	if req.Status != nil && *req.Status == Archived {
		_, err := db.Exec(`UPDATE workflow SET status=$1, name=name||'_ARCHIVED_'||$4, updated_on=$2 WHERE workflow_id=$3;`,
			req.Status, time.Now().UTC(), req.WorkflowId, time.Now().UnixMilli())
		if err != nil {
			return UpdateWorkflow{}, errors.Wrap(err, database.SqlExecutionError)
		}
	} else {
		qb := sq.Update("workflow").PlaceholderFormat(sq.Dollar).Set("updated_on", time.Now().UTC())
		if req.Name != nil {
			qb = qb.Set("name", req.Name)
		}
		if req.Status != nil {
			qb = qb.Set("status", req.Status)
		}
		_, err := qb.Where(sq.Eq{"workflow_id": req.WorkflowId}).RunWith(db).Exec()
		if err != nil {
			if err, ok := err.(*pq.Error); ok {
				if err.Code == "23505" {
					return UpdateWorkflow{}, database.SqlUniqueConstraintError
				}
			}
			return UpdateWorkflow{}, errors.Wrap(err, database.SqlExecutionError)
		}
	}
	return req, nil
}

func removeWorkflow(db *sqlx.DB, workflowId int) (int64, error) {
	res, err := db.Exec(`DELETE FROM workflow WHERE workflow_id = $1;`, workflowId)
	if err != nil {
		return 0, errors.Wrap(err, database.SqlExecutionError)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, database.SqlAffectedRowsCheckError)
	}
	return rowsAffected, nil
}

func GetWorkflowVersion(db *sqlx.DB, workflowId int, versionId int) (WorkflowVersion, error) {
	var wfv WorkflowVersion
	err := db.Get(&wfv, `SELECT * FROM workflow_version WHERE workflow_id = $1 and version = $2  limit 1;`, workflowId, versionId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return WorkflowVersion{}, nil
		}
		return WorkflowVersion{}, errors.Wrap(err, database.SqlExecutionError)
	}
	return wfv, nil
}

func GetWorkflowVersionById(db *sqlx.DB, workflowVersionId int) (WorkflowVersion, error) {
	var wfv WorkflowVersion
	err := db.Get(&wfv, `SELECT * FROM workflow_version WHERE workflow_version_id = $1 limit 1;`, workflowVersionId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return WorkflowVersion{}, nil
		}
		return WorkflowVersion{}, errors.Wrap(err, database.SqlExecutionError)
	}
	return wfv, nil
}

func GetWorkflowVersions(db *sqlx.DB, workflowId int) ([]WorkflowVersion, error) {
	var wfvs []WorkflowVersion
	err := db.Select(&wfvs, `SELECT * FROM workflow_version WHERE workflow_id=$1;`, workflowId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []WorkflowVersion{}, nil
		}
		return []WorkflowVersion{}, errors.Wrap(err, database.SqlExecutionError)
	}
	return wfvs, nil
}

func createWorkflowVersion(db *sqlx.DB, workflowId int, status WorkflowStatus) (WorkflowVersion, error) {
	wfv := WorkflowVersion{}
	wfv.WorkflowId = workflowId
	wfv.Name = "Initial Version"
	wfv.Status = status
	wfv.Version = 1
	wfv.CreatedOn = time.Now().UTC()
	wfv.UpdateOn = wfv.CreatedOn

	err := db.QueryRowx(`INSERT INTO workflow_version (name, version, status, workflow_id, created_on, updated_on)
			VALUES ($1, $2, $3, $4, $5, $6) RETURNING workflow_version_id;`,
		wfv.Name,
		wfv.Version,
		wfv.Status,
		wfv.WorkflowId,
		wfv.CreatedOn,
		wfv.UpdateOn).
		Scan(&wfv.WorkflowVersionId)
	if err != nil {
		return WorkflowVersion{}, errors.Wrap(err, database.SqlExecutionError)
	}
	return wfv, nil
}

func cloneWorkflowVersion(db *sqlx.DB, workflowId int, cloneVersionId *int) (WorkflowVersion, error) {

	var newVersion int
	err := db.Get(&newVersion, `SELECT coalesce(max(version),0)+1 FROM workflow_version WHERE workflow_id=$1;`, workflowId)
	if err != nil {
		return WorkflowVersion{}, errors.Wrap(err, database.SqlExecutionError)
	}

	//Clone latest when no cloneVersionId is passed in
	var wfv WorkflowVersion
	if cloneVersionId == nil {
		err = db.Get(&wfv, `SELECT * FROM workflow_version WHERE workflow_id=$1 ORDER BY version DESC LIMIT 1;`, workflowId)
	} else {
		err = db.Get(&wfv, `SELECT * FROM workflow_version WHERE workflow_id=$1 AND version=$2;`, workflowId, *cloneVersionId)
	}

	if err != nil {
		return WorkflowVersion{}, errors.Wrap(err, database.SqlExecutionError)
	}

	// Get all the nodes belonging to the version to be cloned
	var wfvns []WorkflowVersionNode
	err = db.Select(&wfvns, `SELECT * FROM workflow_version_node WHERE workflow_version_id=$1;`, wfv.WorkflowVersionId)
	if err != nil {
		return WorkflowVersion{}, errors.Wrap(err, database.SqlExecutionError)
	}

	// Get all the links (edges) belonging to the version to be cloned
	var wfvnls []WorkflowVersionNodeLink
	err = db.Select(&wfvnls, `SELECT * FROM workflow_version_node_link WHERE workflow_version_id=$1;`, wfv.WorkflowVersionId)
	if err != nil {
		return WorkflowVersion{}, errors.Wrap(err, database.SqlExecutionError)
	}

	// Start a new transaction so that we can roll back if something goes wrong
	tx := db.MustBegin()
	if err != nil {
		return WorkflowVersion{}, errors.Wrap(err, "adding workflow_version transaction failed to initialise")
	}

	// newWfv is the new workflow version that will be created with all the cloned data
	newWfv := WorkflowVersion{}

	// Populate the new workflow version with the data from the old one
	newWfv.CreatedOn = time.Now().UTC()
	// TODO: Create a better and unique workflow version name or remove the unique constraint and use a default name
	newWfv.Name = fmt.Sprintf("%v", wfv.CreatedOn.Format("20060102.150405.000000"))
	newWfv.UpdateOn = newWfv.CreatedOn
	newWfv.WorkflowId = workflowId
	newWfv.Version = newVersion

	err = tx.QueryRow(`
			INSERT INTO workflow_version (name, version, status, workflow_id, created_on, updated_on)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING workflow_version_id;`,
		newWfv.Name,
		newWfv.Version,
		newWfv.Status,
		newWfv.WorkflowId,
		newWfv.CreatedOn,
		newWfv.UpdateOn).Scan(&newWfv.WorkflowVersionId)

	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return WorkflowVersion{}, errors.Wrap(rollbackErr, "adding workflow_version failed (rollback failed too)")
		}
		return WorkflowVersion{}, errors.Wrap(err, "adding workflow_version failed (rollback done)")
	}

	fromToWorkflowVersionNodeIds := make(map[int]int)

	// Clone the old node and assign them to the new node ids
	for _, workflowVersionNode := range wfvns {
		existingWorkflowVersionNodeId := workflowVersionNode.WorkflowVersionNodeId
		workflowVersionNode.WorkflowVersionId = newWfv.WorkflowVersionId

		err = db.QueryRowx(`
				INSERT INTO workflow_version_node
					(name, status, type, parameters, visibility_settings, workflow_version_id, created_on, updated_on)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING workflow_version_node_id;`,
			workflowVersionNode.Name,
			workflowVersionNode.Status,
			workflowVersionNode.Type,
			workflowVersionNode.Parameters,
			workflowVersionNode.VisibilitySettings,
			workflowVersionNode.WorkflowVersionId,
			workflowVersionNode.CreatedOn,
			workflowVersionNode.UpdateOn).Scan(&workflowVersionNode.WorkflowVersionNodeId)
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				return WorkflowVersion{}, errors.Wrap(rollbackErr, "adding workflow_version_node failed (rollback failed too)")
			}
			return WorkflowVersion{}, errors.Wrap(err, "adding workflow_version_node failed (rollback done)")
		}
		// If there are no errors map old to new workflow_version_node_id
		fromToWorkflowVersionNodeIds[existingWorkflowVersionNodeId] = workflowVersionNode.WorkflowVersionNodeId
	}

	// wfl = Workflow Version Node Link
	// Clone the old links and assign them to the new node ids
	for _, wvnl := range wfvnls {

		wvnl.WorkflowVersionId = newWfv.WorkflowVersionId
		wvnl.ParentWorkflowVersionNodeId = fromToWorkflowVersionNodeIds[wvnl.ParentWorkflowVersionNodeId]
		wvnl.ChildWorkflowVersionNodeId = fromToWorkflowVersionNodeIds[wvnl.ChildWorkflowVersionNodeId]

		err = db.QueryRowx(`
				INSERT INTO workflow_version_node_link
					(name, visibility_settings, parent_output, parent_workflow_version_node_id,
					 child_input, child_workflow_version_node_id, workflow_version_id, created_on, updated_on)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING workflow_version_node_link_id;`,
			wvnl.Name,
			wvnl.VisibilitySettings,
			wvnl.ParentOutput,
			wvnl.ParentWorkflowVersionNodeId,
			wvnl.ChildInput,
			wvnl.ChildWorkflowVersionNodeId,
			wvnl.WorkflowVersionId,
			wvnl.CreatedOn,
			wvnl.UpdateOn).Scan(&wvnl.WorkflowVersionNodeLinkId)
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				return WorkflowVersion{}, errors.Wrap(rollbackErr, "adding workflow_version_node_link failed (rollback failed too)")
			}
			return WorkflowVersion{}, errors.Wrap(err, "adding workflow_version_node_link failed (rollback done)")
		}
	}

	return wfv, nil
}

func setWorkflowVersion(db *sqlx.DB, workflowVersion WorkflowVersion) (WorkflowVersion, error) {
	workflowVersion.UpdateOn = time.Now().UTC()
	_, err := db.Exec(`UPDATE workflow_version SET name=$1, version=$2, status=$3, updated_on=$4 WHERE workflow_version_id=$5;`,
		workflowVersion.Name, workflowVersion.Version, workflowVersion.Status, workflowVersion.UpdateOn, workflowVersion.WorkflowVersionId)
	if err != nil {
		return WorkflowVersion{}, errors.Wrap(err, database.SqlExecutionError)
	}
	return workflowVersion, nil
}

func removeWorkflowVersion(db *sqlx.DB, workflowVersionId int) (int64, error) {
	res, err := db.Exec(`DELETE FROM workflow_version WHERE workflow_version_id = $1;`, workflowVersionId)
	if err != nil {
		return 0, errors.Wrap(err, database.SqlExecutionError)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, database.SqlAffectedRowsCheckError)
	}
	return rowsAffected, nil
}

func GetWorkflowIdsByNodeType(db *sqlx.DB, nodeType core.WorkflowNodeType) ([]int, error) {
	var workflowIds []int
	err := db.Select(&workflowIds, `
		SELECT DISTINCT wfv.workflow_id
		FROM workflow_version_node wfvn
		JOIN workflow_version wfv ON wfv.workflow_version_id = wfvn.workflow_version_id
		WHERE wfvn.type=$1 AND wfvn.status!=$2;`, nodeType, WorkflowNodeDeleted)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []int{}, nil
		}
		return nil, errors.Wrap(err, database.SqlExecutionError)
	}
	return workflowIds, nil
}

func GetActiveEventTriggerNodes(db *sqlx.DB, nodeType core.WorkflowNodeType) ([]WorkflowNode, error) {
	var workflowVersionRootNodeIds []int
	err := db.Select(&workflowVersionRootNodeIds, `
		SELECT wfvn.workflow_version_node_id
		FROM workflow_version_node wfvn
		JOIN workflow_version wfv ON wfv.workflow_version_id = wfvn.workflow_version_id AND wfv.status=$1
		JOIN workflow wf ON wf.workflow_id = wfv.workflow_id AND wf.status=$1
		LEFT JOIN workflow_version_node_link parentLink ON parentLink.child_workflow_version_node_id = wfvn.workflow_version_node_id
		WHERE wfvn.status=$1 AND wfvn.type=$2 AND parentLink.child_workflow_version_node_id IS NULL AND wfv.workflow_version_id IN (
			SELECT ranked.workflow_version_id
			FROM (
				SELECT v_wfv.workflow_version_id, RANK() OVER (PARTITION BY v_wfv.workflow_version_id ORDER BY version DESC) version_rank
				FROM workflow_version v_wfv
				WHERE v_wfv.status=$1
			) ranked
			WHERE ranked.version_rank = 1
		)
		ORDER BY wf.name;`, core.Active, nodeType)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []WorkflowNode{}, nil
		}
		return nil, errors.Wrap(err, database.SqlExecutionError)
	}
	var response []WorkflowNode
	for _, workflowVersionRootNodeId := range workflowVersionRootNodeIds {
		workflowNode, err := GetWorkflowNode(db, workflowVersionRootNodeId)
		if err != nil {
			return nil, errors.Wrapf(err, "Obtaining WorkflowNode for workflowVersionRootNodeId: %v", workflowVersionRootNodeId)
		}
		response = append(response, workflowNode)
	}
	return response, nil
}

func GetActiveSortedStageTriggerNodeForWorkflowVersionId(db *sqlx.DB, workflowVersionId int) ([]WorkflowNode, error) {
	var workflowVersionRootNodeIds []int
	err := db.Select(&workflowVersionRootNodeIds, `
		SELECT wfvn.workflow_version_node_id
		FROM workflow_version_node wfvn
		JOIN workflow_version wfv ON wfv.workflow_version_id = wfvn.workflow_version_id AND wfv.status=$1
		JOIN workflow wf ON wf.workflow_id = wfv.workflow_id AND wfv.status=$1
		LEFT JOIN workflow_version_node_link parentLink ON parentLink.child_workflow_version_node_id = wfvn.workflow_version_node_id
		WHERE wfvn.status=$1 AND wfvn.type=$2 AND wfv.workflow_version_id=$3 AND parentLink.child_workflow_version_node_id IS NULL
		ORDER BY wfvn.stage;`, core.Active, core.WorkflowNodeStageTrigger, workflowVersionId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []WorkflowNode{}, nil
		}
		return nil, errors.Wrap(err, database.SqlExecutionError)
	}
	var response []WorkflowNode
	for _, workflowVersionStageNodeId := range workflowVersionRootNodeIds {
		workflowNode, err := GetWorkflowNode(db, workflowVersionStageNodeId)
		if err != nil {
			return nil, errors.Wrapf(err, "Obtaining WorkflowNode for workflowVersionStageNodeId: %v", workflowVersionStageNodeId)
		}
		response = append(response, workflowNode)
	}
	return response, nil
}

func GetWorkflowVersionNodesByStage(db *sqlx.DB, workflowVersionId int, stage int) ([]WorkflowNode, error) {
	var wfvnIds []int
	err := db.Select(&wfvnIds, `
		SELECT workflow_version_node_id
		FROM workflow_version_node
		WHERE workflow_version_id=$1 AND stage=$2 AND status!=$3;`, workflowVersionId, stage, WorkflowNodeDeleted)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, errors.Wrap(err, database.SqlExecutionError)
	}
	var results []WorkflowNode
	for _, wfvnId := range wfvnIds {
		wfvn, err := GetWorkflowNode(db, wfvnId)
		if err != nil {
			return nil, errors.Wrap(err, database.SqlExecutionError)
		}
		results = append(results, wfvn)
	}
	return results, nil
}

func getWorkflowVersionNodesByWorkflow(db *sqlx.DB, workflowId int) ([]WorkflowNode, error) {
	wfvnIds, err := getWorkflowVersionNodeIdsByWorkflow(db, workflowId)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not get workflowVersionNodeIds for workflowId: %v", workflowId)
	}
	var results []WorkflowNode
	for _, wfvnId := range wfvnIds {
		wfvn, err := GetWorkflowNode(db, wfvnId)
		if err != nil {
			return nil, errors.Wrap(err, database.SqlExecutionError)
		}
		results = append(results, wfvn)
	}
	return results, nil
}

func getWorkflowVersionNodeIdsByWorkflow(db *sqlx.DB, workflowId int) ([]int, error) {
	var wfvnIds []int
	err := db.Select(&wfvnIds, `
		SELECT wfvn.workflow_version_node_id
		FROM workflow_version_node wfvn
		JOIN workflow_version wfv ON wfv.workflow_version_id=wfvn.workflow_version_id
		WHERE wfv.workflow_id=$1 AND wfvn.status!=$2;`, workflowId, WorkflowNodeDeleted)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, errors.Wrap(err, database.SqlExecutionError)
	}
	return wfvnIds, nil
}

// GetWorkflowNode is not recursive and only returns direct parent/child relations without further nesting.
func GetWorkflowNode(db *sqlx.DB, workflowVersionNodeId int) (WorkflowNode, error) {
	var wfvn WorkflowVersionNode
	err := db.Get(&wfvn, `
		SELECT * FROM workflow_version_node WHERE workflow_version_node_id=$1 AND status!=$2;`,
		workflowVersionNodeId, WorkflowNodeDeleted)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return WorkflowNode{}, nil
		}
		return WorkflowNode{}, errors.Wrap(err, database.SqlExecutionError)
	}

	parentNodes, parentNodeLinkDetails, err := getParentNodes(db, workflowVersionNodeId)
	if err != nil {
		return WorkflowNode{}, errors.Wrapf(err, "Obtaining parent nodes for workflowVersionNodeId: %v", workflowVersionNodeId)
	}

	childNodes, childNodeLinkDetails, err := getChildNodes(db, workflowVersionNodeId)
	if err != nil {
		return WorkflowNode{}, errors.Wrapf(err, "Obtaining child nodes for workflowVersionNodeId: %v", workflowVersionNodeId)
	}

	response := wfvn.GetWorkflowNodeStructured()
	response.ParentNodes = parentNodes
	response.ChildNodes = childNodes
	response.LinkDetails = parentNodeLinkDetails
	if response.LinkDetails == nil {
		response.LinkDetails = childNodeLinkDetails
	} else {
		for childNodeLinkId, childNodeLink := range childNodeLinkDetails {
			response.LinkDetails[childNodeLinkId] = childNodeLink
		}
	}

	return response, nil
}

func GetWorkflowNodes(db *sqlx.DB, workflowVersionId int, workflowId int, version int) ([]WorkflowVersionNode, error) {
	// Query all workflow nodes for the given workflow version
	var workflowVersionNodes []WorkflowVersionNodeResponse
	err := db.Select(&workflowVersionNodes, `
		SELECT *
		FROM workflow_version_node
		WHERE workflow_version_id=$1 AND status!=$2
		ORDER BY created_on DESC;`, workflowVersionId, WorkflowNodeDeleted)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []WorkflowVersionNode{}, nil
		}
		return nil, errors.Wrap(err, database.SqlExecutionError)
	}
	// Loop thorugh and replace WorkflowVersionNodeResponse  with WorkflowVersionNode and unmarcahs the parameters
	var response []WorkflowVersionNode
	for _, wvnr := range workflowVersionNodes {
		var wvn WorkflowVersionNode
		wvn.WorkflowVersionNodeId = wvnr.WorkflowVersionNodeId
		wvn.WorkflowVersionId = wvnr.WorkflowVersionId
		wvn.Type = wvnr.Type
		wvn.Stage = wvnr.Stage
		wvn.Name = wvnr.Name
		wvn.VisibilitySettings = wvnr.VisibilitySettings
		wvn.WorkflowId = workflowId
		wvn.Version = version
		wvn.Status = wvnr.Status

		// Unmarshal parameters
		err = json.Unmarshal(wvnr.Parameters, &wvn.Parameters)
		if err != nil {
			return nil, errors.Wrapf(err, "Unmarshalling parameters for workflowVersionNodeId: %v", wvnr.WorkflowVersionNodeId)
		}

		response = append(response, wvn)
	}

	return response, nil
}

func GetWorkflowForest(db *sqlx.DB, workflowVersionId int) (WorkflowForest, error) {
	var rootVersionNodeIds []int

	err := db.Select(&rootVersionNodeIds, `
		SELECT n.workflow_version_node_id
		FROM workflow_version_node n
		LEFT JOIN workflow_version_node_link parentLink ON
			parentLink.child_workflow_version_node_id = n.workflow_version_node_id
		WHERE n.workflow_version_id=$1 AND n.status!=$2 AND
			parentLink.child_workflow_version_node_id IS NULL AND n.stage IS NULL;`,
		workflowVersionId, WorkflowNodeDeleted)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return WorkflowForest{}, nil
		}
		return WorkflowForest{}, errors.Wrap(err, database.SqlExecutionError)
	}

	processedNodes := make(map[int]*WorkflowNode)

	rootNodesStructured := make(map[int][]*WorkflowNode)
	for _, rootVersionNodeId := range rootVersionNodeIds {
		workflowVersionNode, err := GetWorkflowVersionNode(db, rootVersionNodeId)
		if err != nil {
			return WorkflowForest{}, err
		}
		workflowNode := workflowVersionNode.GetWorkflowNodeStructured()
		err = processNodeRecursion(processedNodes, db, &workflowNode, nil, nil)
		if err != nil {
			return WorkflowForest{}, err
		}
		rootNodesStructured[0] = append(rootNodesStructured[0], &workflowNode)
	}

	var versionStageNodeIds []int
	err = db.Select(&versionStageNodeIds, `
		SELECT n.workflow_version_node_id
		FROM workflow_version_node n
		LEFT JOIN workflow_version_node_link parentLink ON
			parentLink.child_workflow_version_node_id = n.workflow_version_node_id
		WHERE n.workflow_version_id=$1 AND n.status!=$2 AND
			parentLink.child_workflow_version_node_id IS NULL AND n.stage IS NOT NULL
		ORDER BY n.stage;`, workflowVersionId, WorkflowNodeDeleted)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return WorkflowForest{}, nil
		}
		return WorkflowForest{}, errors.Wrap(err, database.SqlExecutionError)
	}

	for _, versionStageNodeId := range versionStageNodeIds {
		workflowVersionNode, err := GetWorkflowVersionNode(db, versionStageNodeId)
		if err != nil {
			return WorkflowForest{}, err
		}
		workflowNode := workflowVersionNode.GetWorkflowNodeStructured()
		err = processNodeRecursion(processedNodes, db, &workflowNode, nil, nil)
		if err != nil {
			return WorkflowForest{}, err
		}
		rootNodesStructured[workflowVersionNode.Stage] = append(rootNodesStructured[workflowVersionNode.Stage], &workflowNode)
	}

	return WorkflowForest{SortedStageTrees: rootNodesStructured}, nil
}

func processNodeRecursion(processedNodes map[int]*WorkflowNode, db *sqlx.DB, workflowNode *WorkflowNode,
	workflowParentNode *WorkflowNode, workflowParentNodeLink *WorkflowVersionNodeLink) error {

	if workflowParentNode != nil && workflowParentNodeLink != nil {
		if workflowNode.ParentNodes == nil {
			workflowNode.ParentNodes = make(map[int]*WorkflowNode)
		}
		if workflowNode.LinkDetails == nil {
			workflowNode.LinkDetails = make(map[int]WorkflowVersionNodeLink)
		}
		workflowNode.ParentNodes[workflowParentNodeLink.WorkflowVersionNodeLinkId] = workflowParentNode
		workflowNode.LinkDetails[workflowParentNodeLink.WorkflowVersionNodeLinkId] = *workflowParentNodeLink
	}

	childNodes, childNodeLinkDetails, err := getChildNodes(db, workflowNode.WorkflowVersionNodeId)
	if err != nil {
		return errors.Wrapf(err, "Obtaining child nodes for workflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
	}
	childNodesStructured := make(map[int]*WorkflowNode)
	for childNodeLinkId, childNode := range childNodes {
		alreadyProcessedNode, exists := processedNodes[childNode.WorkflowVersionNodeId]
		if exists && alreadyProcessedNode != nil && alreadyProcessedNode.WorkflowVersionNodeId != 0 {
			childNodesStructured[childNodeLinkId] = alreadyProcessedNode
		} else {
			childNodesStructured[childNodeLinkId] = childNode
			childNodeLinkDetail := childNodeLinkDetails[childNodeLinkId]
			err = processNodeRecursion(processedNodes, db, childNode, workflowNode, &childNodeLinkDetail)
			if err != nil {
				return errors.Wrapf(err, "Obtaining child nodes recursive for workflowVersionNodeId: %v", childNode.WorkflowVersionNodeId)
			}
		}
	}
	workflowNode.ChildNodes = childNodesStructured
	if workflowNode.LinkDetails == nil {
		workflowNode.LinkDetails = make(map[int]WorkflowVersionNodeLink)
	}
	for childNodeLinkId, childNodeLink := range childNodeLinkDetails {
		workflowNode.LinkDetails[childNodeLinkId] = childNodeLink
	}
	return nil
}

func getParentNodes(db *sqlx.DB, workflowVersionNodeId int) (map[int]*WorkflowNode, map[int]WorkflowVersionNodeLink, error) {
	rows, err := db.Queryx(`
		SELECT n.workflow_version_node_id, n.name, n.status, n.stage, n.type, n.parameters, n.visibility_settings,
		       n.workflow_version_id, n.updated_on, l.workflow_version_node_link_id,
		       l.parent_workflow_version_node_id, l.parent_output,
		       l.name linkName, l.visibility_settings,
		       l.child_workflow_version_node_id, l.child_input, l.updated_on
		FROM workflow_version_node_link l
		JOIN workflow_version_node n ON n.workflow_version_node_id=l.parent_workflow_version_node_id AND n.status!=$2
		WHERE l.child_workflow_version_node_id=$1;`, workflowVersionNodeId, WorkflowNodeDeleted)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, nil, errors.Wrap(err, database.SqlExecutionError)
		}
	}
	parentNodes := make(map[int]*WorkflowNode)
	parentNodeLinkDetails := make(map[int]WorkflowVersionNodeLink)
	err = parseNodesResultSet(rows, parentNodes, parentNodeLinkDetails)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "Parsing the resulset for parentNodes with workflowVersionNodeId: %v", workflowVersionNodeId)
	}
	return parentNodes, parentNodeLinkDetails, nil
}

func getChildNodes(db *sqlx.DB, workflowVersionNodeId int) (map[int]*WorkflowNode, map[int]WorkflowVersionNodeLink, error) {
	rows, err := db.Queryx(`
		SELECT n.workflow_version_node_id, n.name, n.status, n.stage, n.type, n.parameters, n.visibility_settings,
		       n.workflow_version_id, n.updated_on, l.workflow_version_node_link_id,
		       l.parent_workflow_version_node_id, l.parent_output,
		       l.name linkName, l.visibility_settings,
		       l.child_workflow_version_node_id, l.child_input, l.updated_on
		FROM workflow_version_node_link l
		JOIN workflow_version_node n ON n.workflow_version_node_id=l.child_workflow_version_node_id AND n.status!=$2
		WHERE l.parent_workflow_version_node_id=$1;`, workflowVersionNodeId, WorkflowNodeDeleted)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, nil, errors.Wrap(err, database.SqlExecutionError)
		}
	}
	childNodes := make(map[int]*WorkflowNode)
	childNodeLinkDetails := make(map[int]WorkflowVersionNodeLink)
	err = parseNodesResultSet(rows, childNodes, childNodeLinkDetails)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "Parsing the resulset for childNodes with workflowVersionNodeId: %v", workflowVersionNodeId)
	}
	return childNodes, childNodeLinkDetails, nil
}

func GetTriggerGroupWorkflowVersionNodeId(db *sqlx.DB, workflowVersionNodeId int) (int, error) {
	wfvn, err := GetWorkflowVersionNode(db, workflowVersionNodeId)
	if err != nil {
		return 0, errors.Wrap(err, database.SqlExecutionError)
	}
	if wfvn.Type == core.WorkflowTrigger {
		return wfvn.WorkflowVersionNodeId, nil
	}
	var workflowVersionGroupNodeId int
	err = db.Get(&workflowVersionGroupNodeId, `
		SELECT n.workflow_version_node_id
		FROM workflow_version_node n
		JOIN workflow_version wfv on wfv.workflow_version_id=n.workflow_version_id
		JOIN workflow wf on wf.workflow_id=wfv.workflow_id
        WHERE n.type=$1 AND
              wf.status = ANY ($2) AND
              wfv.status = ANY ($2) AND
			  wfv.workflow_version_id = $3 AND
              n.stage = $4 AND
			  n.status != $5;`,
		core.WorkflowTrigger, pq.Array([]WorkflowStatus{Active, Inactive}),
		wfvn.WorkflowVersionId, wfvn.Stage, WorkflowNodeDeleted)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return 0, errors.Wrap(err, database.SqlExecutionError)
		}
	}
	return workflowVersionGroupNodeId, nil
}

func parseNodesResultSet(rows *sqlx.Rows, nodes map[int]*WorkflowNode, nodeLinkDetails map[int]WorkflowVersionNodeLink) error {
	for rows.Next() {
		var versionNodeId int
		var parentVersionNodeId int
		var childVersionNodeId int
		var name string
		var status WorkflowNodeStatus
		var stage int
		var nodeType core.WorkflowNodeType
		var parameters []byte
		var visibilitySettings WorkflowNodeVisibilitySettings
		var workflowVersionId int
		var updatedOn time.Time
		var linkUpdatedOn time.Time
		var versionNodeLinkId int
		var parentOutput core.WorkflowParameterLabel
		var linkName string
		var linkVisibilitySettings WorkflowVersionNodeLinkVisibilitySettings
		var childInput core.WorkflowParameterLabel
		err := rows.Scan(&versionNodeId, &name, &status, &stage, &nodeType, &parameters, &visibilitySettings, &workflowVersionId, &updatedOn,
			&versionNodeLinkId, &parentVersionNodeId, &parentOutput, &linkName, &linkVisibilitySettings,
			&childVersionNodeId, &childInput, &linkUpdatedOn)
		if err != nil {
			return errors.Wrap(err, "Obtaining workflow node details from the resultSet")
		}
		nodeLinkDetails[versionNodeLinkId] = WorkflowVersionNodeLink{
			WorkflowVersionNodeLinkId:   versionNodeLinkId,
			ParentWorkflowVersionNodeId: parentVersionNodeId,
			ParentOutput:                parentOutput,
			ChildWorkflowVersionNodeId:  childVersionNodeId,
			ChildInput:                  childInput,
			WorkflowVersionId:           workflowVersionId,
			UpdateOn:                    linkUpdatedOn,
			Name:                        linkName,
			VisibilitySettings:          linkVisibilitySettings,
		}
		nodes[versionNodeLinkId] = &WorkflowNode{
			WorkflowVersionNodeId: versionNodeId,
			WorkflowVersionId:     workflowVersionId,
			Type:                  nodeType,
			Status:                status,
			Stage:                 stage,
			Parameters:            string(parameters),
			VisibilitySettings:    visibilitySettings,
			UpdateOn:              updatedOn,
			Name:                  name,
		}
	}
	return nil
}

type WorkflowVersionNodeResponse struct {
	WorkflowVersionNode
	Parameters []byte
}

func GetWorkflowVersionNode(db *sqlx.DB, workflowVersionNodeId int) (WorkflowVersionNode, error) {
	var g WorkflowVersionNodeResponse
	err := db.Get(&g, `SELECT * FROM workflow_version_node WHERE workflow_version_node_id=$1;`, workflowVersionNodeId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return WorkflowVersionNode{}, nil
		}
		return WorkflowVersionNode{}, errors.Wrap(err, database.SqlExecutionError)
	}
	// Unmarshal the parameters
	wfvn := WorkflowVersionNode{
		WorkflowVersionNodeId: g.WorkflowVersionNodeId,
		Name:                  g.Name,
		Stage:                 g.Stage,
		Status:                g.Status,
		Type:                  g.Type,
		VisibilitySettings:    g.VisibilitySettings,
		WorkflowVersionId:     g.WorkflowVersionId,
		CreatedOn:             g.CreatedOn,
		UpdateOn:              g.UpdateOn,
	}
	err = json.Unmarshal(g.Parameters, &wfvn.Parameters)
	if err != nil {
		return WorkflowVersionNode{}, errors.Wrap(err, "Unmarshalling the parameters for the workflow version node")
	}
	return wfvn, nil
}

func createNode(db *sqlx.DB, req CreateNodeRequest) (wfvn WorkflowVersionNode, err error) {
	wfvn.WorkflowVersionId = req.WorkflowVersionId
	wfvn.VisibilitySettings = req.VisibilitySettings
	wfvn.Type = req.Type
	wfvn.Stage = req.Stage
	wfvn.Name = req.Name
	// TODO: We need to add default parameters for the specific node type

	wfvn.Status = WorkflowNodeActive
	wfvn.CreatedOn = time.Now().UTC()
	wfvn.UpdateOn = wfvn.CreatedOn

	visibilitySettingsJson, err := json.Marshal(req.VisibilitySettings)
	if err != nil {
		return WorkflowVersionNode{}, errors.Wrap(err, "Unmarshalling visibilitySettingsJson")
	}

	if req.Parameters != nil {
		wfvn.Parameters, err = json.Marshal(*req.Parameters)
		if err != nil {
			return WorkflowVersionNode{}, errors.Wrap(err, "JSON Marshaling Parameters")
		}
	} else {
		wfvn.Parameters = []byte("{}")
	}

	err = db.QueryRowx(`INSERT
			INTO workflow_version_node
				(name, stage, status, type, parameters, visibility_settings, workflow_version_id, created_on, updated_on)
			VALUES ((SELECT(
					SELECT
						CASE WHEN count(*) = 0 THEN $1
							 ELSE $1 || ' ' || coalesce((
								 SELECT max(coalesce(regexp_replace(name, ($1 || ' (\d+)'), '\1'), '0')::numeric) + 1
								 FROM workflow_version_node
								 WHERE workflow_version_id = $7 and name ~* ($1 || ' (\d+)')
							 ), '1')
						END
					FROM workflow_version_node
					WHERE name = $1 and workflow_version_id = $7)
			    ), $2, $3, $4, $5, $6, $7, $8, $9)
			RETURNING workflow_version_node_id;`,
		wfvn.Name,
		wfvn.Stage,
		wfvn.Status,
		wfvn.Type,
		wfvn.Parameters,
		visibilitySettingsJson,
		wfvn.WorkflowVersionId,
		wfvn.CreatedOn,
		wfvn.UpdateOn).Scan(&wfvn.WorkflowVersionNodeId)

	if err != nil {
		if err, ok := err.(*pq.Error); ok {
			if err.Code == "23505" {
				return WorkflowVersionNode{}, database.SqlUniqueConstraintError
			}
		}
		return WorkflowVersionNode{}, errors.Wrap(err, database.SqlExecutionError)
	}

	return wfvn, nil
}

func updateNode(db *sqlx.DB, req UpdateNodeRequest) (int, error) {

	qb := sq.Update("workflow_version_node").PlaceholderFormat(sq.Dollar).
		Set("updated_on", time.Now().UTC())

	if req.Name != nil {
		qb = qb.Set("name", req.Name)
	}

	if req.Status != nil {
		qb = qb.Set("status", req.Status)
	}

	if req.Parameters != nil {
		p, err := json.Marshal(*req.Parameters)
		if err != nil {
			return 0, errors.Wrap(err, "JSON Marshaling Parameters")
		}
		qb = qb.Set("parameters", p)
	}

	_, err := qb.Where(sq.Eq{"workflow_version_node_id": req.WorkflowVersionNodeId}).RunWith(db).Exec()
	if err != nil {
		// TODO: Handle wrapped error while still being able to check for unique constraint violation
		return 0, err //nolint:wrapcheck
	}

	return req.WorkflowVersionNodeId, nil
}

func updateNodeDeleted(db *sqlx.DB, workflowNodeId int) error {
	_, err := db.Exec(`
		UPDATE workflow_version_node
		SET updated_on=$1, name=name||'_DELETED_'||$2, status=$3
		WHERE workflow_version_node_id=$4`,
		time.Now().UTC(), time.Now().UnixMilli(), WorkflowNodeDeleted, workflowNodeId)
	if err != nil {
		return errors.Wrap(err, database.SqlExecutionError)
	}

	return nil
}

func updateNodeVisibilitySettings(db *sqlx.DB, workflowVersionNodeId int, visibilitySettings WorkflowNodeVisibilitySettings) error {
	vs, err := json.Marshal(visibilitySettings)
	if err != nil {
		return errors.Wrap(err, "JSON Marshaling Visibility Settings")
	}

	_, err = db.Exec(`UPDATE workflow_version_node
		SET visibility_settings=$1, updated_on=$2
		WHERE workflow_version_node_id=$3;`,
		vs, time.Now(), workflowVersionNodeId)
	if err != nil {
		return errors.Wrap(err, database.SqlExecutionError)
	}

	return nil
}

func deleteNode(db *sqlx.DB, workflowVersionNodeId int) (int, error) {
	tx := db.MustBegin()

	// delete all node logs that are connected to this node,
	_, err := tx.Exec(`
		DELETE FROM workflow_version_node_log
		       WHERE workflow_version_node_id=$1 OR triggering_workflow_version_node_id=$1`,
		workflowVersionNodeId)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return 0, errors.Wrap(rollbackErr, "Deleting the node links while deleting node (rollback failed too)")
		}
		return 0, errors.Wrap(err, "Deleting the node links while deleting node")
	}

	// delete all node Links that are connected to this node
	_, err = tx.Exec(`
		DELETE FROM workflow_version_node_link
		       WHERE parent_workflow_version_node_id=$1 OR child_workflow_version_node_id=$1`,
		workflowVersionNodeId)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return 0, errors.Wrap(rollbackErr, "Deleting the node links while deleting node (rollback failed too)")
		}
		return 0, errors.Wrap(err, "Deleting the node links while deleting node")
	}

	res, err := tx.Exec(`DELETE FROM workflow_version_node WHERE workflow_version_node_id=$1;`,
		workflowVersionNodeId)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return 0, errors.Wrap(rollbackErr, "Deleting the node (rollback failed too)")
		}
		return 0, errors.Wrap(err, "Deleting the node")
	}

	err = tx.Commit()
	if err != nil {
		return 0, errors.Wrap(err, "Committing the transaction while deleting the node")
	}

	_, err = res.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, database.SqlAffectedRowsCheckError)
	}

	return workflowVersionNodeId, nil
}

// deleteStage deletes a stage by deleting all nodes and node links in a stage
func deleteStage(db *sqlx.DB, workflowVersionId int, stage int) error {
	tx := db.MustBegin()
	var workflowVersionNodeIds []int
	err := tx.Select(&workflowVersionNodeIds, `
		SELECT workflow_version_node_id
		FROM workflow_version_node
		WHERE workflow_version_id = $1 AND stage = $2`, workflowVersionId, stage)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return errors.Wrap(rollbackErr, "Obtaining the node id while deleting node (rollback failed too)")
		}
		return errors.Wrap(err, "Obtaining the node id while deleting node")
	}

	// delete all node logs that are connected to this node,
	_, err = tx.Exec(`
		DELETE FROM workflow_version_node_log
		       WHERE workflow_version_node_id = ANY ($1) OR triggering_workflow_version_node_id = ANY ($1)`,
		pq.Array(workflowVersionNodeIds))
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return errors.Wrap(rollbackErr, "Deleting the node links while deleting node (rollback failed too)")
		}
		return errors.Wrap(err, "Deleting the node links while deleting node")
	}

	// Delete all workflow_version_node_link where workflow_version_node_id is in the stage
	_, err = tx.Exec(`
		DELETE FROM workflow_version_node_link
		       WHERE parent_workflow_version_node_id = ANY ($1) OR child_workflow_version_node_id = ANY ($1);`,
		pq.Array(workflowVersionNodeIds))
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return errors.Wrap(rollbackErr, "deleting workflow stage node links failed (rollback failed too)")
		}
		return errors.Wrap(err, "deleting workflow stage node links failed")
	}

	_, err = tx.Exec(`DELETE FROM workflow_version_node WHERE workflow_version_node_id = ANY ($1);`,
		pq.Array(workflowVersionNodeIds))
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return errors.Wrap(rollbackErr, "deleting workflow stage nodes failed (rollback failed too)")
		}
		return errors.Wrap(err, "deleting workflow stage nodes failed")
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "committing workflow stage deletion failed")
	}

	return nil
}

func getWorkflowVersionNodeLink(db *sqlx.DB, workflowVersionNodeLinkId int) (WorkflowVersionNodeLink, error) {
	var wfvnl WorkflowVersionNodeLink
	err := db.Get(&wfvnl, `SELECT * FROM workflow_version_node_link WHERE workflow_version_node_link_id = $1 limit 1;`,
		workflowVersionNodeLinkId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return WorkflowVersionNodeLink{}, nil
		}
		return WorkflowVersionNodeLink{}, errors.Wrap(err, database.SqlExecutionError)
	}
	return wfvnl, nil
}

func GetWorkflowVersionNodeLinks(db *sqlx.DB, workflowVersionId int) ([]WorkflowVersionNodeLink, error) {
	var links []WorkflowVersionNodeLink
	err := db.Select(&links, `
			-- select all the links belonging to this workflow and join in the parent node and add the stage to the link
			SELECT
				l.workflow_version_node_link_id,
				l.parent_workflow_version_node_id,
				l.child_workflow_version_node_id,
				l.parent_output,
				l.child_input,
				l.name,
				l.visibility_settings,
				l.created_on,
				l.updated_on,
				l.workflow_version_id,
				n.stage AS stage
			FROM workflow_version_node_link l
			-- Only joining in the parent node because nodes should never be joiner across stages
			LEFT JOIN workflow_version_node as n ON n.workflow_version_node_id = l.parent_workflow_version_node_id
			WHERE l.workflow_version_id = $1
			ORDER BY l.created_on;
			`, workflowVersionId)
	if err != nil {
		return nil, errors.Wrap(err, database.SqlExecutionError)
	}

	return links, nil
}

func addWorkflowVersionNodeLink(db *sqlx.DB, req CreateWorkflowVersionNodeLinkRequest) (WorkflowVersionNodeLink, error) {
	var wvnl WorkflowVersionNodeLink

	wvnl.CreatedOn = time.Now().UTC()
	wvnl.UpdateOn = time.Now().UTC()
	wvnl.VisibilitySettings = WorkflowVersionNodeLinkVisibilitySettings{}
	visibilitySettingsBytes, err := json.Marshal(wvnl.VisibilitySettings)
	if err != nil {
		return WorkflowVersionNodeLink{}, errors.Wrap(err, "JSON Marshaling Node Link VisibilitySettings")
	}
	// NB: we need to add the same name guard in the SQL if we give the user the ability to change link names.
	wvnl.Name = wvnl.CreatedOn.Format("20060102.150405.000000")

	wvnl.WorkflowVersionId = req.WorkflowVersionId
	wvnl.ParentWorkflowVersionNodeId = req.ParentWorkflowVersionNodeId
	wvnl.ChildWorkflowVersionNodeId = req.ChildWorkflowVersionNodeId
	wvnl.ChildInput = req.ChildInput
	wvnl.ParentOutput = req.ParentOutput

	//if wvnl.ChildInput != wvnl.ParentOutput {
	//	childWorkflowParameterType, err := verifyChildInput(db, wvnl)
	//	if err != nil {
	//		return WorkflowVersionNodeLink{},
	//			errors.Wrapf(err, "verification of child input failed for ChildInput: %v and ChildWorkflowVersionNodeId: %v",
	//				wvnl.ChildInput, wvnl.ChildWorkflowVersionNodeId)
	//	}
	//
	//	parentWorkflowParameterType, err := verifyParentOutput(db, wvnl)
	//	if err != nil {
	//		return WorkflowVersionNodeLink{},
	//			errors.Wrapf(err, "verification of child input failed for ParentOutput: %v and ParentWorkflowVersionNodeId: %v",
	//				wvnl.ParentOutput, wvnl.ParentWorkflowVersionNodeId)
	//	}
	//
	//	if !slices.Contains(commons.GetWorkflowParameterTypeGroup(parentWorkflowParameterType), childWorkflowParameterType) {
	//		return WorkflowVersionNodeLink{},
	//			errors.Wrapf(err, "ParentOutput: %v is incompatible with ChildInput: %v", wvnl.ParentOutput, wvnl.ChildInput)
	//	}
	//}

	err = db.QueryRowx(`INSERT INTO workflow_version_node_link
    	(name,
    	 visibility_settings,
    	 parent_output,
    	 parent_workflow_version_node_id,
    	 child_input,
    	 child_workflow_version_node_id,
    	 workflow_version_id,
    	 created_on,
    	 updated_on)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING workflow_version_node_link_id;`,
		wvnl.Name,
		visibilitySettingsBytes,
		wvnl.ParentOutput,
		wvnl.ParentWorkflowVersionNodeId,
		wvnl.ChildInput,
		wvnl.ChildWorkflowVersionNodeId,
		wvnl.WorkflowVersionId,
		wvnl.CreatedOn,
		wvnl.UpdateOn).Scan(&wvnl.WorkflowVersionNodeLinkId)
	if err != nil {
		if err, ok := err.(*pq.Error); ok {
			if err.Code == "23505" {
				return WorkflowVersionNodeLink{}, database.SqlUniqueConstraintError
			}
		}
		return WorkflowVersionNodeLink{}, errors.Wrap(err, database.SqlExecutionError)
	}
	return wvnl, nil
}

func updateWorkflowVersionNodeLink(db *sqlx.DB, workflowVersionNodeLink WorkflowVersionNodeLink) (WorkflowVersionNodeLink, error) {
	workflowVersionNodeLink.UpdateOn = time.Now().UTC()
	_, err := db.Exec(`UPDATE workflow_version_node_link
		SET name=$1, visibility_settings=$2, parent_output=$3,
		    parent_workflow_version_node_id=$4, child_input=$5, child_workflow_version_node_id=$6,
		    workflow_version_id=$7, updated_on=$8
		WHERE workflow_version_node_link_id=$9;`,
		workflowVersionNodeLink.Name, workflowVersionNodeLink.VisibilitySettings,
		workflowVersionNodeLink.ParentOutput, workflowVersionNodeLink.ParentWorkflowVersionNodeId,
		workflowVersionNodeLink.ChildInput, workflowVersionNodeLink.ChildWorkflowVersionNodeId,
		workflowVersionNodeLink.WorkflowVersionId,
		workflowVersionNodeLink.UpdateOn, workflowVersionNodeLink.WorkflowVersionNodeLinkId)
	if err != nil {
		if err, ok := err.(*pq.Error); ok {
			if err.Code == "23505" {
				return WorkflowVersionNodeLink{}, database.SqlUniqueConstraintError
			}
		}
		return WorkflowVersionNodeLink{}, errors.Wrap(err, database.SqlExecutionError)
	}
	return workflowVersionNodeLink, nil
}

func removeNodeLink(db *sqlx.DB, workflowVersionNodeLinkId int) (int64, error) {
	res, err := db.Exec(`DELETE FROM workflow_version_node_link WHERE workflow_version_node_link_id = $1;`, workflowVersionNodeLinkId)
	if err != nil {
		return 0, errors.Wrap(err, database.SqlExecutionError)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, database.SqlAffectedRowsCheckError)
	}
	return rowsAffected, nil
}

func addWorkflowVersionNodeLog(db *sqlx.DB, workflowVersionNodeLog WorkflowVersionNodeLog) (WorkflowVersionNodeLog, error) {
	workflowVersionNodeLog.CreatedOn = time.Now().UTC()
	_, err := db.Exec(`INSERT INTO workflow_version_node_log
    	(trigger_reference, input_data, output_data, debug_data, error_data, workflow_version_node_id, triggering_workflow_version_node_id, created_on)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8);`,
		workflowVersionNodeLog.TriggerReference,
		workflowVersionNodeLog.InputData, workflowVersionNodeLog.OutputData, workflowVersionNodeLog.DebugData,
		workflowVersionNodeLog.ErrorData, workflowVersionNodeLog.WorkflowVersionNodeId,
		workflowVersionNodeLog.TriggeringWorkflowVersionNodeId, workflowVersionNodeLog.CreatedOn)
	if err != nil {
		return WorkflowVersionNodeLog{}, errors.Wrap(err, database.SqlExecutionError)
	}
	return workflowVersionNodeLog, nil
}

func GetNodeLogs(db *sqlx.DB, workflowVersionNodeId int, maximumResultCount int) ([]WorkflowVersionNodeLog, error) {
	var wfvnls []WorkflowVersionNodeLog
	err := db.Get(&wfvnls,
		"SELECT * FROM workflow_version_node_log WHERE workflow_version_node_id=$1 ORDER BY created_on DESC LIMIT $2;",
		workflowVersionNodeId, maximumResultCount)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []WorkflowVersionNodeLog{}, nil
		}
		return nil, errors.Wrap(err, database.SqlExecutionError)
	}
	return wfvnls, nil
}

func GetWorkflowLogs(db *sqlx.DB, workflowId int, maximumResultCount int) ([]WorkflowVersionNodeLog, error) {
	var wfvnls []WorkflowVersionNodeLog
	err := db.Get(&wfvnls, `SELECT *
				FROM workflow_version_node_log wfvnls
				JOIN workflow_version_node wfvn ON wfvn.workflow_version_node_id=wfvnls.workflow_version_node_id
				JOIN workflow_version wfv ON wfv.workflow_version_id=wfvn.workflow_version_id
				WHERE wfv.workflow_id=$1
				ORDER BY wfvnls.created_on DESC
				LIMIT $2;`, workflowId, maximumResultCount)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []WorkflowVersionNodeLog{}, nil
		}
		return nil, errors.Wrap(err, database.SqlExecutionError)
	}
	return wfvnls, nil
}
