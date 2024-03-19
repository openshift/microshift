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

type BlueprintBuild struct {
	build
	// Commit    bool // TODO: Build ISO without commit
	Installer bool
	Contents  string
	Parent    string
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

	bb := &BlueprintBuild{
		build: build{
			Name:     name,
			Path:     path,
			Composer: opts.Composer,
			Ostree:   opts.Ostree,
		},
		Contents: templatedData.String()}

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

	// looking for parent commit
	if strings.Contains(withoutExt, "-") {
		parts := strings.Split(withoutExt, "-")
		expectedParentFilename := parts[0] + ".toml"

		parentPath := ""
		err = fs.WalkDir(opts.Filesys, ".", func(p string, d fs.DirEntry, err error) error {
			if parentPath != "" {
				return nil
			}
			if d.Name() == expectedParentFilename {
				parentPath = p
			}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("walking through dirs to find parent of %q failed: %w", path, err)
		}

		if parentPath != "" {
			parentData, err := fs.ReadFile(opts.Filesys, parentPath)
			if err != nil {
				return nil, fmt.Errorf("failed to read parent of %q which is %q: %w", path, parentPath, err)
			}
			parentName, err := getTOMLFieldValue(string(parentData), "name")
			if err != nil {
				return nil, fmt.Errorf("failed to read name from %q: %w", parentPath, err)
			}
			bb.Parent = parentName
		}
	}

	return bb, nil
}

func (b *BlueprintBuild) Execute() error {
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
