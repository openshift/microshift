package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/google/go-cmp/cmp"
	"github.com/openshift/api/tools/codegen/pkg/utils"
	"github.com/spf13/pflag"
	"io"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/sets"
	"os"
	"path/filepath"
	kyaml "sigs.k8s.io/yaml"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

type FeatureSetOptions struct {
	In     io.Reader
	Out    io.Writer
	ErrOut io.Writer

	FeatureSetManifestDir string
	OutputFile            string
	Verify                bool
}

func NewFeatureSetOptions(in io.Reader, out, errOut io.Writer) *FeatureSetOptions {
	return &FeatureSetOptions{
		In:                    in,
		Out:                   out,
		ErrOut:                errOut,
		FeatureSetManifestDir: filepath.Join("payload-manifests", "featuregates"),
		OutputFile:            "features.md",
	}
}

func NewFeatureSetFlagsCommand(in io.Reader, out, errOut io.Writer) *cobra.Command {
	o := NewFeatureSetOptions(in, out, errOut)

	cmd := &cobra.Command{
		Use:   "featureset-markdown",
		Short: "featureset-markdown generates a markdown document summarizing current featuregate status.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancelFn := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancelFn()

			if err := o.Validate(); err != nil {
				return err
			}

			return o.Run(ctx)
		},
	}
	o.AddFlags(cmd.Flags())

	return cmd
}

func (o *FeatureSetOptions) Validate() error {
	if len(o.FeatureSetManifestDir) == 0 {
		return fmt.Errorf("--featureset-manifest-path is required")
	}
	if _, err := os.ReadDir(o.FeatureSetManifestDir); err != nil {
		return fmt.Errorf("--featureset-manifest-path cannot be read: %w", err)
	}
	if len(o.OutputFile) == 0 {
		return fmt.Errorf("--output-file is required")
	}

	return nil
}

func (o *FeatureSetOptions) AddFlags(flags *pflag.FlagSet) {
	flags.StringVar(&o.FeatureSetManifestDir, "featureset-manifest-path", o.FeatureSetManifestDir, "path to directory containing the FeatureGate YAMLs for each FeatureSet,ClusterProfile tuple.")
	flags.StringVar(&o.OutputFile, "output-file", o.OutputFile, "path to markdown file detailing FeatureGates.")
	flags.BoolVar(&o.Verify, "verify", o.Verify, "Verify the content has not changed.")
}

func init() {
	rootCmd.AddCommand(NewFeatureSetFlagsCommand(os.Stdin, os.Stdout, os.Stderr))
}

func (o *FeatureSetOptions) Run(ctx context.Context) error {
	allClusterProfiles, allFeatureSets, _, byClusterProfilebyFeatureSet, err := readFeatureGate(ctx, o.FeatureSetManifestDir)
	if err != nil {
		return err
	}

	cols := []columnTuple{}
	md := utils.NewMarkdown("FeatureGate Summary")
	md.NextTableColumn()
	md.Exact("FeatureGate ")
	for _, featureSet := range allFeatureSets.List() {
		for _, clusterProfile := range allClusterProfiles.List() {
			cols = append(cols, columnTuple{
				clusterProfile: clusterProfile,
				featureSet:     featureSet,
			})
			md.NextTableColumn()
			md.Exact(fmt.Sprintf("%v on %v ", featureSet, clusterProfile))
		}
	}
	md.EndTableRow()
	md.NextTableColumn()
	md.Exact("------ ")
	for i := 0; i < len(cols); i++ {
		md.NextTableColumn()
		md.Exact("--- ")
	}
	md.EndTableRow()

	orderedFeatureGates := getOrderedFeatureGates(byClusterProfilebyFeatureSet)
	for _, featureGate := range orderedFeatureGates {
		md.NextTableColumn()
		md.Exact(featureGate)
		for _, col := range cols {
			currFeatureGateInfo := byClusterProfilebyFeatureSet[col.clusterProfile][col.featureSet]
			md.NextTableColumn()
			if currFeatureGateInfo.enabled.Has(featureGate) {
				md.Exact("<span style=\"background-color: #519450\">Enabled</span> ")
			} else {
				//md.Exact(" ")
			}
		}
		md.EndTableRow()
	}

	if o.Verify {
		actualContent, err := os.ReadFile(o.OutputFile)
		if err != nil {
			return fmt.Errorf("failed to verify: %w", err)
		}
		expectedContent := md.ExactBytes()
		if bytes.Equal(actualContent, expectedContent) {
			return nil
		}
		return fmt.Errorf("actual content not match: %v", cmp.Diff(expectedContent, actualContent))
	}

	if err := os.WriteFile(o.OutputFile, md.ExactBytes(), 0644); err != nil {
		return err
	}

	return nil
}

func getOrderedFeatureGates(info map[string]map[string]*featureGateInfo) []string {
	counts := map[string]int{}
	for _, byClusterProfile := range info {
		for _, byFeature := range byClusterProfile {
			for _, featureGate := range byFeature.enabled.List() {
				counts[featureGate] = counts[featureGate] + 1
			}
			for _, featureGate := range byFeature.disabled.List() {
				counts[featureGate] = counts[featureGate] + 0
			}
		}
	}

	toSort := []stringCount{}
	for name, count := range counts {
		toSort = append(toSort, stringCount{
			name:  name,
			count: count,
		})
	}

	sort.Sort(byCount(toSort))
	ret := []string{}
	for _, curr := range toSort {
		ret = append(ret, curr.name)
	}

	return ret
}

type stringCount struct {
	name  string
	count int
}
type byCount []stringCount

func (a byCount) Len() int      { return len(a) }
func (a byCount) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a byCount) Less(i, j int) bool {
	if a[i].count < a[j].count {
		return true
	}
	if a[i].count > a[j].count {
		return false
	}
	if strings.Compare(a[i].name, a[j].name) < 0 {
		return true
	}
	return false
}

type columnTuple struct {
	clusterProfile string
	featureSet     string
}

type featureGateInfo struct {
	clusterProfile string
	featureSet     string

	enabled         sets.String
	disabled        sets.String
	allFeatureGates map[string]bool
}

func readFeatureGate(ctx context.Context, featureSetManifestDir string) (sets.String, sets.String, sets.String, map[string]map[string]*featureGateInfo, error) {
	allClusterProfiles := sets.String{}
	allFeatureSets := sets.String{}
	allFeatureGates := sets.String{}
	clusterProfileToFeatureSetToFeatureGates := map[string]map[string]*featureGateInfo{}

	featureSetManifestFile, err := os.ReadDir(featureSetManifestDir)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("cannot read FeatureSetManifestDir: %w", err)
	}
	for _, currFeatureSetManifestFile := range featureSetManifestFile {
		currFeatureGateInfo := &featureGateInfo{
			enabled:         sets.String{},
			disabled:        sets.String{},
			allFeatureGates: map[string]bool{},
		}

		featureGateFilename := filepath.Join(featureSetManifestDir, currFeatureSetManifestFile.Name())
		featureGateBytes, err := os.ReadFile(featureGateFilename)
		if err != nil {
			return nil, nil, nil, nil, fmt.Errorf("unable to read %q: %w", featureGateFilename, err)
		}

		// use unstructured to pull this information to avoid vendoring openshift/api
		featureGateMap := map[string]interface{}{}
		if err := kyaml.Unmarshal(featureGateBytes, &featureGateMap); err != nil {
			return nil, nil, nil, nil, fmt.Errorf("unable to parse featuregate %q: %w", featureGateFilename, err)
		}
		uncastFeatureGate := unstructured.Unstructured{
			Object: featureGateMap,
		}

		clusterProfiles := clusterOperatorClusterProfilesFrom(uncastFeatureGate.GetAnnotations())
		if len(clusterProfiles) != 1 {
			return nil, nil, nil, nil, fmt.Errorf("expected exactly one clusterProfile from %q: %v", featureGateFilename, clusterProfiles.List())
		}

		clusterProfileShortName, err := utils.ClusterProfileToShortName(clusterProfiles.List()[0])
		if err != nil {
			return nil, nil, nil, nil, fmt.Errorf("unrecognized clusterprofile name %q: %w", clusterProfiles.List()[0], err)
		}
		currFeatureGateInfo.clusterProfile = clusterProfileShortName
		allClusterProfiles.Insert(currFeatureGateInfo.clusterProfile)

		currFeatureGateInfo.featureSet, _, _ = unstructured.NestedString(uncastFeatureGate.Object, "spec", "featureSet")
		if len(currFeatureGateInfo.featureSet) == 0 {
			currFeatureGateInfo.featureSet = "Default"
		}
		allFeatureSets.Insert(currFeatureGateInfo.featureSet)

		uncastFeatureGateSlice, _, err := unstructured.NestedSlice(uncastFeatureGate.Object, "status", "featureGates")
		if err != nil {
			return nil, nil, nil, nil, fmt.Errorf("no slice found %w", err)
		}

		enabledFeatureGates, _, err := unstructured.NestedSlice(uncastFeatureGateSlice[0].(map[string]interface{}), "enabled")
		if err != nil {
			return nil, nil, nil, nil, fmt.Errorf("no enabled found %w", err)
		}
		for _, currGate := range enabledFeatureGates {
			featureGateName, _, err := unstructured.NestedString(currGate.(map[string]interface{}), "name")
			if err != nil {
				return nil, nil, nil, nil, fmt.Errorf("no gate name found %w", err)
			}
			currFeatureGateInfo.enabled.Insert(featureGateName)
			currFeatureGateInfo.allFeatureGates[featureGateName] = true
			allFeatureGates.Insert(featureGateName)
		}

		disabledFeatureGates, _, err := unstructured.NestedSlice(uncastFeatureGateSlice[0].(map[string]interface{}), "disabled")
		if err != nil {
			return nil, nil, nil, nil, fmt.Errorf("no enabled found %w", err)
		}
		for _, currGate := range disabledFeatureGates {
			featureGateName, _, err := unstructured.NestedString(currGate.(map[string]interface{}), "name")
			if err != nil {
				return nil, nil, nil, nil, fmt.Errorf("no gate name found %w", err)
			}
			currFeatureGateInfo.disabled.Insert(featureGateName)
			currFeatureGateInfo.allFeatureGates[featureGateName] = false
			allFeatureGates.Insert(featureGateName)
		}

		existing, ok := clusterProfileToFeatureSetToFeatureGates[currFeatureGateInfo.clusterProfile]
		if !ok {
			existing = map[string]*featureGateInfo{}
			clusterProfileToFeatureSetToFeatureGates[currFeatureGateInfo.clusterProfile] = existing
		}
		existing[currFeatureGateInfo.featureSet] = currFeatureGateInfo
		clusterProfileToFeatureSetToFeatureGates[currFeatureGateInfo.clusterProfile] = existing
	}

	return allClusterProfiles, allFeatureSets, allFeatureGates, clusterProfileToFeatureSetToFeatureGates, nil
}

func clusterOperatorClusterProfilesFrom(annotations map[string]string) sets.String {
	ret := sets.NewString()
	for k, v := range annotations {
		if strings.HasPrefix(k, "include.release.openshift.io/") && v == "false-except-for-the-config-operator" {
			ret.Insert(k)
		}
	}
	return ret
}
