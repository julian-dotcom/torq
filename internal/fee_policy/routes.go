package fee_policy

import (
	"net/http"
	"strconv"

	"github.com/cockroachdb/errors"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/lncapital/torq/pkg/server_errors"
)

func RegisterFeePolicyRoutes(r *gin.RouterGroup, db *sqlx.DB) {
	r.POST("", func(c *gin.Context) { addFeePolicyHandler(c, db) })
	r.PUT(":feePolicyId", func(c *gin.Context) { updateFeePolicyHandler(c, db) })
	r.GET("", func(c *gin.Context) { getFeePolicyListHandler(c, db) })
}

func updateFeePolicyHandler(c *gin.Context, db *sqlx.DB) {
	feePolicyId, err := strconv.Atoi(c.Param("feePolicyId"))
	if err != nil {
		server_errors.SendBadRequest(c, "Failed to find/parse statusId in the request.")
		return
	}

	var requestFeePolicy FeePolicy
	if err := c.BindJSON(&requestFeePolicy); err != nil {
		server_errors.SendBadRequestFromError(c, errors.Wrap(err, server_errors.JsonParseError))
		return
	}

	if err := validateFeePolicy(requestFeePolicy); err != nil {
		server_errors.SendUnprocessableEntity(c, err.Error())
		return
	}

	if requestFeePolicy.FeePolicyId != feePolicyId {
		server_errors.SendBadRequest(c, "Id used on route does not match id in request body")
		return
	}

	updatedFeePolicy, err := updateFeePolicy(db, requestFeePolicy)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Updating fee policy")
		return
	}

	c.JSON(http.StatusOK, updatedFeePolicy)
	return
}

func validateFeePolicy(feePolicy FeePolicy) error {
	if len(feePolicy.Targets) == 0 {
		return errors.New("Must specify at least one target")
	}

	if feePolicy.FeePolicyStrategy == policyStrategyStep && len(feePolicy.Steps) == 0 {
		return errors.New("Strategy of step specified but no steps provided")
	}

	if len(feePolicy.Name) == 0 {
		return errors.New("Name must be specified")
	}

	if feePolicy.Interval < 1 || feePolicy.Interval > 10080 {
		return errors.New("Interval must be between 1 and 10080 minutes although we don't recommend shorter than 10 minutes")
	}

	countOfNotNils := func(thingsToCheck ...*int) int {
		count := 0
		for _, thingToCheck := range thingsToCheck {
			if thingToCheck != nil {
				count++
			}
		}
		return count
	}

	for _, target := range feePolicy.Targets {
		if countOfNotNils(target.CategoryId, target.TagId, target.NodeId, target.ChannelId) != 1 {
			return errors.New("All targets must specify one of a Category, Tag, Node or Channel")
		}
	}

	if feePolicy.FeePolicyStrategy == policyStrategyStep {
		for _, step := range feePolicy.Steps {
			if countOfNotNils(step.FilterMaxBalance, step.FilterMinBalance, step.FilterMaxRatio, step.FilterMinRatio) == 0 {
				return errors.New("Balance and/or Ratio filter must be specified in all steps")
			}
			if step.SetFeePPM == 0 {
				return errors.New("Fee PPM can't be zero on any steps")
			}
		}
	}

	return nil
}

func addFeePolicyHandler(c *gin.Context, db *sqlx.DB) {
	var requestFeePolicy FeePolicy
	if err := c.BindJSON(&requestFeePolicy); err != nil {
		server_errors.SendBadRequestFromError(c, errors.Wrap(err, server_errors.JsonParseError))
		return
	}

	if err := validateFeePolicy(requestFeePolicy); err != nil {
		server_errors.SendUnprocessableEntity(c, err.Error())
		return
	}

	feePolicy, err := addFeePolicy(db, requestFeePolicy)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Inserting fee policy")
		return
	}

	c.JSON(http.StatusOK, feePolicy)
	return
}

func getFeePolicyListHandler(c *gin.Context, db *sqlx.DB) {
	feePolicies, err := getFeePolices(db)
	if err != nil {
		server_errors.WrapLogAndSendServerError(c, err, "Getting fee policies")
		return
	}
	c.JSON(http.StatusOK, feePolicies)
	return
}
