package build

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"text/template"
	"time"

	"github.com/openshift/microshift/test/pkg/testutil"

	"golang.org/x/sync/errgroup"
	"k8s.io/klog/v2"
)

var _ Build = (*BlueprintBuild)(nil)

type BlueprintBuild struct {
	build
	// Commit    bool // TODO: Build ISO without commit
	Installer bool
	Contents  string
	Parent    string
	Aliases   []string
}

func NewBlueprintBuild(path string, opts *PlannerOpts) (*BlueprintBuild, error) {
	filename := filepath.Base(path)
	withoutExt := strings.TrimSuffix(filename, filepath.Ext(filename))
	dir := filepath.Dir(path)

	dataBytes, err := fs.ReadFile(opts.Filesys, path)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", path, err)
	}
	data := string(dataBytes)
	name, err := testutil.GetTOMLFieldValue(data, "name")
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
			Name: name,
			Path: path,
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
			parentName, err := testutil.GetTOMLFieldValue(string(parentData), "name")
			if err != nil {
				return nil, fmt.Errorf("failed to read name from %q: %w", parentPath, err)
			}
			bb.Parent = parentName
		}
	}

	return bb, nil
}

func (b *BlueprintBuild) Execute(opts *Opts) error {
	// TODO: Do we need to remove Blueprint before to adding? It's not required and it only bumps blueprint's version
	err := opts.Composer.AddBlueprint(b.Contents)
	if err != nil {
		return err
	}
	err = opts.Composer.DepsolveBlueprint(b.Name)
	if err != nil {
		return err
	}

	refExists, err := opts.Ostree.DoesRefExists(b.Name)
	if err != nil {
		return err
	}

	eg, _ := errgroup.WithContext(context.TODO())

	if refExists && !opts.Force {
		klog.InfoS("Commit already present in the ostree repository and --force wasn't present - skipping", "blueprint", b.Name)
	}

	if !refExists || opts.Force {
		eg.Go(func() error {
			if err := b.composeCommit(opts); err != nil {
				klog.ErrorS(err, "Composing commit failed", "blueprint", b.Name)
				return err
			}
			return nil
		})
	}

	// TODO: Check if ISO exists

	if b.Installer {
		eg.Go(func() error {
			if err := b.composeInstaller(opts); err != nil {
				klog.ErrorS(err, "Composing installer failed", "blueprint", b.Name)
				return err
			}
			return nil
		})
	}

	err = eg.Wait()
	if err != nil {
		return err
	}

	if len(b.Aliases) != 0 {
		klog.InfoS("Adding aliases", "name", b.Name)
		err = opts.Ostree.CreateAlias(b.Name, b.Aliases...)
		if err != nil {
			return err
		}
	}

	return nil
}

func (b *BlueprintBuild) composeCommit(opts *Opts) error {
	var commitID string
	var err error

	if b.Parent == "" {
		commitID, err = opts.Composer.StartOSTreeCompose(b.Name, "edge-commit", b.Name, "", "", 0)
	} else {
		commitID, err = opts.Composer.StartOSTreeCompose(b.Name, "edge-commit", b.Name, b.Parent, "http://127.0.0.1:8080/repo", 0)
	}
	if err != nil {
		return err
	}

	friendlyName := fmt.Sprintf("%s_edge-commit_%s", b.Name, commitID)

	waitErr := opts.Composer.WaitForCompose(commitID, friendlyName, 15*time.Minute)

	// Get metadata and logs even if composing failed
	metadataErr := opts.Composer.SaveComposeMetadata(commitID, friendlyName)
	logsErr := opts.Composer.SaveComposeLogs(commitID, friendlyName)

	if err := errors.Join(waitErr, metadataErr, logsErr); err != nil {
		return err
	}

	commitArchivePath, err := opts.Composer.SaveComposeImage(commitID, friendlyName, ".tar")
	if err != nil {
		return err
	}
	err = opts.Ostree.ExtractCommit(commitArchivePath)
	if err != nil {
		return err
	}

	return nil
}

func (b *BlueprintBuild) composeInstaller(opts *Opts) error {
	installerID, err := opts.Composer.StartCompose(b.Name, "image-installer", 0)
	if err != nil {
		return err
	}

	friendlyName := fmt.Sprintf("%s_image-installer_%s", b.Name, installerID)

	waitErr := opts.Composer.WaitForCompose(installerID, friendlyName, 25*time.Minute)

	// Get metadata and logs even if composing failed
	metadataErr := opts.Composer.SaveComposeMetadata(installerID, friendlyName)
	logsErr := opts.Composer.SaveComposeLogs(installerID, friendlyName)

	if err := errors.Join(waitErr, metadataErr, logsErr); err != nil {
		return err
	}

	installerPath, err := opts.Composer.SaveComposeImage(installerID, friendlyName, ".iso")
	if err != nil {
		return err
	}

	dest := filepath.Join(opts.ArtifactsMainDir, "vm-storage", b.Name+".iso")
	err = os.Rename(installerPath, dest)
	if err != nil {
		return fmt.Errorf("failed to move installer from %q to %q: %w", installerPath, dest, err)
	}

	klog.InfoS("Moved installer file", "destination", dest)
	return nil
}
