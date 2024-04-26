package audit

import (
	"bytes"
	"embed"
	"fmt"
	"path"
	"strings"

	configv1 "github.com/openshift/api/config/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	auditv1 "k8s.io/apiserver/pkg/apis/audit/v1"
	"sigs.k8s.io/yaml"
)

//go:embed manifests
var assets embed.FS

var (
	basePolicy   auditv1.Policy
	profileRules = map[configv1.AuditProfileType][]auditv1.PolicyRule{}

	auditScheme         = runtime.NewScheme()
	auditCodecs         = serializer.NewCodecFactory(auditScheme)
	auditYamlSerializer = json.NewYAMLSerializer(json.DefaultMetaFactory, auditScheme, auditScheme)

	coreScheme         = runtime.NewScheme()
	coreCodecs         = serializer.NewCodecFactory(coreScheme)
	coreYamlSerializer = json.NewYAMLSerializer(json.DefaultMetaFactory, coreScheme, coreScheme)
)

func init() {
	if err := auditv1.AddToScheme(auditScheme); err != nil {
		panic(err)
	}
	if err := corev1.AddToScheme(coreScheme); err != nil {
		panic(err)
	}

	bs, err := assets.ReadFile("manifests/base-policy.yaml")
	if err != nil {
		panic(err)
	}
	if err := runtime.DecodeInto(coreCodecs.UniversalDecoder(auditv1.SchemeGroupVersion), bs, &basePolicy); err != nil {
		panic(err)
	}

	for _, profile := range []configv1.AuditProfileType{
		configv1.NoneAuditProfileType,
		configv1.DefaultAuditProfileType,
		configv1.WriteRequestBodiesAuditProfileType,
		configv1.AllRequestBodiesAuditProfileType,
	} {
		manifestName := fmt.Sprintf("%s-rules.yaml", strings.ToLower(string(profile)))
		bs, err := assets.ReadFile(path.Join("manifests", manifestName))
		if err != nil {
			panic(err)
		}
		var rules []auditv1.PolicyRule
		if err := yaml.Unmarshal(bs, &rules); err != nil {
			panic(err)
		}
		profileRules[profile] = rules
	}
}

// DefaultPolicy brings back the default.yaml audit policy to init the api
func DefaultPolicy() ([]byte, error) {
	policy, err := GetAuditPolicy(configv1.Audit{Profile: configv1.DefaultAuditProfileType})
	if err != nil {
		return nil, fmt.Errorf("failed to retreive default audit policy: %v", err)
	}

	policy.Kind = "Policy"
	policy.APIVersion = auditv1.SchemeGroupVersion.String()

	var buf bytes.Buffer
	if err := auditYamlSerializer.Encode(policy, &buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// GetAuditPolicy computes the audit policy for the given audit config.
// Note: the returned policy has Kind and APIVersion not set. This is responsibility of the caller
//
//	when serializing it.
//
// Note: the returned policy must not be modifed by the caller prior to a deepcopy.
func GetAuditPolicy(audit configv1.Audit) (*auditv1.Policy, error) {
	p := basePolicy.DeepCopy()
	p.Name = "policy"

	for _, cr := range audit.CustomRules {
		rules, ok := profileRules[cr.Profile]
		if !ok {
			return nil, fmt.Errorf("unknown audit profile %q in customRules for group %q", cr.Profile, cr.Group)
		}

		groupRules := make([]auditv1.PolicyRule, len(rules))
		for i, r := range rules {
			r.DeepCopyInto(&groupRules[i])
			groupRules[i].UserGroups = []string{cr.Group}
		}

		p.Rules = append(p.Rules, groupRules...)
	}

	globalRules, ok := profileRules[audit.Profile]
	if !ok {
		return nil, fmt.Errorf("unknown audit profile %q", audit.Profile)
	}
	p.Rules = append(p.Rules, globalRules...)

	return p, nil
}
