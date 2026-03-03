package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/russross/blackfriday"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/openshift/api/tools/codegen/pkg/sippy"
	"github.com/openshift/api/tools/codegen/pkg/utils"
)

const (
	// all features should have at least this many tests
	requiredNumberOfTests = 5

	// all variant should run at least this many times
	requiredNumberOfTestRunsPerVariant = 14

	// required pass rate.
	// nearly all current tests pass 99% of the time, but in a two week window we lack enough data to say.
	requiredPassRateOfTestsPerVariant = 0.95
)

type FeatureGateTestAnalyzerOptions struct {
	In     io.Reader
	Out    io.Writer
	ErrOut io.Writer

	CurrentFeatureSetDir  string
	PreviousFeatureSetDir string
	OutputDir             string
}

func NewFeatureGateTestAnalyzerOptions(in io.Reader, out, errOut io.Writer) *FeatureGateTestAnalyzerOptions {
	return &FeatureGateTestAnalyzerOptions{
		In:                    in,
		Out:                   out,
		ErrOut:                errOut,
		CurrentFeatureSetDir:  filepath.Join("payload-manifests", "featuregates"),
		PreviousFeatureSetDir: filepath.Join("_tmp", "previous-openshift-api", "payload-manifests", "featuregates"),
	}
}

func NewFeatureGateTestAnalyzerFlagsCommand(in io.Reader, out, errOut io.Writer) *cobra.Command {
	o := NewFeatureGateTestAnalyzerOptions(in, out, errOut)

	cmd := &cobra.Command{
		Use:   "featuregate-test-analyzer",
		Short: "featuregate-test-analyzer looks to see how well tested a particular FeatureGate is.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancelFn := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancelFn()

			if err := o.Complete(); err != nil {
				return err
			}

			if err := o.Validate(); err != nil {
				return err
			}

			return o.Run(ctx)
		},
	}
	o.AddFlags(cmd.Flags())

	return cmd
}

func (o *FeatureGateTestAnalyzerOptions) Validate() error {
	if len(o.CurrentFeatureSetDir) == 0 {
		return fmt.Errorf("--featureset-manifest-path is required")
	}
	if len(o.PreviousFeatureSetDir) == 0 {
		return fmt.Errorf("--previous-featureset-manifest-path is required")
	}
	if _, err := os.ReadDir(o.CurrentFeatureSetDir); err != nil {
		return fmt.Errorf("--featureset-manifest-path cannot be read: %w", err)
	}
	if _, err := os.ReadDir(o.PreviousFeatureSetDir); err != nil {
		return fmt.Errorf("--previous-featureset-manifest-path cannot be read: %w", err)
	}

	return nil
}

func (o *FeatureGateTestAnalyzerOptions) Complete() error {
	artifactDir := os.Getenv("ARTIFACT_DIR")
	if len(artifactDir) > 0 {
		o.OutputDir = artifactDir
	}

	return nil
}

func (o *FeatureGateTestAnalyzerOptions) AddFlags(flags *pflag.FlagSet) {
	flags.StringVar(&o.CurrentFeatureSetDir, "featureset-manifest-path", o.CurrentFeatureSetDir, "path to directory containing the FeatureGate YAMLs for each FeatureGateTestAnalyzer,ClusterProfile tuple.")
	flags.StringVar(&o.PreviousFeatureSetDir, "previous-featureset-manifest-path", o.PreviousFeatureSetDir, "path to directory containing the OLD FeatureGate YAMLs for each FeatureGateTestAnalyzer,ClusterProfile tuple.")
}

func init() {
	rootCmd.AddCommand(NewFeatureGateTestAnalyzerFlagsCommand(os.Stdin, os.Stdout, os.Stderr))
}

func (o *FeatureGateTestAnalyzerOptions) Run(ctx context.Context) error {
	allCurrentClusterProfiles, _, _, currentByClusterProfileByFeatureSetTestAnalyzer, err := readFeatureGate(ctx, o.CurrentFeatureSetDir)
	if err != nil {
		return err
	}
	_, _, _, previousByClusterProfileByFeatureSetTestAnalyzer, err := readFeatureGate(ctx, o.PreviousFeatureSetDir)
	if err != nil {
		return err
	}

	md := utils.NewMarkdown("FeatureGate Promotion Summary")

	recentlyEnabledFeatureGatesToClusterProfiles := map[string]sets.Set[string]{}
	errs := []error{}
	for _, clusterProfile := range allCurrentClusterProfiles.List() {
		// we only need to check test coverage for current cluster profiles
		currentByFeatureSet := currentByClusterProfileByFeatureSetTestAnalyzer[clusterProfile]
		currentDefaultFeatureGateInfo := currentByFeatureSet["Default"]

		var previousDefaultFeatureGateInfo *featureGateInfo
		if previousByFeatureSet, ok := previousByClusterProfileByFeatureSetTestAnalyzer[clusterProfile]; ok {
			previousDefaultFeatureGateInfo = previousByFeatureSet["Default"]
		}

		currentFeatureGateNames := sets.StringKeySet(currentDefaultFeatureGateInfo.allFeatureGates)
		for _, featureGateName := range currentFeatureGateNames.List() {
			currentFeatureGateEnabled := currentDefaultFeatureGateInfo.allFeatureGates[featureGateName]
			if !currentFeatureGateEnabled {
				continue
			}

			previousFeatureGateEnabled := false
			if previousDefaultFeatureGateInfo != nil {
				previousFeatureGateEnabled = previousDefaultFeatureGateInfo.allFeatureGates[featureGateName]
			}
			if currentFeatureGateEnabled == previousFeatureGateEnabled {
				continue
			}

			// we've gone from false to true.
			if _, ok := recentlyEnabledFeatureGatesToClusterProfiles[featureGateName]; !ok {
				recentlyEnabledFeatureGatesToClusterProfiles[featureGateName] = sets.Set[string]{}
			}
			recentlyEnabledFeatureGatesToClusterProfiles[featureGateName].Insert(clusterProfile)
		}
	}

	if len(recentlyEnabledFeatureGatesToClusterProfiles) == 0 {
		md.Textf("No new Default FeatureGates found.\n")
		fmt.Fprintf(o.Out, "No new Default FeatureGates found.\n")
	}

	recentlyEnabledFeatureGates := sets.KeySet(recentlyEnabledFeatureGatesToClusterProfiles)
	for _, enabledFeatureGate := range sets.List(recentlyEnabledFeatureGates) {
		clusterProfiles := recentlyEnabledFeatureGatesToClusterProfiles[enabledFeatureGate]
		md.Title(1, enabledFeatureGate)

		testingResults, err := listTestResultFor(enabledFeatureGate, clusterProfiles)
		if err != nil {
			return err
		}

		writeTestingMarkDown(testingResults, md)

		currErrs := checkIfTestingIsSufficient(enabledFeatureGate, testingResults)
		if len(currErrs) == 0 {
			md.Textf("Sufficient CI testing for %q.\n", enabledFeatureGate)
			fmt.Fprintf(o.Out, "Sufficient CI testing for %q.\n", enabledFeatureGate)
		} else {
			md.Textf("INSUFFICIENT CI testing for %q.\n", enabledFeatureGate)
			md.Textf("* At least five tests are expected for a feature\n")
			md.Textf("* Tests must be be run on every TechPreview platform (ask for an exception if your feature doesn't support a variant)")
			md.Textf("* All tests must run at least 14 times on every platform")
			md.Textf("* All tests must pass at least 95%% of the time")
			md.Text("")
			md.Text("")
			fmt.Fprintf(o.Out, "INSUFFICIENT CI testing for %q.\n", enabledFeatureGate)
		}
		errs = append(errs, currErrs...)

	}

	summaryMarkdown := md.ExactBytes()
	if len(o.OutputDir) > 0 {
		filename := filepath.Join(o.OutputDir, "feature-promotion-summary.md")
		if err := os.WriteFile(filename, summaryMarkdown, 0644); err != nil {
			errs = append(errs, err)
		}

		htmlContent := blackfriday.Run(summaryMarkdown)
		htmlBytes := []byte{}
		htmlBytes = append(htmlBytes, []byte(htmlHeader)...)
		htmlBytes = append(htmlBytes, htmlContent...)
		htmlFilename := filepath.Join(o.OutputDir, "feature-promotion-summary.html")
		if err := os.WriteFile(htmlFilename, htmlBytes, 0644); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

const htmlHeader = `<head>
    <meta charset="UTF-8"><title>FeatureGate Promotion Summary</title>
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/bootstrap/4.6.1/css/bootstrap.min.css" integrity="sha512-T584yQ/tdRR5QwOpfvDfVQUidzfgc2339Lc8uBDtcp/wYu80d7jwBgAxbyMh0a9YM9F8N3tdErpFI8iaGx6x5g==" crossorigin="anonymous">
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/bootstrap-icons/1.5.0/font/bootstrap-icons.min.css">
    <meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">
    <style>
        @media (max-width: 992px) {
            .container {
                width: 100%;
                max-width: none;
            }
        }
        table, th, td {
            border: 1px solid;
            padding: 10px;
        }
    </style>
</head>

`

func checkIfTestingIsSufficient(featureGate string, testingResults map[JobVariant]*TestingResults) []error {
	errs := []error{}
	for jobVariant, testedVariant := range testingResults {
		if len(testedVariant.TestResults) < requiredNumberOfTests {
			errs = append(errs, fmt.Errorf("error: only %d tests found, need at least %d for %q on %v", len(testedVariant.TestResults), requiredNumberOfTests, featureGate, jobVariant))
		}
		for _, testResults := range testedVariant.TestResults {
			if testResults.TotalRuns < requiredNumberOfTestRunsPerVariant {
				errs = append(errs, fmt.Errorf("error: %q only has %d runs, need at least %d runs for %q on %v", testResults.TestName, testResults.TotalRuns, requiredNumberOfTestRunsPerVariant, featureGate, jobVariant))
			}
			if testResults.TotalRuns == 0 {
				continue
			}
			passPercent := float32(testResults.SuccessfulRuns) / float32(testResults.TotalRuns)
			if passPercent < requiredPassRateOfTestsPerVariant {
				displayExpected := int(requiredPassRateOfTestsPerVariant * 100)
				displayActual := int(passPercent * 100)
				errs = append(errs, fmt.Errorf("error: %q only passed %d%%, need at least %d%% for %q on %v", testResults.TestName, displayActual, displayExpected, featureGate, jobVariant))
			}
		}
	}

	return errs
}

func writeTestingMarkDown(testingResults map[JobVariant]*TestingResults, md *utils.Markdown) {
	jobVariantsSet := sets.KeySet(testingResults)
	jobVariants := jobVariantsSet.UnsortedList()
	sort.Sort(OrderedJobVariants(jobVariants))

	md.NextTableColumn()
	md.Exact("Test ")
	for _, jobVariant := range jobVariants {
		md.NextTableColumn()
		columnHeader := fmt.Sprintf("%v <br/> %v <br/> %v ", jobVariant.Topology, jobVariant.Cloud, jobVariant.Architecture)
		if jobVariant.NetworkStack != "" {
			columnHeader = columnHeader + fmt.Sprintf("<br/> %v ", jobVariant.NetworkStack)
		}
		md.Exact(columnHeader)
	}
	md.EndTableRow()
	md.NextTableColumn()
	md.Exact(":------ ")
	for i := 0; i < len(jobVariants); i++ {
		md.NextTableColumn()
		md.Exact(":---: ")
	}
	md.EndTableRow()

	allTests := sets.Set[string]{}
	for _, variantTestingResults := range testingResults {
		for _, currTestingResult := range variantTestingResults.TestResults {
			allTests.Insert(currTestingResult.TestName)
		}
	}

	for _, testName := range sets.List(allTests) {
		md.NextTableColumn()
		md.Exact(fmt.Sprintf("%s ", testName))

		for _, jobVariant := range jobVariants {
			md.NextTableColumn()
			allTesting := testingResults[jobVariant]
			testResults := testResultByName(allTesting.TestResults, testName)
			if testResults == nil {
				md.Exact(fmt.Sprintf("FAIL <br/> %d%% ( %d / %d ) ", 0, 0, 0))
				continue
			}
			failString := ""
			passPercent := float32(testResults.SuccessfulRuns) / float32(testResults.TotalRuns)
			switch {
			case testResults.TotalRuns < requiredNumberOfTestRunsPerVariant:
				failString = "FAIL <br/> "
			case passPercent < requiredPassRateOfTestsPerVariant:
				failString = "FAIL <br/> "
			}
			cellString := fmt.Sprintf("%s%d%% ( %d / %d ) ", failString, int(passPercent*100), testResults.SuccessfulRuns, testResults.TotalRuns)
			md.Exact(cellString)
		}

		md.EndTableRow()
	}
	md.Text("")
	md.Text("")

}

var (
	requiredSelfManagedJobVariants = []JobVariant{
		{
			Cloud:        "aws",
			Architecture: "amd64",
			Topology:     "ha",
		},
		{
			Cloud:        "azure",
			Architecture: "amd64",
			Topology:     "ha",
		},
		{
			Cloud:        "gcp",
			Architecture: "amd64",
			Topology:     "ha",
		},
		{
			Cloud:        "vsphere",
			Architecture: "amd64",
			Topology:     "ha",
		},
		{
			Cloud:        "metal",
			Architecture: "amd64",
			Topology:     "ha",
			NetworkStack: "ipv4",
		},
		{
			Cloud:        "metal",
			Architecture: "amd64",
			Topology:     "ha",
			NetworkStack: "ipv6",
		},
		{
			Cloud:        "metal",
			Architecture: "amd64",
			Topology:     "ha",
			NetworkStack: "dual",
		},
		{
			Cloud:        "aws",
			Architecture: "amd64",
			Topology:     "single",
		},

		// TODO restore these once we run TechPreview jobs that contain them
		//{
		//	Cloud:        "metal-ipi",
		//	Architecture: "amd64",
		//	Topology:     "single",
		//},
	}

	// These are only checked if the feature gate is platform specific
	optionalSelfManagedPlatformVariants = []JobVariant{
		{
			Cloud:        "nutanix",
			Architecture: "amd64",
			Topology:     "ha",
		},
		{
			Cloud:        "openstack",
			Architecture: "amd64",
			Topology:     "ha",
		},
		{
			Cloud:        "metal",
			Architecture: "amd64",
			Topology:     "two-node-arbiter",
			NetworkStack: "ipv4",
		},
		{
			Cloud:        "metal",
			Architecture: "amd64",
			Topology:     "two-node-arbiter",
			NetworkStack: "ipv6",
		},
		{
			Cloud:        "metal",
			Architecture: "amd64",
			Topology:     "two-node-arbiter",
			NetworkStack: "dual",
		},
		{
			Cloud:        "metal",
			Architecture: "amd64",
			Topology:     "two-node-fencing",
			NetworkStack: "ipv4",
		},
		{
			Cloud:        "metal",
			Architecture: "amd64",
			Topology:     "two-node-fencing",
			NetworkStack: "ipv6",
		},
		{
			Cloud:        "metal",
			Architecture: "amd64",
			Topology:     "two-node-fencing",
			NetworkStack: "dual",
		},
	}

	nonHypershiftPlatforms        = regexp.MustCompile("(?i)nutanix|metal|vsphere|openstack|azure|gcp")
	requiredHypershiftJobVariants = []JobVariant{
		{
			Cloud:        "aws",
			Architecture: "amd64",
			Topology:     "hypershift",
		},
		// ibm and powervs?
	}
)

type JobVariant struct {
	Cloud        string
	Architecture string
	Topology     string
	NetworkStack string
}

type OrderedJobVariants []JobVariant

func (a OrderedJobVariants) Len() int      { return len(a) }
func (a OrderedJobVariants) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a OrderedJobVariants) Less(i, j int) bool {
	if strings.Compare(a[i].Topology, a[j].Topology) < 0 {
		return true
	} else if strings.Compare(a[i].Topology, a[j].Topology) > 0 {
		return false
	}

	if strings.Compare(a[i].Cloud, a[j].Cloud) < 0 {
		return true
	} else if strings.Compare(a[i].Cloud, a[j].Cloud) > 0 {
		return false
	}

	if strings.Compare(a[i].Architecture, a[j].Architecture) < 0 {
		return true
	} else if strings.Compare(a[i].Architecture, a[j].Architecture) > 0 {
		return false
	}

	return false
}

type TestingResults struct {
	JobVariant JobVariant

	TestResults []TestResults
}

type TestResults struct {
	TestName       string
	TotalRuns      int
	SuccessfulRuns int
	FailedRuns     int
	FlakedRuns     int
}

func testResultByName(results []TestResults, testName string) *TestResults {
	for _, curr := range results {
		if curr.TestName == testName {
			return &curr
		}
	}
	return nil
}

func listTestResultFor(featureGate string, clusterProfiles sets.Set[string]) (map[JobVariant]*TestingResults, error) {
	fmt.Printf("Query sippy for all test run results for feature gate %q on clusterProfile %q\n", featureGate, sets.List(clusterProfiles))

	results := map[JobVariant]*TestingResults{}

	var jobVariantsToCheck []JobVariant
	if clusterProfiles.Has("Hypershift") && !nonHypershiftPlatforms.MatchString(featureGate) {
		jobVariantsToCheck = append(jobVariantsToCheck, filterVariants(featureGate, requiredHypershiftJobVariants)...)
	}
	if clusterProfiles.Has("SelfManagedHA") {
		// See if the feature gate is specific to any platform
		selfManagedPlatformVariants := filterVariants(featureGate, optionalSelfManagedPlatformVariants, requiredSelfManagedJobVariants)

		// If this isn't a platform specific variant, then check all required ones
		if len(selfManagedPlatformVariants) == 0 {
			selfManagedPlatformVariants = requiredSelfManagedJobVariants
		}

		jobVariantsToCheck = append(jobVariantsToCheck, selfManagedPlatformVariants...)
	}

	for _, jobVariant := range jobVariantsToCheck {
		jobVariantResults, err := listTestResultForVariant(featureGate, jobVariant)
		if err != nil {
			return nil, err
		}
		results[jobVariant] = jobVariantResults
	}

	return results, nil
}

func filterVariants(featureGate string, variantsList ...[]JobVariant) []JobVariant {
	var filteredVariants []JobVariant
	normalizedFeatureGate := strings.ToLower(featureGate)

	for _, variants := range variantsList {
		for _, variant := range variants {
			normalizedCloud := strings.ReplaceAll(strings.ToLower(variant.Cloud), "-ipi", "") // The feature gate probably won't include the install type, but some cloud variants do
			normalizedArchitecture := strings.ToLower(variant.Architecture)
			normalizedTopology := strings.ToLower(variant.Topology)

			if strings.Contains(normalizedFeatureGate, normalizedCloud) || strings.Contains(normalizedFeatureGate, normalizedArchitecture) || matchTwoNodeFeatureGates(normalizedFeatureGate, normalizedTopology) {
				filteredVariants = append(filteredVariants, variant)
			}
		}
	}

	return filteredVariants
}

// getLatestRelease returns the latest release from Sippy.
func getLatestRelease() (string, error) {
	releaseAPI := "https://sippy.dptools.openshift.org/api/releases"
	resp, err := http.Get(releaseAPI)
	if err != nil {
		return "", fmt.Errorf("error fetching data from API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %v", err)
	}

	var result struct {
		Releases []string `json:"releases"`
		Dates    map[string]struct {
			GA               *time.Time `json:"ga,omitempty"`
			DevelopmentStart *time.Time `json:"development_start,omitempty"`
		} `json:"dates"`
	}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return "", fmt.Errorf("error unmarshalling JSON: %v", err)
	}

	if len(result.Releases) == 0 {
		return "", fmt.Errorf("no releases found")
	}

	for _, release := range result.Releases {
		if dates, ok := result.Dates[release]; ok {
			if dates.DevelopmentStart != nil && !dates.DevelopmentStart.IsZero() && time.Now().After(*dates.DevelopmentStart) {
				return release, nil
			}
		}
	}

	return "", fmt.Errorf("no valid development releases found")
}

func listTestResultForVariant(featureGate string, jobVariant JobVariant) (*TestingResults, error) {
	// Substring here matches for both [OCPFeatureGate:...] and [FeatureGate:...]
	testPattern := fmt.Sprintf("FeatureGate:%s]", featureGate)

	// Feature gates used by the installer don't need separate tests, use the overall install tests
	if strings.Contains(featureGate, "Install") {
		testPattern = fmt.Sprintf("install should succeed")
	}

	fmt.Printf("Query sippy for all test run results for pattern %q on variant %#v\n", testPattern, jobVariant)

	defaultTransport := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	sippyClient := &http.Client{
		Timeout:   30 * time.Second,
		Transport: defaultTransport,
	}

	testNameToResults := map[string]*TestResults{}
	queries := sippy.QueriesFor(jobVariant.Cloud, jobVariant.Architecture, jobVariant.Topology, jobVariant.NetworkStack, testPattern)
	var release string
	// if its not main branch, then use the ENV var to determine the release version
	currentRelease := os.Getenv("PULL_BASE_REF")
	if strings.Contains(currentRelease, "release-") {
		// example: release-4.18, release-4.17
		release = strings.TrimPrefix(currentRelease, "release-")
	} else {
		// means its main branch
		var err error
		release, err = getLatestRelease()
		if err != nil {
			return nil, fmt.Errorf("couldn't fetch latest release version: %w", err)
		}
	}
	fmt.Printf("Querying sippy release %s for test run results\n", release)

	for _, currQuery := range queries {
		currURL := &url.URL{
			Scheme: "https",
			Host:   "sippy.dptools.openshift.org",
			Path:   "api/tests",
		}
		queryParams := currURL.Query()
		queryParams.Add("release", release)
		queryParams.Add("period", "default")
		filterJSON, err := json.Marshal(currQuery)
		if err != nil {
			return nil, err
		}
		queryParams.Add("filter", string(filterJSON))
		currURL.RawQuery = queryParams.Encode()

		req, err := http.NewRequest(http.MethodGet, currURL.String(), nil)
		if err != nil {
			return nil, err
		}

		response, err := sippyClient.Do(req)
		if err != nil {
			return nil, err
		}
		if response.StatusCode < 200 || response.StatusCode > 299 {
			return nil, fmt.Errorf("error getting sippy results (status=%d) for: %v", response.StatusCode, currURL.String())
		}
		queryResultBytes, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, err
		}
		response.Body.Close()

		testInfos := []sippy.SippyTestInfo{}
		if err := json.Unmarshal(queryResultBytes, &testInfos); err != nil {
			return nil, err
		}

		for _, currTest := range testInfos {
			testResults, ok := testNameToResults[currTest.Name]
			if !ok {
				testResults = &TestResults{
					TestName: currTest.Name,
				}
			}

			// Try to find enough test results in the last week, but if we have to we can extend
			// the window to two weeks.
			if currTest.CurrentRuns >= requiredNumberOfTestRunsPerVariant {
				testResults.TotalRuns = currTest.CurrentRuns
				testResults.SuccessfulRuns = currTest.CurrentSuccesses
				testResults.FailedRuns = currTest.CurrentFailures
				testResults.FlakedRuns = currTest.CurrentFlakes
			} else {
				fmt.Printf("Insufficient results in last 7 days, increasing lookback to 2 weeks...")
				testResults.TotalRuns += currTest.CurrentRuns + currTest.PreviousRuns
				testResults.SuccessfulRuns += currTest.CurrentSuccesses + currTest.PreviousSuccesses
				testResults.FailedRuns += currTest.CurrentFailures + currTest.PreviousFailures
				testResults.FlakedRuns += currTest.CurrentFlakes + currTest.PreviousFlakes
			}
			testNameToResults[currTest.Name] = testResults
		}
	}

	jobVariantResults := &TestingResults{
		JobVariant:  jobVariant,
		TestResults: nil,
	}
	testNames := sets.StringKeySet(testNameToResults)
	for _, testName := range testNames.List() {
		jobVariantResults.TestResults = append(jobVariantResults.TestResults, *testNameToResults[testName])
	}

	return jobVariantResults, nil
}

// Check for Arbiter and DualReplica or Fencing featureGates as these have special topologies
func matchTwoNodeFeatureGates(featureGate string, topology string) bool {
	if strings.Contains(featureGate, "arbiter") && strings.Contains(topology, "arbiter") {
		return true
	}
	if (strings.Contains(featureGate, "dualreplica") || strings.Contains(featureGate, "fencing")) && strings.Contains(topology, "fencing") {
		return true
	}
	return false
}
