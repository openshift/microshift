package util

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/util/sets"
)

func Must(err error) {
	if err != nil {
		panic(fmt.Errorf("internal error: %v", err))
	}
}

func Default(s string, defaultS string) string {
	if s == "" {
		return defaultS
	}
	return s
}

func PathExistsFS(fsys fs.StatFS, path string) (bool, error) {
	if _, err := fsys.Stat(path); err == nil {
		return true, nil
	} else if errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else {
		return false, fmt.Errorf("checking if path %q exists in the fsys %q failed: %w", path, fsys, err)
	}
}

func PathExists(path string) (bool, error) {
	if _, err := os.Stat(path); err == nil {
		return true, nil
	} else if errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else {
		return false, fmt.Errorf("checking if path %q exists failed: %w", path, err)
	}
}

func MakeDir(path string) error {
	return os.MkdirAll(path, 0700)
}

func PathExistsAndIsNotEmpty(path string, ignores ...string) (bool, error) {
	fi, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		} else {
			return false, fmt.Errorf("checking if path (%s) exists failed: %w", path, err)
		}
	}

	if !fi.IsDir() {
		return fi.Size() != 0, nil
	}

	files, err := os.ReadDir(path)
	if err != nil {
		return false, fmt.Errorf("failed to ReadDir %q: %w", path, err)
	}

	iset := sets.New[string](ignores...)
	for _, f := range files {
		if iset.Has(f.Name()) {
			continue
		}
		return true, nil
	}

	return false, nil
}

// StartHealthCheck starts a server for a simple health check endpoint
// Returns a start and shutdown handler.
//
// Note: typically servers return a non-nil error, here we return nil
// if the server was naturally shutdown.
func HealthCheckServer(ctx context.Context, path, port string) (start func() error, shutdown func() error) {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	livenessMux := http.NewServeMux()
	livenessMux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintf(w, "ok")
	})

	server := http.Server{
		ReadTimeout: time.Second * 10,
		Addr:        ":" + port,
		Handler:     livenessMux,
	}

	start = func() error {
		err := server.ListenAndServe()
		if err != http.ErrServerClosed {
			return err
		}
		return nil
	}

	shutdown = func() error {
		ctx, cancel := context.WithTimeout(ctx, time.Millisecond*10)
		defer cancel()
		err := server.Shutdown(ctx)
		if err != http.ErrServerClosed {
			return err
		}
		return nil
	}

	return start, shutdown
}

type LogFilePath string

// Write writes data to the desired filename, calling multiple times will overwrite
// the file with the most recent data.
func (l LogFilePath) Write(data []byte) error {
	err := MakeDir(filepath.Dir(string(l)))
	if err != nil {
		return err
	}
	return os.WriteFile(string(l), data, 0600)
}

// Remove Will remove the existing file if it exists.
func (l LogFilePath) Remove() error {
	exists, err := PathExists(string(l))
	if exists {
		return os.Remove(string(l))
	}
	return err
}

// GetTempPathArgs returns the directory in which to create a temp file
// along with the pattern for the temp file name.
func getTempPathArgs(path string) (string, string) {
	dir, file := filepath.Split(path)
	pattern := file + ".tmp."
	return dir, pattern
}

// CreateTempFile creates a temporary file from given path and returns
// resulting file
func CreateTempFile(path string) (*os.File, error) {
	dir, pattern := getTempPathArgs(path)
	return os.CreateTemp(dir, pattern)
}

// CreateTempDir creates a temporary directory from given path and returns
// the pathname of the new directory.
func CreateTempDir(path string) (string, error) {
	dir, pattern := getTempPathArgs(path)
	return os.MkdirTemp(dir, pattern)
}

func IsOSTree() (bool, error) {
	return PathExists("/run/ostree-booted")
}

func GetOSVersion() (string, error) {
	data, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return "", fmt.Errorf("error reading /etc/os-release: %v", err)
	}
	content := string(data)

	var version string
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "VERSION_ID=") {
			version = strings.TrimPrefix(line, "VERSION_ID=")
			version = strings.Trim(version, `"`)
			break
		}
	}
	if version == "" {
		return "", fmt.Errorf("VERSION_ID not found in /etc/os-release")
	}
	return version, nil
}

func IsBootc() bool {
	cmd := exec.Command("bootc", "status", "--booted", "--json")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	var result struct {
		Status struct {
			Type string `json:"type"`
		} `json:"status"`
	}

	err = json.Unmarshal(output, &result)
	if err != nil {
		return false
	}

	return result.Status.Type == "bootcHost"
}
