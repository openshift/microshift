package apiserver

import (
	auditV1 "k8s.io/apiserver/pkg/apis/audit/v1"

	configV1 "github.com/openshift/api/config/v1"
	"github.com/openshift/library-go/pkg/operator/apiserver/audit"
)

func GetPolicy(forProfile string) (*auditV1.Policy, error) {
	ac := configV1.Audit{
		Profile:     configV1.AuditProfileType(forProfile),
		CustomRules: nil,
	}

	ap, err := audit.GetAuditPolicy(ac)
	if err != nil {
		return nil, err
	}
	return ap.DeepCopy(), nil
}
