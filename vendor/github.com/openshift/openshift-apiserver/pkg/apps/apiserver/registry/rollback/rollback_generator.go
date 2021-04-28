package rollback

import (
	appsapi "github.com/openshift/openshift-apiserver/pkg/apps/apis/apps"
)

// RollbackGenerator generates a new deployment config by merging a pair of deployment
// configs in a configurable way.
type RollbackGenerator interface {
	// GenerateRollback creates a new deployment config by merging to onto from
	// based on the options provided by spec. The latestVersion of the result is
	// unconditionally incremented, as rollback candidates should be possible
	// to be deployed manually regardless of other system behavior such as
	// triggering.
	//
	// Any image change triggers on the new config are disabled to prevent
	// triggered deployments from immediately replacing the rollback.
	GenerateRollback(from, to *appsapi.DeploymentConfig, spec *appsapi.DeploymentConfigRollbackSpec) (*appsapi.DeploymentConfig, error)
}

// NewRollbackGenerator returns a new rollback generator.
func NewRollbackGenerator() RollbackGenerator {
	return &rollbackGenerator{}
}

type rollbackGenerator struct{}

func (g *rollbackGenerator) GenerateRollback(from, to *appsapi.DeploymentConfig, spec *appsapi.DeploymentConfigRollbackSpec) (*appsapi.DeploymentConfig, error) {
	rollback := from.DeepCopy()

	// construct the candidate deploymentConfig based on the rollback spec
	if spec.IncludeTemplate {
		rollback.Spec.Template = to.Spec.Template.DeepCopy()
	}

	if spec.IncludeReplicationMeta {
		rollback.Spec.Replicas = to.Spec.Replicas
		rollback.Spec.Selector = map[string]string{}
		for k, v := range to.Spec.Selector {
			rollback.Spec.Selector[k] = v
		}
	}

	if spec.IncludeTriggers {
		rollback.Spec.Triggers = []appsapi.DeploymentTriggerPolicy{}
		for _, curr := range to.Spec.Triggers {
			rollback.Spec.Triggers = append(rollback.Spec.Triggers, *curr.DeepCopy())
		}
	}

	if spec.IncludeStrategy {
		rollback.Spec.Strategy = *to.Spec.Strategy.DeepCopy()
	}

	// Disable any image change triggers.
	for _, trigger := range rollback.Spec.Triggers {
		if trigger.Type == appsapi.DeploymentTriggerOnImageChange {
			trigger.ImageChangeParams.Automatic = false
		}
	}

	// TODO: add a new cause?
	// TODO: Instantiate instead of incrementing latestVersion
	rollback.Status.LatestVersion++

	return rollback, nil
}
