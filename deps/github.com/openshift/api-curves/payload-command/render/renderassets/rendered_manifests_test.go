package assets

import (
	"testing"

	"k8s.io/apimachinery/pkg/runtime"
)

func TestRenderedManifest_GetDecodedObj(t *testing.T) {
	type fields struct {
		OriginalFilename string
		Content          []byte
		decodedObj       runtime.Object
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "yaml-decode",
			fields: fields{
				OriginalFilename: "featuregates.yaml",
				Content:          []byte(exampleFeatureGateYaml),
				decodedObj:       nil,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &RenderedManifest{
				OriginalFilename: tt.fields.OriginalFilename,
				Content:          tt.fields.Content,
				decodedObj:       tt.fields.decodedObj,
			}
			got, err := c.GetDecodedObj()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDecodedObj() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got == nil {
				t.Errorf("failed to get obj")
			}
		})
	}
}

const exampleFeatureGateYaml = `apiVersion: config.openshift.io/v1
kind: FeatureGate
metadata:
  creationTimestamp: "2023-11-14T16:11:49Z"
  generation: 1
  managedFields:
  - apiVersion: config.openshift.io/v1
    fieldsType: FieldsV1
    fieldsV1:
      f:spec: {}
    manager: cluster-bootstrap
    operation: Update
    time: "2023-11-14T16:11:49Z"
  - apiVersion: config.openshift.io/v1
    fieldsType: FieldsV1
    fieldsV1:
      f:status:
        .: {}
        f:featureGates:
          .: {}
          k:{"version":"4.15.0-0.ci.test-2023-11-14-160047-ci-op-tfcc12t7-latest"}:
            .: {}
            f:enabled: {}
            f:version: {}
    manager: cluster-bootstrap
    operation: Update
    subresource: status
    time: "2023-11-14T16:11:49Z"
  - apiVersion: config.openshift.io/v1
    fieldsType: FieldsV1
    fieldsV1:
      f:status:
        f:featureGates:
          k:{"version":"4.15.0-0.ci.test-2023-11-14-160047-ci-op-tfcc12t7-latest"}:
            f:disabled: {}
    manager: cluster-config-operator
    operation: Update
    subresource: status
    time: "2023-11-14T16:15:02Z"
  name: cluster
  resourceVersion: "5719"
  uid: e5e91e59-9e1a-4af4-bfa5-a81466b3356a
spec: {}
status:
  featureGates:
  - disabled:
    - name: AdminNetworkPolicy
    - name: AutomatedEtcdBackup
    - name: DNSNameResolver
    - name: DynamicResourceAllocation
    - name: EventedPLEG
    - name: GatewayAPI
    - name: InsightsConfigAPI
    - name: MachineAPIOperatorDisableMachineHealthCheckController
    - name: MachineAPIProviderOpenStack
    - name: MachineConfigNodes
    - name: MaxUnavailableStatefulSet
    - name: MetricsServer
    - name: RouteExternalCertificate
    - name: SigstoreImageVerification
    - name: ValidatingAdmissionPolicy
    enabled:
    - name: AlibabaPlatform
    - name: AzureWorkloadIdentity
    - name: BuildCSIVolumes
    - name: CloudDualStackNodeIPs
    - name: ExternalCloudProvider
    - name: ExternalCloudProviderAzure
    - name: ExternalCloudProviderExternal
    - name: ExternalCloudProviderGCP
    - name: KMSv1
    - name: OpenShiftPodSecurityAdmission
    - name: PrivateHostedZoneAWS
    version: 4.15.0-0.ci.test-2023-11-14-160047-ci-op-tfcc12t7-latest`
