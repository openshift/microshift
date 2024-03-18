package compose

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"slices"
	"strings"
	"text/template"
	"time"

	"k8s.io/klog/v2"
)

type Build interface {
	Execute(Composer) error
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

	// Filesys is filesystem used to obtain files by given path.
	Filesys fs.FS

	// TplData is a struct used as templating input.
	TplData *TemplatingData
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

type build struct {
	Name string
	Path string
}

type BlueprintBuild struct {
	build
	// Commit    bool // TODO: Build ISO without commit
	Installer bool
	Contents  string
	Aliases   []string
}

func NewBlueprintBuild(path string, opts *BuildOpts) (*BlueprintBuild, error) {
	filename := filepath.Base(path)
	withoutExt := strings.TrimSuffix(filename, filepath.Ext(filename))
	dir := filepath.Dir(path)

	dataBytes, err := fs.ReadFile(opts.Filesys, path)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", path, err)
	}
	data := string(dataBytes)
	name, err := getTOMLFieldValue(data, "name")
	if err != nil {
		return nil, fmt.Errorf("failed to obtain value of field %q in file %q", "name", path)
	}

	if strings.Contains(name, "{{") {
		nameTpl, err := template.New(fmt.Sprintf("name-of-%s", filename)).Parse(name)
		if err != nil {
			return nil, fmt.Errorf("failed to template name of %q: %q", filename, name)
		}
		templatedName := strings.Builder{}
		if err := nameTpl.Execute(&templatedName, opts.TplData); err != nil {
			return nil, fmt.Errorf("failed to execute template %q: %w", name, err)
		}
		name = templatedName.String()
	}

	tpl, err := template.New(name).Parse(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template %s: %w", "", err)
	}
	templatedData := strings.Builder{}
	if err := tpl.Execute(&templatedData, opts.TplData); err != nil {
		return nil, fmt.Errorf("failed to execute template %q: %w", path, err)
	}

	bb := &BlueprintBuild{build: build{Name: name, Path: path}, Contents: templatedData.String()}

	// blueprint.alias file contains aliases for commit defined in blueprint.toml
	potentialAliasFile := fmt.Sprintf("%s.alias", withoutExt)
	if exists, err := fileExistsInDir(opts.Filesys, dir, potentialAliasFile); err != nil {
		return nil, err
	} else if exists {
		data, err := fs.ReadFile(opts.Filesys, filepath.Join(dir, potentialAliasFile))
		if err != nil {
			return nil, err
		}
		bb.Aliases = slices.DeleteFunc(strings.Split(string(data), "\n"), func(line string) bool { return line == "" })
	}

	if opts.BuildInstallers {
		// If blueprint.image-installer exists, then ISO installer can be built.
		if exists, err := fileExistsInDir(opts.Filesys, dir, fmt.Sprintf("%s.image-installer", withoutExt)); err != nil {
			return nil, err
		} else if exists {
			bb.Installer = true
		}
	}
	return bb, nil
}

func (b *BlueprintBuild) Execute(Composer) error {
	klog.InfoS("Building blueprint", "name", b.Name)
	time.Sleep(1 * time.Second)
	klog.InfoS("Blueprint done", "name", b.Name)

	if b.Installer {
		klog.InfoS("Building installer", "name", b.Name)
		time.Sleep(1 * time.Second)
		klog.InfoS("Installer built", "name", b.Name)
	}

	if len(b.Aliases) != 0 {
		klog.InfoS("Adding aliases", "name", b.Name)
	}

	return nil
}

type ImageFetcher struct {
	build
	Url string
}

func NewImageFetcher(path string, opts *BuildOpts) (*ImageFetcher, error) {
	filename := filepath.Base(path)
	withoutExt := strings.TrimSuffix(filename, filepath.Ext(filename))

	dataBytes, err := fs.ReadFile(opts.Filesys, path)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", path, err)
	}

	tpl, err := template.New(withoutExt).Parse(string(dataBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to parse template %s: %w", "", err)
	}
	templatedData := strings.Builder{}
	if err := tpl.Execute(&templatedData, opts.TplData); err != nil {
		return nil, fmt.Errorf("failed to execute template %q: %w", path, err)
	}

	return &ImageFetcher{
		build: build{
			Name: withoutExt,
			Path: path,
		},
		Url: templatedData.String(),
	}, nil
}

func (i *ImageFetcher) Execute(Composer) error {
	klog.InfoS("Downloading image", "name", i.Name)
	time.Sleep(1 * time.Second)
	klog.InfoS("Image downloaded", "name", i.Name)
	return nil
}

func fileExistsInDir(filesys fs.FS, dir, filename string) (bool, error) {
	entries, err := fs.ReadDir(filesys, dir)
	if err != nil {
		return false, err
	}
	for _, entry := range entries {
		if entry.Name() == filename {
			return true, nil
		}
	}
	return false, nil
}
