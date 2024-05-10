package build

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/openshift/microshift/test/pkg/testutil"

	"k8s.io/klog/v2"
)

var _ Build = (*BlueprintBuild)(nil)

type BlueprintBuild struct {
	build
	// Commit    bool // TODO: Build ISO without commit

	Installer            bool
	InstallerDestination string

	Contents string
	Parent   string
	Aliases  []string
}

func NewBlueprintBuild(path string, opts *PlannerOpts) (*BlueprintBuild, error) {
	klog.InfoS("Constructing BlueprintBuild", "path", path)

	start := time.Now()

	filename := filepath.Base(path)
	withoutExt := strings.TrimSuffix(filename, filepath.Ext(filename))
	dir := filepath.Dir(path)

	dataBytes, err := fs.ReadFile(opts.BlueprintsFS, path)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", path, err)
	}
	data := string(dataBytes)

	// Get `name` directly from the TOML file to not operate on assumption
	// that filename without extension is name of the blueprint.
	// Also, if we template the whole file and it ends up blank we cannot get the name.
	name, err := testutil.GetTOMLFieldValue(data, "name")
	if err != nil {
		return nil, fmt.Errorf("failed to obtain value of field %q in file %q", "name", path)
	}

	// If name has contains templating syntax, it needs to be templated to get true name
	if strings.Contains(name, "{{") {
		name, err = opts.TplData.Template(fmt.Sprintf("name-of-%s", filename), name)
		if err != nil {
			return nil, err
		}
	}

	templatedData, err := opts.TplData.Template(name, data)
	if err != nil {
		opts.Events.AddEvent(&testutil.FailedEvent{
			Event: testutil.Event{
				Name:      name,
				Suite:     "render",
				ClassName: "blueprint",
				Start:     start,
				End:       time.Now(),
			},
			Message: "Failed to render template",
			Content: err.Error(),
		})
		return nil, err
	}
	opts.Events.AddEvent(&testutil.Event{
		Name:      name,
		Suite:     "render",
		ClassName: "blueprint",
		Start:     start,
		End:       time.Now(),
		SystemOut: templatedData,
	})

	bb := &BlueprintBuild{
		build: build{
			Name: name,
			Path: path,
		},
		Contents: templatedData,
	}

	// blueprint.alias file contains aliases for commit defined in blueprint.toml
	potentialAliasFile := fmt.Sprintf("%s.alias", withoutExt)
	if exists, err := fileExistsInDir(opts.BlueprintsFS, dir, potentialAliasFile); err != nil {
		return nil, err
	} else if exists {
		data, err := fs.ReadFile(opts.BlueprintsFS, filepath.Join(dir, potentialAliasFile))
		if err != nil {
			return nil, err
		}
		bb.Aliases = slices.DeleteFunc(strings.Split(string(data), "\n"), func(line string) bool {
			// Remove empty lines and lines with whitespace only.
			return strings.TrimSpace(line) == ""
		})
	}

	if opts.BuildInstallers {
		// If blueprint.image-installer exists, then ISO installer can be built.
		if exists, err := fileExistsInDir(opts.BlueprintsFS, dir, fmt.Sprintf("%s.image-installer", withoutExt)); err != nil {
			return nil, err
		} else if exists {
			bb.Installer = true
			bb.InstallerDestination = filepath.Join(opts.Paths.VMStorageDir, name+".iso")
		}
	}

	// looking for parent commit
	if strings.Contains(withoutExt, "-") {
		parts := strings.Split(withoutExt, "-")
		expectedParentFilename := parts[0] + ".toml"

		parentPath := ""
		err = fs.WalkDir(opts.BlueprintsFS, ".", func(p string, d fs.DirEntry, err error) error {
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
			parentData, err := fs.ReadFile(opts.BlueprintsFS, parentPath)
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

func (b *BlueprintBuild) Prepare(opts *Opts) error {
	// TODO: Do we need to remove Blueprint before to adding? It's not required and it only bumps blueprint's version
	start := time.Now()

	err := opts.Composer.AddBlueprint(b.Contents)
	if err != nil {
		opts.Events.AddEvent(&testutil.FailedEvent{
			Event: testutil.Event{
				Name:      b.Name,
				Suite:     "add",
				ClassName: "add",
				Start:     start,
				End:       time.Now(),
			},
			Message: "Adding blueprint failed",
			Content: err.Error(),
		})
		return err
	}
	err = opts.Composer.DepsolveBlueprint(b.Name)
	if err != nil {
		opts.Events.AddEvent(&testutil.FailedEvent{
			Event: testutil.Event{
				Name:      b.Name,
				Suite:     "depsolve",
				ClassName: "depsolve",
				Start:     start,
				End:       time.Now(),
			},
			Message: "Depsolve failed",
			Content: err.Error(),
		})
		return err
	}
	opts.Events.AddEvent(&testutil.Event{
		Name:      b.Name,
		Suite:     "depsolve",
		ClassName: "depsolve",
		Start:     start,
		End:       time.Now(),
	})
	return nil
}

func (b *BlueprintBuild) Execute(ctx context.Context, opts *Opts) error {
	refExists, err := opts.Ostree.DoesRefExists(b.Name)
	if err != nil {
		return err
	}
	start := time.Now()

	aeg := testutil.NewAllErrGroup()

	skipCommit := refExists && !opts.Force

	if skipCommit {
		klog.InfoS("Commit already present in the ostree repository and --force wasn't present - skipping", "blueprint", b.Name)
		opts.Events.AddEvent(&testutil.SkippedEvent{
			Event: testutil.Event{
				Name:      b.Name,
				Suite:     "compose",
				ClassName: "commit",
				Start:     start,
				End:       time.Now(),
			},
			Message: "Commit already present in the ostree repository and --force wasn't present - skipping",
		})

	} else {
		aeg.Go(func() error {
			if err := testutil.Retry(ctx, opts.Retries, opts.RetryInterval, func() error { return b.composeCommit(ctx, opts) }); err != nil {
				klog.ErrorS(err, "Composing commit failed", "blueprint", b.Name)

				opts.Events.AddEvent(&testutil.FailedEvent{
					Event: testutil.Event{
						Name:      b.Name,
						Suite:     "compose",
						ClassName: "commit",
						Start:     start,
						End:       time.Now(),
					},
					Message: "Composing commit failed",
					Content: err.Error(),
				})

				return err
			}

			opts.Events.AddEvent(&testutil.Event{
				Name:      b.Name,
				Suite:     "compose",
				ClassName: "commit",
				Start:     start,
				End:       time.Now(),
			})

			return nil
		})
	}

	if b.Installer {
		if isoExists, err := opts.Utils.PathExistsAndIsNotEmpty(b.InstallerDestination); err != nil {
			return err
		} else if isoExists && !opts.Force {
			klog.InfoS("ISO installer already present in vm-storage and --force wasn't present - skipping", "blueprint", b.Name)

			opts.Events.AddEvent(&testutil.SkippedEvent{
				Event: testutil.Event{
					Name:      b.Name,
					Suite:     "compose",
					ClassName: "installer",
					Start:     start,
					End:       time.Now(),
				},
				Message: "ISO installer already present in vm-storage and --force wasn't present - skipping",
			})

		} else {
			aeg.Go(func() error {
				if err := testutil.Retry(ctx, opts.Retries, opts.RetryInterval, func() error { return b.composeInstaller(ctx, opts) }); err != nil {
					klog.ErrorS(err, "Composing installer failed", "blueprint", b.Name)
					opts.Events.AddEvent(&testutil.FailedEvent{
						Event: testutil.Event{
							Name:      b.Name,
							Suite:     "compose",
							ClassName: "installer",
							Start:     start,
							End:       time.Now(),
						},
						Message: "Composing installer failed",
						Content: err.Error(),
					})
					return err
				}

				opts.Events.AddEvent(&testutil.Event{
					Name:      b.Name,
					Suite:     "compose",
					ClassName: "installer",
					Start:     start,
					End:       time.Now(),
				})
				return nil
			})
		}
	}

	err = aeg.Wait()
	if err != nil {
		klog.ErrorS(err, "Building blueprint failed")
		return err
	}

	if !skipCommit && len(b.Aliases) != 0 {
		klog.InfoS("Adding aliases", "name", b.Name)
		err = opts.Ostree.CreateAlias(b.Name, b.Aliases...)
		if err != nil {
			opts.Events.AddEvent(&testutil.FailedEvent{
				Event: testutil.Event{
					Name:      b.Name,
					Suite:     "compose",
					ClassName: "alias",
					Start:     start,
					End:       time.Now(),
				},
				Message: "Adding aliases failed",
				Content: err.Error(),
			})
			return err
		}

		opts.Events.AddEvent(&testutil.Event{
			Name:      b.Name,
			Suite:     "compose",
			ClassName: "alias",
			Start:     start,
			End:       time.Now(),
			SystemOut: strings.Join(b.Aliases, " "),
		})
	}

	return nil
}

func (b *BlueprintBuild) composeCommit(ctx context.Context, opts *Opts) error {
	var commitID string
	var err error

	start := time.Now()
	klog.InfoS("Starting commit compose procedure", "blueprint", b.Name)

	commitID, err = opts.Composer.StartOSTreeCompose(b.Name, "edge-commit", b.Name, b.Parent)
	if err != nil {
		return err
	}

	friendlyName := fmt.Sprintf("%s_edge-commit_%s", b.Name, commitID)

	waitErr := opts.Composer.WaitForCompose(ctx, commitID, friendlyName, 15*time.Minute)

	// Get metadata and logs even if composing failed, unless context was cancelled
	if errors.Is(waitErr, context.Canceled) {
		return waitErr
	}

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

	klog.InfoS("Commit compose procedure done", "blueprint", b.Name, "elapsed", time.Since(start))

	return nil
}

func (b *BlueprintBuild) composeInstaller(ctx context.Context, opts *Opts) error {
	start := time.Now()
	klog.InfoS("Starting installer compose procedure", "blueprint", b.Name)

	installerID, err := opts.Composer.StartCompose(b.Name, "image-installer")
	if err != nil {
		return err
	}

	friendlyName := fmt.Sprintf("%s_image-installer_%s", b.Name, installerID)

	waitErr := opts.Composer.WaitForCompose(ctx, installerID, friendlyName, 25*time.Minute)

	// Get metadata and logs even if composing failed, unless context was cancelled
	if errors.Is(waitErr, context.Canceled) {
		return waitErr
	}

	metadataErr := opts.Composer.SaveComposeMetadata(installerID, friendlyName)
	logsErr := opts.Composer.SaveComposeLogs(installerID, friendlyName)

	if err := errors.Join(waitErr, metadataErr, logsErr); err != nil {
		return err
	}

	installerPath, err := opts.Composer.SaveComposeImage(installerID, friendlyName, ".iso")
	if err != nil {
		return err
	}

	dest := filepath.Join(opts.Paths.VMStorageDir, b.Name+".iso")
	err = opts.Utils.Rename(installerPath, dest)
	if err != nil {
		return fmt.Errorf("failed to move installer from %q to %q: %w", installerPath, dest, err)
	}

	klog.InfoS("Moved installer file", "destination", dest)

	klog.InfoS("Installer procedure done", "blueprint", b.Name, "elapsed", time.Since(start))

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
