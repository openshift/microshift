package render

import (
	"flag"
	"fmt"
	configv1 "github.com/openshift/api/config/v1"
	"github.com/openshift/api/features"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"os"
	"path/filepath"
	"strings"
)

var (
	clusterProfileToShortName = map[features.ClusterProfileName]string{
		features.Hypershift:  "Hypershift",
		features.SelfManaged: "SelfManagedHA",
	}
)

// WriteFeatureSets holds values to drive the render command.
type WriteFeatureSets struct {
	PayloadVersion string
	AssetOutputDir string
}

func (o *WriteFeatureSets) AddFlags(fs *flag.FlagSet) {
	fs.StringVar(&o.PayloadVersion, "payload-version", o.PayloadVersion, "Version that will eventually be placed into ClusterOperator.status.  This normally comes from the CVO set via env var: OPERATOR_IMAGE_VERSION.")
	fs.StringVar(&o.AssetOutputDir, "asset-output-dir", o.AssetOutputDir, "Output path for rendered manifests.")
}

// Validate verifies the inputs.
func (o *WriteFeatureSets) Validate() error {
	return nil
}

// Complete fills in missing values before command execution.
func (o *WriteFeatureSets) Complete() error {
	return nil
}

// Run contains the logic of the render command.
func (o *WriteFeatureSets) Run() error {
	err := os.MkdirAll(o.AssetOutputDir, 0755)
	if err != nil {
		return err
	}

	statusByClusterProfileByFeatureSet := features.AllFeatureSets()
	newLegacyFeatureGates := sets.Set[string]{}
	for clusterProfile, byFeatureSet := range statusByClusterProfileByFeatureSet {
		for featureSetName, featureGateStatuses := range byFeatureSet {
			for _, curr := range featureGateStatuses.Enabled {
				if curr.EnhancementPR == "FeatureGate predates 4.18" {
					newLegacyFeatureGates.Insert(string(curr.FeatureGateAttributes.Name))
				}
			}
			for _, curr := range featureGateStatuses.Disabled {
				if curr.EnhancementPR == "FeatureGate predates 4.18" {
					newLegacyFeatureGates.Insert(string(curr.FeatureGateAttributes.Name))
				}
			}
			currentDetails := FeaturesGateDetailsFromFeatureSets(featureGateStatuses, o.PayloadVersion)

			featureGateInstance := &configv1.FeatureGate{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster",
					Annotations: map[string]string{
						// we can't do this because it will get the manifest included by the CVO and that isn't what we want
						// this makes it interesting to indicate which cluster-profile the cluster-config-operator should use
						//string(clusterProfile): "true",
						string(clusterProfile): "false-except-for-the-config-operator",
					},
				},
				Spec: configv1.FeatureGateSpec{
					FeatureGateSelection: configv1.FeatureGateSelection{
						FeatureSet: featureSetName,
					},
				},
				Status: configv1.FeatureGateStatus{
					FeatureGates: []configv1.FeatureGateDetails{
						*currentDetails,
					},
				},
			}

			featureGateOutBytes := writeFeatureGateV1OrDie(featureGateInstance)
			featureSetFileName := fmt.Sprintf("featureGate-%s-%s.yaml", clusterProfileToShortName[clusterProfile], featureSetName)
			if len(featureSetName) == 0 {
				featureSetFileName = fmt.Sprintf("featureGate-%s-%s.yaml", clusterProfileToShortName[clusterProfile], "Default")
			}

			destFile := filepath.Join(o.AssetOutputDir, featureSetFileName)
			if err := os.WriteFile(destFile, []byte(featureGateOutBytes), 0644); err != nil {
				return fmt.Errorf("error writing FeatureGate manifest: %w", err)
			}
		}
	}

	if illegalNewFeatureGates := newLegacyFeatureGates.Difference(legacyFeatureGates); len(illegalNewFeatureGates) > 0 {
		shameText := `If you are reading this, it is because you have tried to bypass the enhancement check for a new feature and caught by the backstop check.
Take the time to write up what you're going to accomplish and how we'll know it works in https://github.com/openshift/enhancements/.
If we don't know what we're trying to build and we don't know how to confirm it works as designed, we cannot expect to be successful delivering new features.
Complaints can be directed to @deads2k, approvers must not merge this PR.`
		err := fmt.Errorf(shameText+"\nFeatureGates: %v", strings.Join(illegalNewFeatureGates.UnsortedList(), ", "))
		return err

	}

	return nil
}
