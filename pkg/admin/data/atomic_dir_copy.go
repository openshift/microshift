package data

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/openshift/microshift/pkg/util"
	"k8s.io/klog/v2"
)

var (
	// cpArgs contains arguments passed to 'cp' command required to copy MicroShift
	// data dir as a backup with Copy-On-Write feature.
	cpArgs = []string{
		"--verbose",
		"--recursive",    // copy directories recursively
		"--preserve",     // preserve mode, ownership, timestamps
		"--reflink=auto", // enable Copy-on-Write copy
	}
)

// AtomicDirCopy atomically copies directory.
// It performs a two operation: copies source path to a destination
// location with temporary name, then renames it to final name.
// On Unix systems, the rename operation is atomic. This ensures that the file
// is either fully renamed or not at all, which helps prevent issues like partial
// file copies in cases of power failure or unexpected program termination.
type AtomicDirCopy struct {
	Source      string
	Destination string

	intermediatePath string
}

func (c *AtomicDirCopy) CopyToIntermediate() error {
	if c == nil {
		return nil
	}
	var err error

	c.intermediatePath, err = util.CreateTempDir(c.Destination)
	if err != nil {
		return err
	}
	cmd := exec.Command("cp", append(cpArgs, c.Source, c.intermediatePath)...) //nolint:gosec

	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	err = cmd.Run()

	if err != nil {
		klog.ErrorS(nil, "Failed to make an intermediate copy", "cmd", cmd,
			"stdout", strings.ReplaceAll(outb.String(), "\n", `, `),
			"stderr", errb.String())
		copyErr := fmt.Errorf("failed to copy %q to %q: %w", c.Source, c.intermediatePath, err)
		if err := c.RollbackIntermediate(); err != nil {
			return errors.Join(copyErr, err)
		}
		return copyErr
	}
	klog.InfoS("Made an intermediate copy", "cmd", cmd)

	return nil
}

func (c *AtomicDirCopy) RollbackIntermediate() error {
	if c == nil {
		return nil
	}
	if c.intermediatePath == "" {
		return nil
	}
	if err := os.RemoveAll(c.intermediatePath); err != nil {
		return fmt.Errorf("failed to remove intermediate path %q: %w", c.intermediatePath, err)
	}
	return nil
}

func (c *AtomicDirCopy) RenameToFinal() error {
	if c == nil {
		return nil
	}
	var src, dest string
	if c.intermediatePath == "" {
		// Empty value means there was no intermediate copy.
		src = c.Source
		dest = c.Destination
	} else {
		// Path was copied to a temporary location with .tmp suffix.
		// Now it needs to be renamed into final destination.
		// This two-step operation should provide a high guarantee that
		// copying is complete and not partial thanks to rename being OS/filesystem atomic.
		src = c.intermediatePath
		dest = c.Destination
	}

	// Delete the destination if it's a non-empty directory.
	// This is a limitation of the POSIX's rename.
	if err := removeDirIfExists(dest); err != nil {
		return err
	}

	if err := os.Rename(src, dest); err != nil {
		klog.ErrorS(err, "Failed to rename the path - removing", "path", src)
		if removeErr := os.RemoveAll(src); removeErr != nil {
			return errors.Join(err, fmt.Errorf("failed to remove %q: %w", src, removeErr))
		}
		return err
	}
	klog.InfoS("Renamed to final destination", "src", src, "dest", dest)
	return nil
}

func removeDirIfExists(path string) error {
	fileInfo, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("failed to stat %q: %w", path, err)
	}

	if fileInfo.IsDir() {
		return os.RemoveAll(path)
	}

	return nil
}
