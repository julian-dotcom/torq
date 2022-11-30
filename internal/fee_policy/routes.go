package fee_policy

import (
	"net/http"

	"github.com/cockroachdb/errors"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/lncapital/torq/pkg/server_errors"
)

func RegisterFeePolicyRoutes(r *gin.RouterGroup, db *sqlx.DB) {
	r.POST("", func(c *gin.Context) { addFeePolicyHandler(c, db) })
	r.GET("", func(c *gin.Context) { getFeePolicyListHandler(c, db) })
}

func addFeePolicyHandler(c *gin.Context, db *sqlx.DB) {
	var requestFeePolicy FeePolicy
	if err := c.BindJSON(&requestFeePolicy); err != nil {
		server_errors.SendBadRequestFromError(c, errors.Wrap(err, server_errors.JsonParseError))
		return
	}

	if len(requestFeePolicy.Targets) == 0 {
		server_errors.SendUnprocessableEntity(c, "Must specify at least one target")
		return
	}

	if requestFeePolicy.FeePolicyStrategy == policyStrategyStep && len(requestFeePolicy.Steps) == 0 {
		server_errors.SendUnprocessableEntity(c, "Strategy of step specified but no steps provided")
		return
	}

	if len(requestFeePolicy.Name) == 0 {
		server_errors.SendUnprocessableEntity(c, "Name must be specified")
		return
	}

	if requestFeePolicy.Interval < 1 || requestFeePolicy.Interval > 10080 {
		server_errors.SendUnprocessableEntity(c, "Interval must be between 1 and 10080 minutes although we don't recommend shorter than 10 minutes")
		return
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

	for _, target := range requestFeePolicy.Targets {
		if countOfNotNils(target.CategoryId, target.TagId, target.NodeId, target.ChannelId) != 1 {
			server_errors.SendUnprocessableEntity(c, "All targets must specify one of a Category, Tag, Node or Channel")
			return
		}
	}

	if requestFeePolicy.FeePolicyStrategy == policyStrategyStep {
		for _, step := range requestFeePolicy.Steps {
			if countOfNotNils(step.FilterMaxBalance, step.FilterMinBalance, step.FilterMaxRatio, step.FilterMinRatio) == 0 {
				server_errors.SendUnprocessableEntity(c, "Balance and/or Ratio filter must be specified in all steps")
				return
			}
			if step.SetFeePPM == 0 {
				server_errors.SendUnprocessableEntity(c, "Fee PPM can't be zero on any steps")
				return
			}
		}
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
