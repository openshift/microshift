package schemapatch

import (
	"bytes"
	"fmt"
	"golang.org/x/tools/go/packages"
	"io"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/sets"
	"os"
	"os/exec"
	"path/filepath"
	"sigs.k8s.io/controller-tools/pkg/crd/markers"
	"sigs.k8s.io/controller-tools/pkg/genall"
	"sigs.k8s.io/controller-tools/pkg/loader"
	"sigs.k8s.io/controller-tools/pkg/schemapatcher"
	"strings"
)

const openshiftFeatureSetEnv = "OPENSHIFT_REQUIRED_FEATURESET"

// executeSchemaPatchForManifestWithBinary executes the controller-gen binary with the schemapatch:manifests arg.
func executeSchemaPatchForManifestWithBinary(controllerGen string, dir string, versionPaths []string, buf *bytes.Buffer, requiredFeatureSets sets.String) error {
	if requiredFeatureSets.Len() > 0 {
		// The controller generator picks up feature sets from an env var.
		if err := os.Setenv(openshiftFeatureSetEnv, strings.Join(requiredFeatureSets.List(), ",")); err != nil {
			return fmt.Errorf("could not set env %s: %w", openshiftFeatureSetEnv, err)
		}

		defer os.Unsetenv(openshiftFeatureSetEnv)
	}

	args := []string{}

	args = append(args, manifestsArg(dir))
	args = append(args, pathsArgs(versionPaths)...)
	args = append(args, "output:stdout")

	cmd := exec.Command(controllerGen, args...)

	// Ensure we get the output from the command.
	cmd.Stdout = buf
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error running controller-gen: %w", err)
	}

	return nil
}

// manifestArg generates the schemapatch:manifests arg for the controller-gen binary.
func manifestsArg(versionPath string) string {
	return fmt.Sprintf("schemapatch:manifests=\"%s\"", versionPath)
}

// pathsArg generates the paths arg for the controller-gen binary.
func pathsArgs(versionPaths []string) []string {
	paths := []string{}

	for _, path := range versionPaths {
		paths = append(paths, fmt.Sprintf("paths=%s", path))
	}

	return paths
}

// executeSchemaPatchForManifest executes the schemapatch generator against the generation context passed.
// When the controller-gen binary is available, it will be used to generate the patch, else the code integration
// will be used.
// The output of the patch is written to the buffer passed.
func executeSchemaPatchForManifest(gc schemaPatchGenerationContext, buf *bytes.Buffer, versionPaths []string, rt *genall.Runtime, controllerGen string) error {
	// To generate a single schema we must put the manifest in a directory of its own.
	// Use a temp directory and remove it once the function exits.
	dir, err := os.MkdirTemp("", "schemapatch")
	if err != nil {
		return fmt.Errorf("could not create temp dir: %w", err)
	}
	defer os.RemoveAll(dir)

	if err := os.WriteFile(filepath.Join(dir, "manifest.yaml"), gc.manifestData, 0644); err != nil {
		return fmt.Errorf("could not write manifest to temp dir: %w", err)
	}

	// If controllerGen is not empty, use the binary instead of the code integration.
	if controllerGen != "" {
		return executeSchemaPatchForManifestWithBinary(controllerGen, dir, versionPaths, buf, gc.requiredFeatureSets)
	}

	markers.RequiredFeatureSets.Insert(gc.requiredFeatureSets.List()...)
	defer func() {
		markers.RequiredFeatureSets = sets.NewString()
	}()

	gen := schemapatcher.Generator{
		ManifestsPath: dir,
	}

	ctx := rt.GenerationContext
	ctx.OutputRule = &outputToBuffer{buf}

	if err := gen.RegisterMarkers(ctx.Collector.Registry); err != nil {
		return fmt.Errorf("could not register markers: %w", err)
	}

	if err := gen.Generate(&ctx); err != nil {
		return fmt.Errorf("could not run schemapatch generator: %w", err)
	}

	if errs := collectErrors(&ctx); len(errs) > 0 {
		return kerrors.NewAggregate(errs)
	}

	return nil
}

// loadGroupRuntime builds a genall.Runtime based on the package paths for all version that are passed.
// This allows the runtime to be shared between each version when it's generated.
func loadGroupRuntime(paths []string) (*genall.Runtime, error) {
	generators := &genall.Generators{}
	return generators.ForRoots(paths...)
}

// outputToBuffer is a WriteCloser that writes to a buffer.
// This is used as the output for the controller-gen generator integration.
type outputToBuffer struct {
	*bytes.Buffer
}

// Open implements the Open method of the io.WriteCloser interface.
func (o *outputToBuffer) Open(_ *loader.Package, _ string) (io.WriteCloser, error) {
	return o, nil
}

// Close implements the Close method of the io.WriteCloser interface.
func (o *outputToBuffer) Close() error {
	return nil
}

// collectErrors iterates through the packages generated and their imports
// looking for errors to collate.
// It ignores type errors as we only provide partial types through to the
// generator for transitively imported packages.
// This should help identify issues with dependent packages.
func collectErrors(ctx *genall.GenerationContext) []error {
	var errs []error
	uniqueErrors := sets.NewString()

	for _, root := range ctx.Roots {
		if len(root.Errors) == 0 {
			// This package generated correctly.
			// Don't worry about inspecting the imports for errors.
			continue
		}

		for _, err := range root.Errors {
			if err.Kind == packages.TypeError {
				// Ignore type errors as these are common
				// and don't typically cause bad generation output.
				continue
			}

			if !uniqueErrors.Has(err.Error()) {
				errs = append(errs, err)
				uniqueErrors.Insert(err.Error())
			}

		}

		for _, imp := range root.Imports() {
			for _, err := range imp.Errors {
				if err.Kind == packages.TypeError {
					// Ignore type errors as these are common
					// and don't typically cause bad generation output.
					continue
				}

				if !uniqueErrors.Has(err.Error()) {
					errs = append(errs, err)
					uniqueErrors.Insert(err.Error())
				}
			}
		}
	}

	return errs
}
