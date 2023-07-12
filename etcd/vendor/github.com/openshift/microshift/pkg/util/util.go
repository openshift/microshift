package util

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
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

func PathExists(path string) (bool, error) {
	if _, err := os.Stat(path); err == nil {
		return true, nil
	} else if errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else {
		return false, fmt.Errorf("checking if path (%s) exists failed: %w", path, err)
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
		fmt.Fprintf(w, "ok")
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
