package multilogger

import (
	"log"
	"os"
	"path/filepath"

	"github.com/go-logr/stdr"
	"k8s.io/klog/v2"
)

type fileLogger struct {
	logger *klog.Logger
}

func (f *fileLogger) InfoS(msg string, kv ...interface{}) {
	klog.InfoS(msg, kv...)
	f.logger.Info(msg, kv...)
}

func (f *fileLogger) ErrorS(err error, msg string, kv ...interface{}) {
	klog.ErrorS(err, msg, kv...)
	f.logger.Error(err, msg, kv...)
}

// NewFileLogger creates a logger that writes to both klog and the specified file path.
// The intention is to be best effort logging, meaning that it will try to create
// the full file path if it does not exist.
func NewFileLogger(filePath string) (*fileLogger, error) {
	err := os.MkdirAll(filepath.Dir(filePath), 0755)
	if err != nil {
		return nil, err
	}

	f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	slog := stdr.NewWithOptions(log.New(f, "", log.LstdFlags), stdr.Options{LogCaller: stdr.All})
	logger := klog.New(slog.GetSink())

	return &fileLogger{
		logger: &logger,
	}, nil
}

// MustNewFileLogger Will panic if a filepath cannot be created
func MustNewFileLogger(filePath string) *fileLogger {
	fl, err := NewFileLogger(filePath)
	if err != nil {
		klog.Fatalf("failed to create file multi logger for filepath (%s): %w", filePath, err)
	}
	return fl
}
