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

type MicroshiftData struct {
	Start         time.Time     `json:"start"`
	ServicesStart time.Time     `json:"servicesStart"`
	Ready         time.Time     `json:"ready"`
	TimeToReady   time.Duration `json:"timeToReady"`
}

type StartupData struct {
	Services   []ServiceData  `json:"services"`
	Microshift MicroshiftData `json:"microshift"`
}

type StartupRecorder struct {
	Data StartupData

	ServiceCount int
	AllLogged    chan struct{}
	m            sync.Mutex
}

func New() *StartupRecorder {
	return &StartupRecorder{
		AllLogged: make(chan struct{}),
	}
}

func (l *StartupRecorder) ServiceReady(serviceName string, dependencies []string, start time.Time) {
	ready := time.Now()
	timeToReady := ready.Sub(start)

	serviceData := ServiceData{
		Name:         serviceName,
		Dependencies: dependencies,
		Start:        start,
		Ready:        ready,
		TimeToReady:  timeToReady,
	}

	klog.InfoS("SERVICE READY", "service", serviceName, "since-start", timeToReady)

	l.m.Lock()
	defer l.m.Unlock()
	l.Data.Services = append(l.Data.Services, serviceData)
	l.ServiceCount--
	if l.ServiceCount == 0 {
		close(l.AllLogged)
	}
}

func (l *StartupRecorder) MicroshiftStarts(start time.Time) {
	klog.InfoS("MICROSHIFT STARTING")
	l.Data.Microshift.Start = start
}

func (l *StartupRecorder) MicroshiftReady() {
	ready := time.Now()

	klog.InfoS("MICROSHIFT READY", "since-start", time.Since(l.Data.Microshift.Start))
	l.Data.Microshift.Ready = ready
	l.Data.Microshift.TimeToReady = ready.Sub(l.Data.Microshift.Start)

	l.OutputData()
}

func (l *StartupRecorder) ServicesStart(start time.Time) {
	klog.InfoS("MICROSHIFT STARTING SERVICES", "since-start", time.Since(start))
	l.Data.Microshift.ServicesStart = start
}

func (l *StartupRecorder) OutputData() {
	go func() {
		<-l.AllLogged

		jsonOutput, err := json.Marshal(l.Data)
		if err != nil {
			klog.Error("Failed to marshal startup data")
		}

		klog.Infof("Startup data: %s", string(jsonOutput))

		path, ok := os.LookupEnv("STARTUP_LOGS_PATH")
		if ok {
			err = os.WriteFile(path, jsonOutput, 0600)
			if err != nil {
				klog.Error("Failed to write startup data to file")
			}
		}
	}()

}
