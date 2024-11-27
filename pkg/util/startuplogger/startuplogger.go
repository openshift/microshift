package startuplogger

import (
	"encoding/json"
	"os"
	"time"
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
	services []ServiceData
	//microshiftStart time.Time
	//microshiftReady time.Time
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

func (l *StartupLogger) OutputData() error {
	var output []ServiceData

	output = append(output, l.services...)

	jsonOutput, _ := json.MarshalIndent(output, "", "  ")
	//TODO find suitable export location
	return os.WriteFile("/tmp/service_logs.json", jsonOutput, 0644)

}
