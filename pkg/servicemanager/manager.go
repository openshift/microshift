package servicemanager

import (
	"context"
	"errors"
	"fmt"
	"syscall"
	"time"

	"github.com/openshift/microshift/pkg/util/sigchannel"
	"github.com/openshift/microshift/pkg/util/startuplogger"
	"k8s.io/klog/v2"
)

type ServiceManager struct {
	name string
	deps []string

	services   []Service
	serviceMap map[string]Service
}

func NewServiceManager() *ServiceManager {
	return &ServiceManager{
		name: "service-manager",
		deps: []string{},

		services:   []Service{},
		serviceMap: make(map[string]Service),
	}
}
func (s *ServiceManager) Name() string           { return s.name }
func (s *ServiceManager) Dependencies() []string { return s.deps }

func (m *ServiceManager) AddService(s Service) error {
	if s == nil {
		return fmt.Errorf("service must not be <nil>")
	}
	if _, exists := m.serviceMap[s.Name()]; exists {
		return fmt.Errorf("service '%s' added more than once", s.Name())
	}
	for _, dependency := range s.Dependencies() {
		// Enforce that services can only be added after adding their dependencies,
		// i.e. they'll always remain topology sorted. Should we want to relax this
		// constraint later, we can add topo sorting in the Run() step.
		if _, exists := m.serviceMap[dependency]; !exists {
			return fmt.Errorf("dependecy '%s' of service '%s' not yet defined", dependency, s.Name())
		}
	}

	m.services = append(m.services, s)
	m.serviceMap[s.Name()] = s
	return nil
}

func (m *ServiceManager) Run(ctx context.Context, ready chan<- struct{}, stopped chan<- struct{}, startLog *startuplogger.StartupLogger) error {
	defer close(stopped)

	services := m.services
	// No need for topological sorting here as long as we enforce order while adding services.
	// services, err := m.topoSort(services)
	// if err != nil {
	// 	fmt.Error("error: %v", err)
	// }

	readyMap := make(map[string]<-chan struct{})
	stoppedMap := make(map[string]<-chan struct{})

	for _, service := range services {
		// Compile a list of ready channels of the service's dependencies (if any).
		depsReadyList := []<-chan struct{}{}
		for _, dependency := range service.Dependencies() {
			depsReadyList = append(depsReadyList, readyMap[dependency])
		}

		// Wait until all of the service's dependencies signalled readiness.
		// If the context gets canceled before, return immediately.
		select {
		case <-sigchannel.And(depsReadyList):
		case <-ctx.Done():
			// Wait for all services to stop before returning
			// so MicroShift doesn't quit abruptly
			<-sigchannel.And(values(stoppedMap))
			return ctx.Err()
		}

		// Start the service and store its ready and stopped channels
		serviceReady, serviceStopped := m.asyncRun(ctx, service, startLog)
		readyMap[service.Name()] = serviceReady
		stoppedMap[service.Name()] = serviceStopped
	}

	// If we receive readiness signals from all services, signal readiness of manager
	go func() {
		<-sigchannel.And(values(readyMap))
		close(ready)
	}()

	// Stop manager when all services stopped
	<-sigchannel.And(values(stoppedMap))
	return ctx.Err()
}

func (m *ServiceManager) asyncRun(ctx context.Context, service Service, startLog *startuplogger.StartupLogger) (<-chan struct{}, <-chan struct{}) {
	ready, stopped := make(chan struct{}), make(chan struct{})

	klog.WithMicroshiftLoggerComponent(service.Name(), func() {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					klog.Errorf("%s panicked: %s", service.Name(), r)
					klog.Error("Stopping MicroShift")
					if err := syscall.Kill(syscall.Getpid(), syscall.SIGTERM); err != nil {
						klog.Warningf("error killing process: %v", err)
					}
					if !sigchannel.IsClosed(stopped) {
						close(stopped)
					}
				}
			}()

			klog.InfoS("SERVICE STARTING", "service", service.Name())
			svcStart := time.Now()
			go func() {
				<-ready
				klog.InfoS("SERVICE READY", "service", service.Name(), "since-start", time.Since(svcStart))
				startLog.LogService(service.Name(), service.Dependencies(), svcStart, time.Now())
			}()
			go func() {
				<-stopped
				klog.InfoS("SERVICE STOPPED", "service", service.Name(), "since-start", time.Since(svcStart))
			}()

			if err := service.Run(ctx, ready, stopped); err != nil && !errors.Is(err, context.Canceled) {
				klog.ErrorS(err, "SERVICE FAILED - stopping MicroShift", "service", service.Name(), "since-start", time.Since(svcStart))
				if err := syscall.Kill(syscall.Getpid(), syscall.SIGTERM); err != nil {
					klog.Warningf("error killing process: %v", err)
				}
			} else {
				klog.InfoS("SERVICE COMPLETED", "service", service.Name(), "since-start", time.Since(svcStart))
			}
		}()
	})
	return ready, stopped
}

func values(m map[string]<-chan struct{}) []<-chan struct{} {
	values := make([]<-chan struct{}, 0, len(m))
	for _, v := range m {
		values = append(values, v)
	}
	return values
}

//---- topological sorting of directed acyclic graphs via DFS traversal -----

// type markers map[Service]bool

// // Find remaining unmarked nodes and visit them until all nodes are marked.
// func (m *ServiceManager) topoSort(services []Service) ([]Service, error) {
// 	sorted := []Service{}

// 	permanent := make(markers)
// 	temporary := make(markers)

// 	for foundUnmarked := true; foundUnmarked; {
// 		foundUnmarked = false
// 		for _, service := range services {
// 			if !marked(service, permanent) {
// 				if err := m.visit(&sorted, service, permanent, temporary); err != nil {
// 					return nil, err
// 				}
// 				foundUnmarked = true
// 			}
// 		}
// 	}

// 	return sorted, nil
// }

// func mark(service Service, m markers) {
// 	m[service] = true
// }

// func unmark(service Service, m markers) {
// 	delete(m, service)
// }

// func marked(service Service, m markers) bool {
// 	_, ok := m[service]
// 	return ok
// }

// // Recursively visit all of a node's dependencies.
// func (m *ServiceManager) visit(sorted *[]Service, service Service, permanent markers, temporary markers) error {
// 	if marked(service, permanent) {
// 		return nil
// 	}
// 	if marked(service, temporary) {
// 		return fmt.Errorf("detected cyclic dependencies")
// 	}

// 	mark(service, temporary)
// 	for _, name := range service.Dependencies() {
// 		dependency, exists := m.serviceMap[name]
// 		if !exists {
// 			return fmt.Errorf("unknown dependency '%s' for service '%s'", dependency, name)
// 		}
// 		m.visit(sorted, dependency, permanent, temporary)
// 	}
// 	unmark(service, temporary)

// 	mark(service, permanent)
// 	*sorted = append(*sorted, service)

// 	return nil
// }
