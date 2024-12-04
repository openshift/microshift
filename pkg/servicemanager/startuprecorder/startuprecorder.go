package startuplogger

import (
	"encoding/json"
	"os"
	"time"

	"k8s.io/klog/v2"
)

type ServiceData struct {
	Name         string    `json:"name"`
	Dependencies []string  `json:"dependencies"`
	Start        time.Time `json:"start"`
	Ready        time.Time `json:"ready"`
}

func (s ServiceData) MarshalJSON() ([]byte, error) {
	type Alias ServiceData
	return json.Marshal(struct {
		Alias
		StartupTime string `json:"duration"`
	}{
		Alias:       (Alias)(s),
		StartupTime: s.Ready.Sub(s.Start).String(),
	})
}

type StartupLogger struct {
	services        []ServiceData
	microshiftStart time.Time
	microshiftReady time.Time
}

func (l StartupLogger) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		MicroshiftStart string        `json:"microshiftStart"`
		MicroshiftReady string        `json:"microshiftReady"`
		Services        []ServiceData `json:"services"`
	}{
		MicroshiftStart: l.microshiftStart.Format(time.RFC3339),
		MicroshiftReady: l.microshiftReady.Format(time.RFC3339),
		Services:        l.services,
	})
}

func NewStartupLogger() *StartupLogger {
	return &StartupLogger{}
}

func (l *StartupLogger) LogService(serviceName string, dependencies []string, start time.Time, ready time.Time) {
	serviceData := ServiceData{
		Name:         serviceName,
		Dependencies: dependencies,
		Start:        start,
		Ready:        ready,
	}

	l.services = append(l.services, serviceData)
}

func (l *StartupLogger) LogMicroshiftStart(start time.Time) {
	l.microshiftStart = start
}

func (l *StartupLogger) LogMicroshiftReady(ready time.Time) {
	l.microshiftReady = ready
}

func (l *StartupLogger) OutputData() error {
	jsonOutput, _ := json.MarshalIndent(l, "", "  ")

	klog.Info(string(jsonOutput))

	path, save := os.LookupEnv("STARTUP_LOGS_PATH")
	if save {
		return os.WriteFile(path, jsonOutput, 0644)
	}

	return nil
}
