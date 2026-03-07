package deepcopy

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/openshift/api/tools/codegen/pkg/generation"
	"github.com/openshift/api/tools/codegen/pkg/utils"
	"k8s.io/gengo/v2/parser"
	"k8s.io/gengo/v2/types"
	"k8s.io/klog/v2"

	"k8s.io/code-generator/cmd/deepcopy-gen/args"
	"k8s.io/code-generator/cmd/deepcopy-gen/generators"

	gengogenerator "k8s.io/gengo/v2/generator"
)

// generateDeepcopyFunctions generates the DeepCopy functions for the given API package paths.
func generateDeepcopyFunctions(p *parser.Parser, universe types.Universe, path, packagePath, outputBaseFileName, headerFilePath string, verify bool) error {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// The deepcopy generator cannot import from an absolute path.
	inputPath, err := filepath.Rel(wd, path)
	if err != nil {
		return fmt.Errorf("failed to get relative path for %s: %w", path, err)
	}
	// The path must start with `./` to be considered a relative path
	// by the generator.
	inputPath = fmt.Sprintf(".%s%s", string(os.PathSeparator), inputPath)

	pathPrefix, err := utils.GetPathPrefix(wd, inputPath, packagePath)
	if err != nil {
		return fmt.Errorf("failed to get path prefix: %w", err)
	}

	args := &args.Args{
		BoundingDirs: []string{packagePath},
		GoHeaderFile: headerFilePath,
		OutputFile:   outputBaseFileName,
	}

	klog.V(2).Infof("Generating deepcopy into %s", filepath.Join(wd, strings.TrimPrefix(packagePath, pathPrefix)))

	myTargets := func(context *gengogenerator.Context) []gengogenerator.Target {
		return generators.GetTargets(context, args)
	}

	if err := generation.Execute(p, universe,
		generators.NameSystems(),
		generators.DefaultNameSystem(),
		myTargets,
		[]string{packagePath},
	); err != nil {
		return fmt.Errorf("error executing deepcopy generator: %w", err)
	}

	if verify {
		diff, err := utils.GitDiff(inputPath, outputBaseFileName)
		if err != nil {
			return fmt.Errorf("could not calculate git diff: %w", err)
		}
		if len(diff) > 0 {
			return fmt.Errorf("deepcopy for %s is out of date, please regenerate the deepcopy code:\n%s", packagePath, diff)
		}
	}

	return nil
}
