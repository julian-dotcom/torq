package workflows

import (
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/pkg/errors"

	"github.com/lncapital/torq/internal/database"
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

// GetWorkflowVersionNode is not recursive and only returns direct parent/child relations without further nesting.
func GetWorkflowVersionNode(db *sqlx.DB, workflowVersionNodeId int) (WorkflowNodeStructured, error) {
	var wfvn WorkflowVersionNode
	err := db.Get(&wfvn, `SELECT * FROM workflow_version_node WHERE workflow_version_node_id=$1;`, workflowVersionNodeId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return WorkflowNodeStructured{}, nil
		}
		return WorkflowNodeStructured{}, errors.Wrap(err, database.SqlExecutionError)
	}

	parentNodes, err := getParentNodes(db, workflowVersionNodeId)
	if err != nil {
		return WorkflowNodeStructured{}, errors.Wrapf(err, "Obtaining parent nodes for workflowVersionNodeId: %v", workflowVersionNodeId)
	}

	childNodes, err := getChildNodes(db, workflowVersionNodeId)
	if err != nil {
		return WorkflowNodeStructured{}, errors.Wrapf(err, "Obtaining child nodes for workflowVersionNodeId: %v", workflowVersionNodeId)
	}

	response := wfvn.GetWorkflowNodeStructured()
	if len(parentNodes) > 0 {
		var parentNodesStructured []*WorkflowNodeStructured
		for _, parentNode := range parentNodes {
			parentNodeStructured := parentNode.GetWorkflowNodeStructured()
			parentNodesStructured = append(parentNodesStructured, &parentNodeStructured)
		}
		response.ParentNodes = parentNodesStructured
	}
	if len(childNodes) > 0 {
		var childNodesStructured []*WorkflowNodeStructured
		for _, childNode := range childNodes {
			childNodeStructured := childNode.GetWorkflowNodeStructured()
			childNodesStructured = append(childNodesStructured, &childNodeStructured)
		}
		response.ParentNodes = childNodesStructured
	}
	return response, nil
}

// GetWorkflowStructured this code assumes recursive loops from the root nodes are not allowed
func GetWorkflowStructured(db *sqlx.DB, workflowVersionId int) (WorkflowStructured, error) {
	processedNodes := make(map[int]*WorkflowNodeStructured)
	var rootNodes []WorkflowVersionNode
	err := db.Select(&rootNodes, `
		SELECT n.*
		FROM workflow_version_node n
		LEFT JOIN workflow_version_node_link parentLink ON parentLink.child_workflow_version_node_id = n.workflow_version_node_id
		WHERE n.workflow_version_id=$1 AND parentLink.child_workflow_version_node_id IS NULL;`, workflowVersionId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return WorkflowStructured{}, nil
		}
		return WorkflowStructured{}, errors.Wrap(err, database.SqlExecutionError)
	}

	var rootNodesStructured []*WorkflowNodeStructured
	if len(rootNodes) > 0 {
		for _, workflowNode := range rootNodes {
			workflowNodeStructured := workflowNode.GetWorkflowNodeStructured()
			workflowNodeStructured.ParentNodes = []*WorkflowNodeStructured{}
			err = processNodeRecursion(processedNodes, db, &workflowNodeStructured)
			if err != nil {
				return WorkflowStructured{}, err
			}
		}
	}
	return WorkflowStructured{RootNodes: rootNodesStructured}, nil
}

func processNodeRecursion(processedNodes map[int]*WorkflowNodeStructured, db *sqlx.DB, workflowNode *WorkflowNodeStructured) error {
	childNodes, err := getChildNodes(db, workflowNode.WorkflowVersionNodeId)
	if err != nil {
		return errors.Wrapf(err, "Obtaining child nodes for workflowVersionNodeId: %v", workflowNode.WorkflowVersionNodeId)
	}
	var childNodesStructured []*WorkflowNodeStructured
	for _, childNode := range childNodes {
		alreadyProcessedNode := processedNodes[childNode.WorkflowVersionNodeId]
		if alreadyProcessedNode != nil && alreadyProcessedNode.WorkflowVersionNodeId != 0 {
			alreadyProcessedNode.ParentNodes = append(alreadyProcessedNode.ParentNodes, workflowNode)
			childNodesStructured = append(childNodesStructured, alreadyProcessedNode)
		} else {
			childNodeStructured := childNode.GetWorkflowNodeStructured()
			childNodeStructured.ParentNodes = append(childNodeStructured.ParentNodes, workflowNode)
			childNodesStructured = append(childNodesStructured, &childNodeStructured)
			err = processNodeRecursion(processedNodes, db, &childNodeStructured)
			if err != nil {
				return errors.Wrapf(err, "Obtaining child nodes recursive for workflowVersionNodeId: %v", childNodeStructured.WorkflowVersionNodeId)
			}
		}
	}
	workflowNode.ChildNodes = childNodesStructured
	return nil
}

func getParentNodes(db *sqlx.DB, workflowVersionNodeId int) ([]WorkflowVersionNode, error) {
	var parentNodes []WorkflowVersionNode
	err := db.Select(&parentNodes, `
		SELECT n.*
		FROM workflow_version_node_link l
		JOIN workflow_version_node n ON n.workflow_version_node_id=l.child_workflow_version_node_id
		WHERE l.child_workflow_version_node_id=$1
		ORDER BY l.child_parameter_index;`, workflowVersionNodeId)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, errors.Wrap(err, database.SqlExecutionError)
		}
	}
	return parentNodes, nil
}

func getChildNodes(db *sqlx.DB, workflowVersionNodeId int) ([]WorkflowVersionNode, error) {
	var childNodes []WorkflowVersionNode
	err := db.Select(&childNodes, `
		SELECT n.*
		FROM workflow_version_node_link l
		JOIN workflow_version_node n ON n.workflow_version_node_id=l.parent_workflow_version_node_id
		WHERE l.parent_workflow_version_node_id=$1
		ORDER BY l.parent_parameter_index;`, workflowVersionNodeId)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, errors.Wrap(err, database.SqlExecutionError)
		}
	}
	return childNodes, nil
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
