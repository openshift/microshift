package e2e

//go:generate go run -mod vendor ./annotate/cmd -- ./annotate/generated/zz_generated.annotations.go

// This file duplicates most of test/e2e/e2e_test.go but limits the included
// tests (via include.go) to tests that are relevant to openshift.

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	et "github.com/openshift-eng/openshift-tests-extension/pkg/extension/extensiontests"
	"gopkg.in/yaml.v2"

	// Never, ever remove the line with "/ginkgo". Without it,
	// the ginkgo test runner will not detect that this
	// directory contains a Ginkgo test suite.
	// See https://github.com/kubernetes/kubernetes/issues/74827
	// "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/types"

	corev1 "k8s.io/api/core/v1"
	kclientset "k8s.io/client-go/kubernetes"
	"k8s.io/component-base/version"
	conformancetestdata "k8s.io/kubernetes/test/conformance/testdata"
	"k8s.io/kubernetes/test/e2e"
	"k8s.io/kubernetes/test/e2e/framework"
	"k8s.io/kubernetes/test/e2e/framework/testfiles"
	e2etestingmanifests "k8s.io/kubernetes/test/e2e/testing-manifests"
	testfixtures "k8s.io/kubernetes/test/fixtures"
	"k8s.io/kubernetes/test/utils/image"
)

func TestMain(m *testing.M) {
	var versionFlag bool
	flag.CommandLine.BoolVar(&versionFlag, "version", false, "Displays version information.")

	// Register test flags, then parse flags.
	e2e.HandleFlags()

	if framework.TestContext.ListImages {
		for _, v := range image.GetImageConfigs() {
			fmt.Println(v.GetE2EImage())
		}
		os.Exit(0)
	}
	if versionFlag {
		fmt.Printf("%s\n", version.Get())
		os.Exit(0)
	}

	// Enable embedded FS file lookup as fallback
	testfiles.AddFileSource(e2etestingmanifests.GetE2ETestingManifestsFS())
	testfiles.AddFileSource(testfixtures.GetTestFixturesFS())
	testfiles.AddFileSource(conformancetestdata.GetConformanceTestdataFS())

	if framework.TestContext.ListConformanceTests {
		var tests []struct {
			Testname    string `yaml:"testname"`
			Codename    string `yaml:"codename"`
			Description string `yaml:"description"`
			Release     string `yaml:"release"`
			File        string `yaml:"file"`
		}

		data, err := testfiles.Read("test/conformance/testdata/conformance.yaml")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if err := yaml.Unmarshal(data, &tests); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if err := yaml.NewEncoder(os.Stdout).Encode(tests); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	// Ensure the test namespaces have disabled SCCs and label syncer.
	framework.TestContext.CreateTestingNS = func(ctx context.Context, baseName string, c kclientset.Interface, labels map[string]string) (*corev1.Namespace, error) {
		return CreateTestingNS(ctx, baseName, c, labels, true)
	}

	framework.AfterReadingAllFlags(&framework.TestContext)

	// TODO: Deprecating repo-root over time... instead just use gobindata_util.go , see #23987.
	// Right now it is still needed, for example by
	// test/e2e/framework/ingress/ingress_utils.go
	// for providing the optional secret.yaml file and by
	// test/e2e/framework/util.go for cluster/log-dump.
	if framework.TestContext.RepoRoot != "" {
		testfiles.AddFileSource(testfiles.RootFileSource{Root: framework.TestContext.RepoRoot})
	}

	rand.Seed(time.Now().UnixNano())
	os.Exit(m.Run())
}

func TestE2E(t *testing.T) {
	// In order to properly skip tests, we must add the labels that the OTE external binary supplies to the test name
	// This will then be used by Ginkgo to skip specific tests
	oteCmd := exec.Command("k8s-tests-ext", "list", "tests")
	// We can't have OTE also add annotations to the spec names to map to labels, or they won't match the actual spec names
	//TODO(sgoeddel): once annotation logic is removed, this can be as well
	oteCmd.Env = append(oteCmd.Env, "OMIT_ANNOTATIONS=true")
	oteCmd.Stderr = os.Stderr
	output, err := oteCmd.Output()
	if err != nil {
		t.Fatalf("Error running ote list tests command: %v", err)
	}
	var specs et.ExtensionTestSpecs
	if err = json.Unmarshal(output, &specs); err != nil {
		t.Fatalf("Error parsing ote list tests output: %v", err)
	}

	nameToLabels := make(map[string][]string, len(specs))
	for _, spec := range specs {
		nameToLabels[spec.Name] = spec.Labels.UnsortedList()
	}

	ginkgo.GetSuite().SetAnnotateFn(func(name string, node types.TestSpec) {
		if newLabels, ok := nameToLabels[name]; ok {
			for _, label := range newLabels {
				// Only add the label to the name if it isn't already present to avoid test names that are too long
				if !strings.Contains(name, label) {
					node.AppendText(fmt.Sprintf(" %s", label))
				}
			}
		} else {
			// If the name isn't found in the mapping, it is because the test has been disabled via OTE
			node.AppendText(" [Disabled:missing]")
		}
		if strings.Contains(name, "Kubectl client Kubectl prune with applyset should apply and prune objects") {
			fmt.Printf("Trying to annotate %q\n", name)
		}

	})

	e2e.RunE2ETests(t)
}
