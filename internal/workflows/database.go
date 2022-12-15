package workflows

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/pkg/errors"

	"github.com/lncapital/torq/internal/database"
	"github.com/lncapital/torq/pkg/commons"
)

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

func GetWorkflowByWorkflowVersionId(db *sqlx.DB, workflowVersionId int) (Workflow, error) {
	var wf Workflow
	err := db.Get(&wf, `
		SELECT wf.*
		FROM workflow_version wfv
		JOIN workflow wf ON wf.workflow_id=wfv.workflow_id
		WHERE wfv.workflow_version_id=$1;`, workflowVersionId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Workflow{}, nil
		}
		return Workflow{}, errors.Wrap(err, database.SqlExecutionError)
	}
	return wf, nil
}

func GetWorkflows(db *sqlx.DB) ([]Workflow, error) {
	var wfs []Workflow
	err := db.Select(&wfs, `SELECT * FROM workflow;`)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []Workflow{}, nil
		}
		return []Workflow{}, errors.Wrap(err, database.SqlExecutionError)
	}
	return wfs, nil
}

func addWorkflow(db *sqlx.DB, workflow Workflow) (Workflow, error) {
	workflow.CreatedOn = time.Now().UTC()
	workflow.UpdateOn = workflow.CreatedOn
	err := db.QueryRowx(`INSERT INTO workflow (name, type, status, created_on, updated_on)
		VALUES ($1, $2, $3, $4, $5) RETURNING workflow_id;`,
		workflow.Name, workflow.Type, workflow.Status, workflow.CreatedOn, workflow.UpdateOn).Scan(&workflow.WorkflowId)
	if err != nil {
		if err, ok := err.(*pq.Error); ok {
			if err.Code == "23505" {
				return Workflow{}, errors.Wrap(err, database.SqlUniqueConstraintError)
			}
		}
		return Workflow{}, errors.Wrap(err, database.SqlExecutionError)
	}
	return workflow, nil
}

func setWorkflow(db *sqlx.DB, workflow Workflow) (Workflow, error) {
	workflow.UpdateOn = time.Now().UTC()
	_, err := db.Exec(`UPDATE workflow SET name=$1, type=$2, status=$3, updated_on=$5 WHERE workflow_id=$6;`,
		workflow.Name, workflow.Type, workflow.Status, workflow.UpdateOn, workflow.WorkflowId)
	if err != nil {
		if err, ok := err.(*pq.Error); ok {
			if err.Code == "23505" {
				return Workflow{}, errors.Wrap(err, database.SqlUniqueConstraintError)
			}
		}
		return Workflow{}, errors.Wrap(err, database.SqlExecutionError)
	}
	return workflow, nil
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

func GetWorkflowVersion(db *sqlx.DB, workflowVersionId int) (WorkflowVersion, error) {
	var wfv WorkflowVersion
	err := db.Get(&wfv, `SELECT * FROM workflow_version WHERE workflow_version_id=$1;`, workflowVersionId)
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

func cloneWorkflowVersion(db *sqlx.DB, workflowId int, version int) (WorkflowVersion, error) {
	var wfv WorkflowVersion
	var err error
	if version == 0 {
		//close latest or when nothing exists create new empty version
		err = db.Get(&wfv, `SELECT * FROM workflow_version WHERE workflow_id=$1 ORDER BY version DESC LIMIT 1;`, workflowId)
	} else {
		err = db.Get(&wfv, `SELECT * FROM workflow_version WHERE workflow_id=$1 AND version=$2;`, workflowId, version)
	}
	if err != nil {
		return WorkflowVersion{}, errors.Wrap(err, database.SqlExecutionError)
	}
	if wfv.WorkflowVersionId == 0 {
		wfv = WorkflowVersion{}
		wfv.CreatedOn = time.Now().UTC()
		wfv.UpdateOn = wfv.CreatedOn
		wfv.WorkflowId = workflowId
		wfv.Version = version
		err := db.QueryRowx(`INSERT INTO workflow_version (version, status, workflow_id, created_on, updated_on)
		VALUES ($1, $2, $3, $4, $5) RETURNING workflow_version_id;`,
			wfv.Version, wfv.Status, wfv.WorkflowId, wfv.CreatedOn, wfv.UpdateOn).Scan(&wfv.WorkflowVersionId)
		if err != nil {
			return WorkflowVersion{}, errors.Wrap(err, database.SqlExecutionError)
		}
	} else {
		// TODO ACTUALLY CLONE
	}
	return wfv, nil
}

func setWorkflowVersion(db *sqlx.DB, workflowVersion WorkflowVersion) (WorkflowVersion, error) {
	workflowVersion.UpdateOn = time.Now().UTC()
	_, err := db.Exec(`UPDATE workflow_version SET version=$1, status=$2, updated_on=$3 WHERE workflow_version_id=$4;`,
		workflowVersion.Version, workflowVersion.Status, workflowVersion.UpdateOn, workflowVersion.WorkflowVersionId)
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

func GetActiveTriggerNodes(db *sqlx.DB) ([]WorkflowNode, error) {
	var workflowVersionRootNodeIds []int
	err := db.Select(&workflowVersionRootNodeIds, `
		SELECT wfvn.workflow_version_node_id
		FROM workflow_version_node wfvn
		JOIN workflow_version wfv ON wfv.workflow_version_id = wfvn.workflow_version_id AND wfv.status=1
		JOIN workflow wf ON wf.workflow_id = wfv.workflow_id AND wfv.status=1
		LEFT JOIN workflow_version_node_link parentLink ON parentLink.child_workflow_version_node_id = wfvn.workflow_version_node_id
		WHERE wfvn.status=$1 AND wfvn.type IN ($2,$3) AND parentLink.child_workflow_version_node_id IS NULL AND wfv.workflow_version_id IN (
			SELECT ranked.workflow_version_id
			FROM (
				SELECT v_wfv.workflow_version_id, RANK() OVER (PARTITION BY v_wfv.workflow_version_id ORDER BY version DESC) version_rank
				FROM workflow_version v_wfv
				WHERE v_wfv.status=$1
			) ranked
			WHERE ranked.version_rank = 1
		);`, commons.Active, commons.WorkflowNodeTimeTrigger, commons.WorkflowNodeChannelBalanceEventTrigger)
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

func ShouldTrigger(db *sqlx.DB, nodeSettings commons.ManagedNodeSettings, triggerNode WorkflowNode) (bool, error) {
	triggerSettings := commons.GetTriggerSettingsByWorkflowVersionId(nodeSettings.NodeId, triggerNode.WorkflowVersionId)
	if triggerSettings.Status == commons.Active {
		return false, errors.New(fmt.Sprintf("Trigger is already active with reference: %v.", triggerSettings.Reference))
	}

	defer commons.SetTriggerVerificationTime(nodeSettings.NodeId, triggerNode.WorkflowVersionId, time.Now())

	switch triggerNode.Type {
	case commons.WorkflowNodeTimeTrigger:
		triggerParameters, err := getWorkflowNodeParameters(triggerNode)
		if err != nil {
			return false, errors.Wrapf(err, "Obtaining trigger parameters for WorkflowVersionId: %v", triggerNode.WorkflowVersionId)
		}
		for _, triggerParameter := range triggerParameters.Parameters {
			if triggerParameter.Type == commons.WorkflowParameterTimeInSeconds && triggerParameter.ValueNumber != 0 {
				if triggerSettings.BootTime == nil {
					return true, nil
				}
				if int(time.Since(*triggerSettings.BootTime).Seconds()) > triggerParameter.ValueNumber {
					return true, nil
				}
			}
		}
	case commons.WorkflowNodeChannelBalanceEventTrigger:
		if triggerSettings.WorkflowVersionId != 0 {
			// Trigger is event based and was already initially verified since Torq booted
			return false, nil
		}
		return true, nil
	}
	return false, errors.New(fmt.Sprintf("Could not find the necessary parameters to verify ShouldTrigger for WorkflowVersionId: %v", triggerNode.WorkflowVersionId))
}

func getWorkflowNodeParameters(triggerNode WorkflowNode) (WorkflowNodeParameters, error) {
	var triggerParameters WorkflowNodeParameters
	err := json.Unmarshal([]byte(triggerNode.Parameters), &triggerParameters)
	if err != nil {
		return WorkflowNodeParameters{}, errors.Wrap(err, "JSON unmarshal")
	}
	return triggerParameters, nil
}

func getWorkflowNodeParameter(parameters WorkflowNodeParameters, parameterType commons.WorkflowParameterType) WorkflowNodeParameter {
	for _, parameter := range parameters.Parameters {
		if parameter.Type == parameterType {
			return parameter
		}
	}
	return WorkflowNodeParameter{}
}

// GetWorkflowNode is not recursive and only returns direct parent/child relations without further nesting.
func GetWorkflowNode(db *sqlx.DB, workflowVersionNodeId int) (WorkflowNode, error) {
	var wfvn WorkflowVersionNode
	err := db.Get(&wfvn, `SELECT * FROM workflow_version_node WHERE workflow_version_node_id=$1;`, workflowVersionNodeId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return WorkflowNode{}, nil
		}
		return WorkflowNode{}, errors.Wrap(err, database.SqlExecutionError)
	}

	parentNodes, err := getParentNodes(db, workflowVersionNodeId)
	if err != nil {
		return WorkflowNode{}, errors.Wrapf(err, "Obtaining parent nodes for workflowVersionNodeId: %v", workflowVersionNodeId)
	}

	childNodes, err := getChildNodes(db, workflowVersionNodeId)
	if err != nil {
		return WorkflowNode{}, errors.Wrapf(err, "Obtaining child nodes for workflowVersionNodeId: %v", workflowVersionNodeId)
	}

	response := wfvn.GetWorkflowNodeStructured()
	if len(parentNodes) > 0 {
		var parentNodesStructured map[int][]*WorkflowNode
		for parameterIndex, parentNodeArray := range parentNodes {
			for _, parentNode := range parentNodeArray {
				parentNodesStructured[parameterIndex] = append(parentNodesStructured[parameterIndex], &parentNode)
			}
		}
		response.ParentNodes = parentNodesStructured
	}
	if len(childNodes) > 0 {
		var childNodesStructured map[int][]*WorkflowNode
		for parameterIndex, childNodeArray := range childNodes {
			for _, childNode := range childNodeArray {
				childNodesStructured[parameterIndex] = append(childNodesStructured[parameterIndex], &childNode)
			}
		}
		response.ParentNodes = childNodesStructured
	}
	return response, nil
}

// GetWorkflowTree this code assumes recursive loops from the root nodes are not allowed
func GetWorkflowTree(db *sqlx.DB, workflowVersionId int) (WorkflowTree, error) {
	processedNodes := make(map[int]*WorkflowNode)
	var rootNodes []WorkflowVersionNode
	err := db.Select(&rootNodes, `
		SELECT n.*
		FROM workflow_version_node n
		LEFT JOIN workflow_version_node_link parentLink ON parentLink.child_workflow_version_node_id = n.workflow_version_node_id
		WHERE n.workflow_version_id=$1 AND parentLink.child_workflow_version_node_id IS NULL;`, workflowVersionId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return WorkflowTree{}, nil
		}
		return WorkflowTree{}, errors.Wrap(err, database.SqlExecutionError)
	}

	var rootNodesStructured []*WorkflowNode
	if len(rootNodes) > 0 {
		for _, workflowNode := range rootNodes {
			workflowNodeStructured := workflowNode.GetWorkflowNodeStructured()
			workflowNodeStructured.ParentNodes = map[int][]*WorkflowNode{}
			err = processNodeRecursion(processedNodes, db, &workflowNodeStructured)
			if err != nil {
				return WorkflowTree{}, err
			}
		}
	}
	return WorkflowTree{RootNodes: rootNodesStructured}, nil
}

func processNodeRecursion(processedNodes map[int]*WorkflowNode, db *sqlx.DB, workflowNode *WorkflowNode) error {
	childNodes, err := getChildNodes(db, workflowNode.WorkflowVersionNodeId)
	if err != nil {
		return errors.Wrapf(err, "Obtaining child nodes for workflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
	}
	childNodesStructured := make(map[int][]*WorkflowNode)
	for parameterIndex, childNodeArray := range childNodes {
		for _, childNode := range childNodeArray {
			alreadyProcessedNode := processedNodes[childNode.WorkflowVersionNodeId]
			if alreadyProcessedNode != nil && alreadyProcessedNode.WorkflowVersionNodeId != 0 {
				//alreadyProcessedNode.ParentNodes = append(alreadyProcessedNode.ParentNodes, workflowNode)
				childNodesStructured[parameterIndex] = append(childNodesStructured[parameterIndex], alreadyProcessedNode)
			} else {
				//childNodeStructured.ParentNodes = append(childNodeStructured.ParentNodes, workflowNode)
				childNodesStructured[parameterIndex] = append(childNodesStructured[parameterIndex], &childNode)
				err = processNodeRecursion(processedNodes, db, &childNode)
				if err != nil {
					return errors.Wrapf(err, "Obtaining child nodes recursive for workflowVersionNodeId: %v", childNode.WorkflowVersionNodeId)
				}
			}
		}
	}
	workflowNode.ChildNodes = childNodesStructured
	return nil
}

func getParentNodes(db *sqlx.DB, workflowVersionNodeId int) (map[int][]WorkflowNode, error) {
	rows, err := db.Query(`
		SELECT n.workflow_version_node_id, n.name, n.status, n.type, n.parameters, n.visibility_settings,
		       n.workflow_version_id, n.updated_on, l.child_parameter_index
		FROM workflow_version_node_link l
		JOIN workflow_version_node n ON n.workflow_version_node_id=l.child_workflow_version_node_id
		WHERE l.child_workflow_version_node_id=$1
		ORDER BY l.child_parameter_index;`, workflowVersionNodeId)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, errors.Wrap(err, database.SqlExecutionError)
		}
	}
	parentNodes := make(map[int][]WorkflowNode)
	err = parseNodesResultSet(rows, parentNodes)
	if err != nil {
		return nil, errors.Wrapf(err, "Parsing the resulset for parentNodes with workflowVersionNodeId: %v", workflowVersionNodeId)
	}
	return parentNodes, nil
}

func getChildNodes(db *sqlx.DB, workflowVersionNodeId int) (map[int][]WorkflowNode, error) {
	rows, err := db.Query(`
		SELECT n.workflow_version_node_id, n.name, n.status, n.type, n.parameters, n.visibility_settings,
		       n.workflow_version_id, n.updated_on, l.parent_parameter_index
		FROM workflow_version_node_link l
		JOIN workflow_version_node n ON n.workflow_version_node_id=l.parent_workflow_version_node_id
		WHERE l.parent_workflow_version_node_id=$1
		ORDER BY l.parent_parameter_index;`, workflowVersionNodeId)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, errors.Wrap(err, database.SqlExecutionError)
		}
	}
	childNodes := make(map[int][]WorkflowNode)
	err = parseNodesResultSet(rows, childNodes)
	if err != nil {
		return nil, errors.Wrapf(err, "Parsing the resulset for childNodes with workflowVersionNodeId: %v", workflowVersionNodeId)
	}
	return childNodes, nil
}

func parseNodesResultSet(rows *sql.Rows, parentNodes map[int][]WorkflowNode) error {
	for rows.Next() {
		var versionNodeId int
		var name string
		var status commons.Status
		var nodeType commons.WorkflowNodeType
		var parameters string
		var visibilitySettings string
		var versionId int
		var updatedOn time.Time
		var parameterIndex int
		err := rows.Scan(&versionNodeId, &name, &status, &nodeType, &parameters, &visibilitySettings, &versionId, &updatedOn, &parameterIndex)
		if err != nil {
			return errors.Wrap(err, "Obtaining nodeId and publicKey from the resultSet")
		}
		parentNodes[parameterIndex] = append(parentNodes[parameterIndex], WorkflowNode{
			WorkflowVersionNodeId: versionNodeId,
			WorkflowVersionId:     versionId,
			Type:                  nodeType,
			Status:                status,
			Parameters:            parameters,
			VisibilitySettings:    visibilitySettings,
			UpdateOn:              updatedOn,
			Name:                  name,
		})
	}
	return nil
}

func GetWorkflowVersionNode(db *sqlx.DB, workflowVersionNodeId int) (WorkflowVersionNode, error) {
	var wfvn WorkflowVersionNode
	err := db.Get(&wfvn, `SELECT * FROM workflow_version_node WHERE workflow_version_node_id=$1;`, workflowVersionNodeId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return WorkflowVersionNode{}, nil
		}
		return WorkflowVersionNode{}, errors.Wrap(err, database.SqlExecutionError)
	}
	return wfvn, nil
}

func addWorkflowVersionNode(db *sqlx.DB, workflowVersionNode WorkflowVersionNode) (WorkflowVersionNode, error) {
	workflowVersionNode.CreatedOn = time.Now().UTC()
	workflowVersionNode.UpdateOn = workflowVersionNode.CreatedOn
	err := db.QueryRowx(`INSERT INTO workflow_version_node
    	(name, status, type, parameters, visibility_settings, workflow_version_id, created_on, updated_on)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING workflow_version_node_id;`,
		workflowVersionNode.Name, workflowVersionNode.Status, workflowVersionNode.Type, workflowVersionNode.Parameters,
		workflowVersionNode.VisibilitySettings, workflowVersionNode.WorkflowVersionId, workflowVersionNode.CreatedOn,
		workflowVersionNode.UpdateOn).Scan(&workflowVersionNode.WorkflowVersionNodeId)
	if err != nil {
		if err, ok := err.(*pq.Error); ok {
			if err.Code == "23505" {
				return WorkflowVersionNode{}, errors.Wrap(err, database.SqlUniqueConstraintError)
			}
		}
		return WorkflowVersionNode{}, errors.Wrap(err, database.SqlExecutionError)
	}
	return workflowVersionNode, nil
}

func setWorkflowVersionNode(db *sqlx.DB, workflowVersionNode WorkflowVersionNode) (WorkflowVersionNode, error) {
	workflowVersionNode.UpdateOn = time.Now().UTC()
	_, err := db.Exec(`UPDATE workflow_version_node
		SET name=$1, status=$2, type=$3, parameters=$4, visibility_settings=$5, workflow_version_id=$6, updated_on=$7
		WHERE workflow_version_node_id=$8;`,
		workflowVersionNode.Name, workflowVersionNode.Status, workflowVersionNode.Type, workflowVersionNode.Parameters,
		workflowVersionNode.VisibilitySettings, workflowVersionNode.WorkflowVersionId, workflowVersionNode.UpdateOn,
		workflowVersionNode.WorkflowVersionNodeId)
	if err != nil {
		if err, ok := err.(*pq.Error); ok {
			if err.Code == "23505" {
				return WorkflowVersionNode{}, errors.Wrap(err, database.SqlUniqueConstraintError)
			}
		}
		return WorkflowVersionNode{}, errors.Wrap(err, database.SqlExecutionError)
	}
	return workflowVersionNode, nil
}

func removeWorkflowVersionNode(db *sqlx.DB, workflowVersionNodeId int) (int64, error) {
	res, err := db.Exec(`DELETE FROM workflow_version_node WHERE workflow_version_node_id = $1;`, workflowVersionNodeId)
	if err != nil {
		return 0, errors.Wrap(err, database.SqlExecutionError)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, database.SqlAffectedRowsCheckError)
	}
	return rowsAffected, nil
}

func addWorkflowVersionNodeLink(db *sqlx.DB, workflowVersionNodeLink WorkflowVersionNodeLink) (WorkflowVersionNodeLink, error) {
	workflowVersionNodeLink.CreatedOn = time.Now().UTC()
	workflowVersionNodeLink.UpdateOn = workflowVersionNodeLink.CreatedOn
	err := db.QueryRowx(`INSERT INTO workflow_version_node_link
    	(visibility_settings, parent_parameter_index, parent_workflow_version_node_id, child_parameter_index, child_workflow_version_node_id, created_on, updated_on)
		VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING workflow_version_node_link_id;`,
		workflowVersionNodeLink.VisibilitySettings,
		workflowVersionNodeLink.ParentParameterIndex, workflowVersionNodeLink.ParentWorkflowVersionNodeId,
		workflowVersionNodeLink.ChildParameterIndex, workflowVersionNodeLink.ChildWorkflowVersionNodeId,
		workflowVersionNodeLink.CreatedOn, workflowVersionNodeLink.UpdateOn).Scan(&workflowVersionNodeLink.WorkflowVersionNodeLinkId)
	if err != nil {
		if err, ok := err.(*pq.Error); ok {
			if err.Code == "23505" {
				return WorkflowVersionNodeLink{}, errors.Wrap(err, database.SqlUniqueConstraintError)
			}
		}
		return WorkflowVersionNodeLink{}, errors.Wrap(err, database.SqlExecutionError)
	}
	return workflowVersionNodeLink, nil
}

func setWorkflowVersionNodeLink(db *sqlx.DB, workflowVersionNodeLink WorkflowVersionNodeLink) (WorkflowVersionNodeLink, error) {
	workflowVersionNodeLink.UpdateOn = time.Now().UTC()
	_, err := db.Exec(`UPDATE workflow_version_node_link
		SET visibility_settings=$1, parent_parameter_index=$2, parent_workflow_version_node_id=$3,
		    child_parameter_index=$4, child_workflow_version_node_id=$5, updated_on=$6
		WHERE workflow_version_node_link_id=$7;`,
		workflowVersionNodeLink.VisibilitySettings,
		workflowVersionNodeLink.ParentParameterIndex, workflowVersionNodeLink.ParentWorkflowVersionNodeId,
		workflowVersionNodeLink.ChildParameterIndex, workflowVersionNodeLink.ChildWorkflowVersionNodeId,
		workflowVersionNodeLink.UpdateOn, workflowVersionNodeLink.WorkflowVersionNodeLinkId)
	if err != nil {
		if err, ok := err.(*pq.Error); ok {
			if err.Code == "23505" {
				return WorkflowVersionNodeLink{}, errors.Wrap(err, database.SqlUniqueConstraintError)
			}
		}
		return WorkflowVersionNodeLink{}, errors.Wrap(err, database.SqlExecutionError)
	}
	return workflowVersionNodeLink, nil
}

func removeWorkflowVersionNodeLink(db *sqlx.DB, workflowVersionNodeLinkId int) (int64, error) {
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
    	(node_id, trigger_reference, input_data, output_data, debug_data, error_data, workflow_version_node_id, triggered_workflow_version_node_id, created_on)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);`,
		workflowVersionNodeLog.NodeId, workflowVersionNodeLog.TriggerReference,
		workflowVersionNodeLog.InputData, workflowVersionNodeLog.OutputData, workflowVersionNodeLog.DebugData,
		workflowVersionNodeLog.ErrorData, workflowVersionNodeLog.WorkflowVersionNodeId,
		workflowVersionNodeLog.TriggeredWorkflowVersionNodeId, workflowVersionNodeLog.CreatedOn)
	if err != nil {
		return WorkflowVersionNodeLog{}, errors.Wrap(err, database.SqlExecutionError)
	}
	return workflowVersionNodeLog, nil
}

func GetWorkflowVersionNodeLogs(db *sqlx.DB, workflowVersionNodeId int, maximumResultCount int) ([]WorkflowVersionNodeLog, error) {
	var wfvnls []WorkflowVersionNodeLog
	err := db.Get(&wfvnls,
		fmt.Sprintf(
			"SELECT * FROM workflow_version_node_log WHERE workflow_version_node_id=$1 ORDER BY created_on DESC LIMIT %d;", maximumResultCount),
		workflowVersionNodeId)
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
	err := db.Get(&wfvnls,
		fmt.Sprintf(
			"SELECT * "+
				"FROM workflow_version_node_log wfvnls "+
				"JOIN workflow_version_node wfvn ON wfvn.workflow_version_node_id=wfvnls.workflow_version_node_id "+
				"JOIN workflow_version wfv ON wfv.workflow_version_id=wfvn.workflow_version_id "+
				"WHERE wfv.workflow_id=$1 "+
				"ORDER BY created_on DESC "+
				"LIMIT %d;", maximumResultCount),
		workflowId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []WorkflowVersionNodeLog{}, nil
		}
		return nil, errors.Wrap(err, database.SqlExecutionError)
	}
	return wfvnls, nil
}
