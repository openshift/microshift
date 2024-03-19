package compose

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"k8s.io/klog/v2"
)

type Build interface {
	Execute() error
}

type build struct {
	Name  string
	Path  string
	Force bool

	Composer Composer
	Ostree   Ostree
}

// BuildGroup is a collection of Builds than can run in parallel
type BuildGroup []Build

// BuildPlan is collection of BuildGroups that run sequentially
type BuildPlan []BuildGroup

type BuildOpts struct {
	// BuildInstallers decides with ISO images should be built.
	BuildInstallers bool

	// SourceOnly decides if only blueprints with `source` in the filename should be built.
	// Such blueprints are typically built using currently checked out code.
	SourceOnly bool

	// Force causes images to be rebuilt even if they exist already in ostree repo or vm-storage (ISO)
	Force bool

	// Filesys is filesystem used to obtain files by given path.
	Filesys fs.FS

	// TplData is a struct used as templating input.
	TplData *TemplatingData

	// Composer is an interface to the composer remote API
	Composer Composer

	// Composer is an interface to the ostree repository
	Ostree Ostree
}

type BuildPlanner struct {
	Opts *BuildOpts
}

func (b *BuildPlanner) ConstructBuildTree(path string) (BuildPlan, error) {
	var toBuild BuildPlan

	base := filepath.Base(path)
	if strings.Contains(base, "layer") {
		if layer, err := b.layer(path); err != nil {
			return nil, err
		} else {
			toBuild = layer
		}
	} else if strings.Contains(base, "group") {
		if grp, err := b.group(path); err != nil {
			return nil, err
		} else {
			toBuild = BuildPlan{grp}
		}
	} else if strings.Contains(base, ".toml") || strings.Contains(base, ".image-fetcher") {
		if build, err := b.file(path); err != nil {
			return nil, err
		} else if build != nil {
			toBuild = BuildPlan{BuildGroup{build}}
		}
	} else if strings.Contains(base, ".image-installer") || strings.Contains(base, ".alias") {
		return nil, fmt.Errorf("passing .image-installer or .alias files directly is not supported - only .toml and .image-fetcher file are supported")
	} else {
		return nil, fmt.Errorf("unknown artifact to build")
	}

	klog.InfoS("Constructed BuildRequest", "groups", len(toBuild), "build-request", toBuild)

	return toBuild, nil
}

func (b *BuildPlanner) layer(path string) (BuildPlan, error) {
	klog.InfoS("Constructing BuildTree of the layer", "path", path)

	entries, err := fs.ReadDir(b.Opts.Filesys, path)
	if err != nil {
		return nil, err
	}

	toBuild := make(BuildPlan, 0, len(entries))

	for _, e := range entries {
		if e.IsDir() && strings.HasPrefix(e.Name(), "group") {
			if files, err := b.group(filepath.Join(path, e.Name())); err != nil {
				return nil, err
			} else if len(files) != 0 {
				toBuild = append(toBuild, files)
			}
		}
	}

	return toBuild, nil
}

func (b *BuildPlanner) group(path string) (BuildGroup, error) {
	klog.InfoS("Constructing BuiltTree of the group", "path", path)

	entries, err := fs.ReadDir(b.Opts.Filesys, path)
	if err != nil {
		return nil, err
	}

	toBuild := make(BuildGroup, 0, len(entries))
	for _, e := range entries {
		entryPath := filepath.Join(path, e.Name())
		if e.IsDir() {
			return nil, fmt.Errorf("unexpected directory inside group: %s", entryPath)
		}

		if build, err := b.file(entryPath); err != nil {
			return nil, err
		} else if build != nil {
			toBuild = append(toBuild, build)
		}
	}

	return toBuild, nil
}

func (b *BuildPlanner) file(path string) (Build, error) {
	filename := filepath.Base(path)

	if b.Opts.SourceOnly && !strings.Contains(filename, "source") {
		klog.InfoS("SourceOnly mode - skipping image", "path", path)
		return nil, nil
	}

	switch filepath.Ext(filename) {
	case ".image-installer", ".alias":
		return nil, nil
	case ".image-fetcher":
		return NewImageFetcher(path, b.Opts)
	}

	return NewBlueprintBuild(path, b.Opts)
}
