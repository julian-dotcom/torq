package workflows

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"github.com/rs/zerolog/log"
	"golang.org/x/exp/slices"

	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"

	"github.com/lncapital/torq/internal/cache"
	"github.com/lncapital/torq/internal/core"
	"github.com/lncapital/torq/internal/database"
	"github.com/lncapital/torq/pkg/server_errors"
)

const workflowLogCount = 100

type ManualTriggerEvent struct {
	core.EventData
	WorkflowVersionNodeId int `json:"workflowVersionNodeId"`
}

func RegisterWorkflowRoutes(r *gin.RouterGroup, db *sqlx.DB) {

	// Workflows Crud (/api/workflows)
	r.GET("", func(c *gin.Context) { getWorkflowsHandler(c, db) })
	r.POST("", func(c *gin.Context) { createWorkflowHandler(c, db) })
	r.PUT("", func(c *gin.Context) { updateWorkflowHandler(c, db) })
	r.DELETE("/:workflowId", func(c *gin.Context) { removeWorkflowHandler(c, db) })
	r.POST("/trigger", func(c *gin.Context) { workFlowTriggerHandler(c, db) })

	// Workflow Logs
	r.GET("/logs/:workflowId", func(c *gin.Context) { getWorkflowLogsHandler(c, db) })

	wv := r.Group("/:workflowId/versions")
	{
		// Get all versions of a workflow
		wv.GET("", func(c *gin.Context) { getWorkflowVersionsHandler(c, db) })
		// Get a workflow version
		wv.GET("/:versionId", func(c *gin.Context) { getNodesHandler(c, db) })
		// Clone a workflow version (also used to simply add a new version)
		wv.POST("/clone", func(c *gin.Context) { cloneWorkflowVersionHandler(c, db) })
		wv.PUT("", func(c *gin.Context) { updateWorkflowVersionHandler(c, db) })
		wv.DELETE("/:versionId", func(c *gin.Context) { removeWorkflowVersionHandler(c, db) })
		// Delete a workflow Stage
		wv.DELETE("/:versionId/stage/:stage", func(c *gin.Context) { deleteStageHandler(c, db) })
	}

	// Add, update, delete nodes to a workflow version
	nodes := r.Group("/nodes")
	{
		nodes.POST("", func(c *gin.Context) { addNodeHandler(c, db) })
		nodes.PUT("", func(c *gin.Context) { updateNodeHandler(c, db) })
		nodes.PUT("/:nodeId/visibility-settings", func(c *gin.Context) { updateNodeLVisibilitySettingsHandler(c, db) })
		nodes.DELETE("/:nodeId", func(c *gin.Context) { removeNodeHandler(c, db) })

		// Workflow Node Logs
		nodes.GET("/:nodeId/logs", func(c *gin.Context) { getNodeLogsHandler(c, db) })
	}

	// Add, update, delete node links
	links := r.Group("/links")
	{
		links.POST("", func(c *gin.Context) { addNodeLinkHandler(c, db) })
		links.PUT("", func(c *gin.Context) { updateNodeLinkHandler(c, db) })
		links.DELETE("/:linkId", func(c *gin.Context) { removeNodeLinkHandler(c, db) })
	}

}

func getWorkflowsHandler(c *gin.Context, db *sqlx.DB) {
	workflows, err := GetWorkflows(db)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Getting workflows.")
		return
	}
	c.JSON(http.StatusOK, workflows)
}

func createWorkflowHandler(c *gin.Context, db *sqlx.DB) {

	storedWorkflow, err := createWorkflow(db)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Adding workflow.")
		return
	}

	wv, err := createWorkflowVersion(db, storedWorkflow.WorkflowId, Active)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Adding workflow version to new workflow.")
		return
	}

	initialNode := CreateNodeRequest{
		WorkflowVersionId: wv.WorkflowVersionId,
		Name:              "Trigger",
		Type:              core.WorkflowTrigger,
		Stage:             1,
	}
	_, err = createNode(db, initialNode)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Adding initial trigger to new workflow.")
		return
	}

	c.JSON(http.StatusOK, wv)
}

func workFlowTriggerHandler(c *gin.Context, db *sqlx.DB) {
	var workflow WorkflowToTrigger
	if err := c.BindJSON(&workflow); err != nil {
		server_errors.SendBadRequestFromError(c, errors.Wrap(err, server_errors.JsonParseError))
		return
	}
	manualTriggerEvent := ManualTriggerEvent{
		EventData: core.EventData{
			EventTime: time.Now(),
		},
		WorkflowVersionNodeId: workflow.WorkflowVersionNodeId,
	}
	reference := fmt.Sprintf("%v_%v", workflow.WorkflowVersionId, time.Now().UTC().Format("20060102.150405.000000"))
	cache.ScheduleTrigger(reference, workflow.WorkflowVersionId, core.WorkflowNodeManualTrigger,
		workflow.WorkflowVersionNodeId, manualTriggerEvent)

	c.JSON(http.StatusOK, map[string]interface{}{"message": "Successfully triggered Workflow."})
}

func updateWorkflowHandler(c *gin.Context, db *sqlx.DB) {
	var req UpdateWorkflow
	if err := c.BindJSON(&req); err != nil {
		server_errors.SendBadRequestFromError(c, errors.Wrap(err, server_errors.JsonParseError))
		return
	}

	if req.Status != nil && *req.Status == Active {
		workflowNodes, err := getWorkflowVersionNodesByWorkflow(db, req.WorkflowId)
		if err != nil {
			server_errors.WrapLogAndSendServerError(c, err,
				fmt.Sprintf("Validating workflow for workflowId: %v", req.WorkflowId))
			return
		}
		var rebalancerFocus RebalancerFocus
		for _, workflowNode := range workflowNodes {
			switch workflowNode.Type {
			case core.WorkflowNodeRebalanceConfigurator, core.WorkflowNodeRebalanceAutoRun:
				var rebalanceConfiguration RebalanceConfiguration
				err := json.Unmarshal([]byte(workflowNode.Parameters), &rebalanceConfiguration)
				if err != nil {
					server_errors.WrapLogAndSendServerError(c, err,
						fmt.Sprintf("Validating workflow for workflowId: %v", req.WorkflowId))
					return
				}
				if rebalancerFocus == "" {
					rebalancerFocus = rebalanceConfiguration.Focus
				} else if rebalancerFocus != rebalanceConfiguration.Focus {
					se := server_errors.SingleServerError(
						fmt.Sprintf("Rebalance configuration focus mismatch for %v", workflowNode.Name))
					server_errors.SendBadRequestFieldError(c, se)
					return
				}
			}
		}
	}
	if req.Status != nil && *req.Status != Active {
		wfvnIds, err := getWorkflowVersionNodeIdsByWorkflow(db, req.WorkflowId)
		if err != nil {
			log.Error().Err(err).Msgf(
				"Could not get the workflow version nodes to cancel the rebalances associated with it for workflowId: %v",
				req.WorkflowId)
		}
		cancelRebalancersByOriginIds(core.RebalanceRequestWorkflowNode, wfvnIds)
	}
	storedWorkflow, err := updateWorkflow(db, req)
	if err != nil {
		if errors.Is(err, database.SqlUniqueConstraintError) {
			se := server_errors.SingleFieldError("name", "Name already exists.")
			se.AddServerError(err.Error())
			server_errors.SendBadRequestFieldError(c, se)
			return
		}
		server_errors.WrapLogAndSendServerError(c, err, fmt.Sprintf("Setting workflow for workflowId: %v", req.WorkflowId))
		return
	}
	if req.Status != nil {
		workflowIds, err := GetWorkflowIdsByNodeType(db, core.WorkflowNodeCronTrigger)
		if err != nil {
			log.Error().Err(err).Msg("Could not obtain workflowIds for WorkflowNodeCronTrigger")
		}
		if slices.Contains(workflowIds, req.WorkflowId) {
			ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
			defer cancel()
			success := cache.InactivateCoreService(ctxWithTimeout, core.CronService)
			if success {
				success = cache.ActivateCoreService(ctxWithTimeout, core.CronService)
			}
			if !success {
				server_errors.WrapLogAndSendServerError(c, err, "Could not restart CronService.")
				return
			}
		}
	}

	c.JSON(http.StatusOK, storedWorkflow)
}

// TODO: update removeWorkflowHandler to remove a workflow and all of its versions, nodes and links.
//
//	At the moment it only removes the workflow and is not in use.
func removeWorkflowHandler(c *gin.Context, db *sqlx.DB) {
	workflowId, err := strconv.Atoi(c.Param("workflowId"))
	if err != nil {
		server_errors.SendBadRequest(c, "Failed to find/parse workflowId in the request.")
		return
	}
	count, err := removeWorkflow(db, workflowId)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, fmt.Sprintf("Removing workflow for workflowId: %v", workflowId))
		return
	}
	c.JSON(http.StatusOK, map[string]interface{}{"message": fmt.Sprintf("Successfully deleted %v workflow(s).", count)})
}

func deleteStageHandler(c *gin.Context, db *sqlx.DB) {

	workflowId, err := strconv.Atoi(c.Param("workflowId"))
	if err != nil {
		server_errors.SendBadRequest(c, "Failed to find/parse workflowId in the request.")
		return
	}
	versionId, err := strconv.Atoi(c.Param("versionId"))
	if err != nil {
		server_errors.SendBadRequest(c, "Failed to find/parse versionId in the request.")
		return
	}

	stage, err := strconv.Atoi(c.Param("stage"))
	if err != nil {
		server_errors.SendBadRequest(c, "Failed to find/parse stage in the request.")
		return
	}

	workflowVersion, err := GetWorkflowVersion(db, workflowId, versionId)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, fmt.Sprintf("Getting workflow version for workflowId: %v version %v", workflowId, versionId))
		return
	}

	err = updateStageDeleted(db, workflowVersion.WorkflowVersionId, stage)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, fmt.Sprintf("Deleting stage %v for workflowVersionId: %v", stage, workflowVersion.WorkflowVersionId))
		return
	}
	c.JSON(http.StatusOK, map[string]interface{}{"message": "Successfully deleted stage."})
}

//func getWorkflowVersionHandler(c *gin.Context, db *sqlx.DB) {
//	workflowId, err := strconv.Atoi(c.Param("workflowId"))
//	if err != nil {
//		server_errors.SendBadRequest(c, "Failed to find/parse workflowId in the request.")
//		return
//	}
//	versionId, err := strconv.Atoi(c.Param("versionId"))
//	if err != nil {
//		server_errors.SendBadRequest(c, "Failed to find/parse versionId in the request.")
//		return
//	}
//
//	workflowVersion, err := GetWorkflowVersion(db, versionId, workflowId)
//	if err != nil {
//		server_errors.WrapLogAndSendServerError(c, err, fmt.Sprintf("Getting workflow for workflowId: %v version %v", workflowId, versionId))
//		return
//	}
//	c.JSON(http.StatusOK, workflowVersion)
//}

func getWorkflowVersionsHandler(c *gin.Context, db *sqlx.DB) {
	workflowId, err := strconv.Atoi(c.Param("workflowId"))
	if err != nil {
		server_errors.SendBadRequest(c, "Failed to find/parse workflowId in the request.")
		return
	}
	workflowVersions, err := GetWorkflowVersions(db, workflowId)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Getting workflow versions.")
		return
	}
	c.JSON(http.StatusOK, workflowVersions)
}

func cloneWorkflowVersionHandler(c *gin.Context, db *sqlx.DB) {

	workflowId, err := strconv.Atoi(c.Param("workflowId"))
	if err != nil {
		server_errors.SendBadRequest(c, "Failed to find/parse workflowId in the request.")
		return
	}

	version, err := strconv.Atoi(c.Param("version"))
	if err != nil {
		server_errors.SendBadRequest(c, "Failed to find/parse version in the request.")
		return
	}

	storedWorkflowVersion, err := cloneWorkflowVersion(db, workflowId, &version)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Adding workflow version.")
		return
	}

	c.JSON(http.StatusOK, storedWorkflowVersion)
}

func updateWorkflowVersionHandler(c *gin.Context, db *sqlx.DB) {
	var wfv WorkflowVersion
	if err := c.BindJSON(&wfv); err != nil {
		server_errors.SendBadRequestFromError(c, errors.Wrap(err, server_errors.JsonParseError))
		return
	}
	storedWorkflowVersion, err := setWorkflowVersion(db, wfv)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, fmt.Sprintf("Setting workflow version for workflowVersionId: %v", wfv.WorkflowVersionId))
		return
	}

	c.JSON(http.StatusOK, storedWorkflowVersion)
}

func removeWorkflowVersionHandler(c *gin.Context, db *sqlx.DB) {
	workflowId, err := strconv.Atoi(c.Param("workflowId"))
	if err != nil {
		server_errors.SendBadRequest(c, "Failed to find/parse workflowId in the request.")
		return
	}

	versionId, err := strconv.Atoi(c.Param("versionId"))
	if err != nil {
		server_errors.SendBadRequest(c, "Failed to find/parse versionId in the request.")
		return
	}

	workflowVersion, err := GetWorkflowVersion(db, workflowId, versionId)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, fmt.Sprintf("Getting workflow version for workflowId: %v version %v", workflowId, versionId))
		return
	}

	count, err := removeWorkflowVersion(db, workflowVersion.WorkflowVersionId)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, fmt.Sprintf("Removing workflow version for workflow: %v version %v ", workflowId, versionId))
		return
	}
	c.JSON(http.StatusOK, map[string]interface{}{"message": fmt.Sprintf("Successfully deleted %v workflow version(s).", count)})
}

//func getWorkflowVersionNodeHandler(c *gin.Context, db *sqlx.DB) {
//	workflowVersionNodeId, err := strconv.Atoi(c.Param("workflowVersionNodeId"))
//	if err != nil {
//		server_errors.SendBadRequest(c, "Failed to find/parse workflowVersionNodeId in the request.")
//		return
//	}
//	workflowVersionNode, err := GetWorkflowVersionNode(db, workflowVersionNodeId)
//	if err != nil {
//		server_errors.WrapLogAndSendServerError(c, err, fmt.Sprintf("Getting workflow version for workflowVersionNodeId: %v", workflowVersionNodeId))
//		return
//	}
//	c.JSON(http.StatusOK, workflowVersionNode)
//}

func getNodesHandler(c *gin.Context, db *sqlx.DB) {
	workflowId, err := strconv.Atoi(c.Param("workflowId"))
	if err != nil {
		server_errors.SendBadRequest(c, "Failed to find/parse workflowId in the request.")
		return
	}
	versionId, err := strconv.Atoi(c.Param("versionId"))
	if err != nil {
		server_errors.SendBadRequest(c, "Failed to find/parse versionId in the request.")
		return
	}

	workflow, err := GetWorkflow(db, workflowId)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, fmt.Sprintf("Getting workflow for workflowId: %v", workflowId))
		return
	}

	workflowVersion, err := GetWorkflowVersion(db, workflowId, versionId)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, fmt.Sprintf("Getting workflow version for workflowId: %v version %v", workflowId, versionId))
		return
	}

	//workflowForest, err := GetWorkflowForest(db, workflowVersion.WorkflowVersionId)
	//if err != nil {
	//	server_errors.WrapLogAndSendServerError(c, err, "Getting workflow forest.")
	//	return
	//}

	nodes, err := GetWorkflowNodes(db, workflowVersion.WorkflowVersionId, workflow.WorkflowId, workflowVersion.Version)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Getting workflow nodes.")
		return
	}

	// get workflow version node links
	workflowVersionNodeLinks, err := GetWorkflowVersionNodeLinks(db, workflowVersion.WorkflowVersionId)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Getting workflow version node links.")
		return
	}

	r := WorkflowPage{
		Workflow: workflow,
		Version:  workflowVersion,
		Nodes:    nodes,
		Links:    workflowVersionNodeLinks,
	}

	c.JSON(http.StatusOK, r)
}

func addNodeHandler(c *gin.Context, db *sqlx.DB) {
	var req CreateNodeRequest
	if err := c.BindJSON(&req); err != nil {
		server_errors.SendBadRequestFromError(c, errors.Wrap(err, server_errors.JsonParseError))
		return
	}
	workflowVersion, err := GetWorkflowVersionById(db, req.WorkflowVersionId)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Get workflow version")
		return
	}
	workflow, err := GetWorkflow(db, workflowVersion.WorkflowId)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Get workflow")
		return
	}
	if !(workflowVersion.Status == Inactive || workflow.Status == Inactive) {
		server_errors.SendUnprocessableEntity(c, "Can't make changes to a workflow unless it's inactive")
		return
	}

	storedWorkflowVersionNode, err := createNode(db, req)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Adding workflow version node.")
		return
	}
	c.JSON(http.StatusOK, storedWorkflowVersionNode)
}

func updateNodeHandler(c *gin.Context, db *sqlx.DB) {

	//if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
	//	err := v.RegisterValidation("validParams", ValidateParams)
	//	if err != nil {
	//		return
	//	}
	//}

	var req UpdateNodeRequest
	if err := c.BindJSON(&req); err != nil {
		server_errors.SendBadRequestFromError(c, errors.Wrap(err, server_errors.JsonParseError))
		return
	}

	workflowNode, err := GetWorkflowVersionNode(db, req.WorkflowVersionNodeId)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Get workflow version node")
		return
	}
	workflowVersion, err := GetWorkflowVersionById(db, workflowNode.WorkflowVersionId)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Get workflow version")
		return
	}
	workflow, err := GetWorkflow(db, workflowVersion.WorkflowId)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Get workflow")
		return
	}
	if !(workflowVersion.Status == Inactive || workflow.Status == Inactive) {
		server_errors.SendUnprocessableEntity(c, "Can't make changes to a workflow unless it's inactive")
		return
	}

	// Validate the request
	resp, err := updateNode(db, req)
	if err != nil {
		pqErr, ok := err.(*pq.Error)
		if !ok {
			server_errors.WrapLogAndSendServerError(c, err, fmt.Sprintf("Setting workflow version node for workflowVersionNodeId: %v", req.WorkflowVersionNodeId))
		}
		switch true {
		case pqErr.Constraint == "workflow_version_node_workflow_version_id_name_key":
			se := server_errors.SingleFieldError("name", "Name already exists.")
			se.AddServerError(err.Error())
			server_errors.SendBadRequestFieldError(c, se)
		default:
			server_errors.WrapLogAndSendServerError(c, pqErr, fmt.Sprintf("Setting workflow version node for workflowVersionNodeId: %v", req.WorkflowVersionNodeId))
		}

		return
	}

	c.JSON(http.StatusOK, resp)
}

func updateNodeLVisibilitySettingsHandler(c *gin.Context, db *sqlx.DB) {
	workflowVersionNodeId, err := strconv.Atoi(c.Param("nodeId"))
	if err != nil {
		server_errors.SendBadRequest(c, "Failed to find/parse workflowVersionNodeId in the request.")
		return
	}

	var req WorkflowNodeVisibilitySettings
	if err := c.BindJSON(&req); err != nil {
		server_errors.SendBadRequestFromError(c, errors.Wrap(err, server_errors.JsonParseError))
		return
	}

	err = updateNodeVisibilitySettings(db, workflowVersionNodeId, req)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Update node visibility settings in the db")
	}

	c.JSON(http.StatusOK, workflowVersionNodeId)
}

func removeNodeHandler(c *gin.Context, db *sqlx.DB) {
	workflowVersionNodeId, err := strconv.Atoi(c.Param("nodeId"))
	if err != nil {
		server_errors.SendBadRequest(c, "Failed to find/parse workflowVersionNodeId in the request.")
		return
	}

	workflowNode, err := GetWorkflowVersionNode(db, workflowVersionNodeId)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Get workflow version node")
		return
	}
	workflowVersion, err := GetWorkflowVersionById(db, workflowNode.WorkflowVersionId)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Get workflow version")
		return
	}
	workflow, err := GetWorkflow(db, workflowVersion.WorkflowId)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Get workflow")
		return
	}
	if !(workflowVersion.Status == Inactive || workflow.Status == Inactive) {
		server_errors.SendUnprocessableEntity(c, "Can't make changes to a workflow unless it's inactive")
		return
	}

	err = updateNodeDeleted(db, workflowVersionNodeId)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, fmt.Sprintf("Removing workflow version node for workflowVersionNodeId: %v", workflowVersionNodeId))
		return
	}
	c.JSON(http.StatusOK, map[string]interface{}{"message": "Successfully deleted workflow version node."})
}

func addNodeLinkHandler(c *gin.Context, db *sqlx.DB) {
	var wfvnl CreateWorkflowVersionNodeLinkRequest
	if err := c.BindJSON(&wfvnl); err != nil {
		server_errors.SendBadRequestFromError(c, errors.Wrap(err, server_errors.JsonParseError))
		return
	}
	workflowVersion, err := GetWorkflowVersionById(db, wfvnl.WorkflowVersionId)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Get workflow version")
		return
	}
	workflow, err := GetWorkflow(db, workflowVersion.WorkflowId)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Get workflow")
		return
	}
	if !(workflowVersion.Status == Inactive || workflow.Status == Inactive) {
		server_errors.SendUnprocessableEntity(c, "Can't make changes to a workflow unless it's inactive")
		return
	}
	storedWorkflowVersionNodeLink, err := addWorkflowVersionNodeLink(db, wfvnl)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Adding workflow version node link.")
		return
	}
	c.JSON(http.StatusOK, storedWorkflowVersionNodeLink)
}

func updateNodeLinkHandler(c *gin.Context, db *sqlx.DB) {
	var wfvnl WorkflowVersionNodeLink
	if err := c.BindJSON(&wfvnl); err != nil {
		server_errors.SendBadRequestFromError(c, errors.Wrap(err, server_errors.JsonParseError))
		return
	}
	workflowVersion, err := GetWorkflowVersionById(db, wfvnl.WorkflowVersionId)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Get workflow version")
		return
	}
	workflow, err := GetWorkflow(db, workflowVersion.WorkflowId)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Get workflow")
		return
	}
	if !(workflowVersion.Status == Inactive || workflow.Status == Inactive) {
		server_errors.SendUnprocessableEntity(c, "Can't make changes to a workflow unless it's inactive")
		return
	}
	storedWorkflowVersionNodeLink, err := updateWorkflowVersionNodeLink(db, wfvnl)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, fmt.Sprintf("Setting workflow for WorkflowVersionNodeLinkId: %v", wfvnl.WorkflowVersionNodeLinkId))
		return
	}

	c.JSON(http.StatusOK, storedWorkflowVersionNodeLink)
}

func removeNodeLinkHandler(c *gin.Context, db *sqlx.DB) {
	workflowVersionNodeLinkId, err := strconv.Atoi(c.Param("linkId"))
	if err != nil {
		server_errors.SendBadRequest(c, "Failed to find/parse workflowVersionNodeLinkId in the request.")
		return
	}
	wfvnl, err := getWorkflowVersionNodeLink(db, workflowVersionNodeLinkId)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Get workflow version node link")
		return
	}
	workflowVersion, err := GetWorkflowVersionById(db, wfvnl.WorkflowVersionId)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Get workflow version")
		return
	}
	workflow, err := GetWorkflow(db, workflowVersion.WorkflowId)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Get workflow")
		return
	}
	if !(workflowVersion.Status == Inactive || workflow.Status == Inactive) {
		server_errors.SendUnprocessableEntity(c, "Can't make changes to a workflow unless it's inactive")
		return
	}
	count, err := removeNodeLink(db, workflowVersionNodeLinkId)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, fmt.Sprintf("Removing workflow version node link for workflowVersionNodeLinkId: %v", workflowVersionNodeLinkId))
		return
	}
	c.JSON(http.StatusOK, map[string]interface{}{"message": fmt.Sprintf("Successfully deleted %v workflow version node link(s).", count)})
}

func getWorkflowLogsHandler(c *gin.Context, db *sqlx.DB) {
	workflowId, err := strconv.Atoi(c.Param("workflowId"))
	if err != nil {
		server_errors.SendBadRequest(c, "Failed to find/parse workflowId in the request.")
		return
	}
	workflowLogs, err := GetWorkflowLogs(db, workflowId, workflowLogCount)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, fmt.Sprintf("Getting workflow logs for workflowId: %v", workflowId))
		return
	}
	c.JSON(http.StatusOK, workflowLogs)
}

func getNodeLogsHandler(c *gin.Context, db *sqlx.DB) {
	workflowVersionNodeId, err := strconv.Atoi(c.Param("nodeId"))
	if err != nil {
		server_errors.SendBadRequest(c, "Failed to find/parse workflowVersionNodeId in the request.")
		return
	}
	workflowVersionNodeLogs, err := GetNodeLogs(db, workflowVersionNodeId, workflowLogCount)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, fmt.Sprintf("Getting workflow version node logs for workflowVersionNodeId: %v", workflowVersionNodeId))
		return
	}
	c.JSON(http.StatusOK, workflowVersionNodeLogs)
}
