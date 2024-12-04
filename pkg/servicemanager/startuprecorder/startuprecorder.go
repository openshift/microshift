package startuprecorder

import (
	"encoding/json"
	"os"
	"sync"
	"time"

	"k8s.io/klog/v2"
)

type ServiceData struct {
	Name         string        `json:"name"`
	Dependencies []string      `json:"dependencies"`
	Start        time.Time     `json:"start"`
	Ready        time.Time     `json:"ready"`
	TimeToReady  time.Duration `json:"timeToReady"`
}

type StartupRecorder struct {
	Services              []ServiceData `json:"services"`
	MicroshiftStart       time.Time     `json:"microshiftStart"`
	MicroshiftReady       time.Time     `json:"microshiftReady"`
	MicroshiftTimeToReady time.Duration `json:"microshiftTimeToReady"`

	m sync.Mutex
}

func New() *StartupRecorder {
	return &StartupRecorder{}
}

func (l *StartupRecorder) LogServiceReady(serviceName string, dependencies []string, start time.Time, ready time.Time) {
	serviceData := ServiceData{
		Name:         serviceName,
		Dependencies: dependencies,
		Start:        start,
		Ready:        ready,
		TimeToReady:  ready.Sub(start),
	}

	l.m.Lock()
	l.Services = append(l.Services, serviceData)
	l.m.Unlock()

	klog.InfoS("SERVICE READY", "service", serviceName, "since-start", ready.Sub(start))
}

func (l *StartupRecorder) LogMicroshiftStart(start time.Time) {
	klog.InfoS("MICROSHIFT STARTING")
	l.MicroshiftStart = start
}

func (l *StartupRecorder) LogMicroshiftReady(ready time.Time) {
	klog.InfoS("MICROSHIFT READY", "since-start", time.Since(l.MicroshiftStart))
	l.MicroshiftReady = ready
	l.MicroshiftTimeToReady = ready.Sub(l.MicroshiftStart)
}

func (l *StartupRecorder) OutputData() error {
	jsonOutput, _ := json.Marshal(l)

	klog.Info(string(jsonOutput))

	path, ok := os.LookupEnv("STARTUP_LOGS_PATH")
	if ok {
		return os.WriteFile(path, jsonOutput, 0644)
	}

	return nil
}
