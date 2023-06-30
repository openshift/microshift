package prerun

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/controllers"
	"github.com/openshift/microshift/pkg/node"
	"github.com/openshift/microshift/pkg/servicemanager"
	"github.com/openshift/microshift/pkg/util"

	"k8s.io/klog/v2"
)

const (
	gracefulShutdownTimeout = time.Second * 30
	readyStateTimeout       = time.Second * 10
)

// runMinimalMicroshift initializes a minimum run of microshift core services during pre-run
// to help with storage migration.
//
// Returns a handler to shutdown MicroShfit services and error if services never became ready.
func runMinimalMicroshift(cfg *config.Config) (context.CancelFunc, error) {
	// Establish the context we will use to control execution
	runCtx, runCancel := context.WithCancel(context.Background())
	m := servicemanager.NewServiceManager()
	util.Must(m.AddService(node.NewNetworkConfiguration(cfg)))
	util.Must(m.AddService(controllers.NewEtcd(cfg)))
	util.Must(m.AddService(controllers.NewKubeAPIServer(cfg)))

	// Start core services up
	ready, stopped := make(chan struct{}), make(chan struct{})
	go func() {
		klog.Infof("Started %s", m.Name())
		if err := m.Run(runCtx, ready, stopped); err != nil {
			klog.Errorf("Stopped %s: %v", m.Name(), err)
		} else {
			klog.Infof("%s completed", m.Name())
		}
	}()

	// Connect os signal handler
	sigTerm := make(chan os.Signal, 1)
	signal.Notify(sigTerm, os.Interrupt, syscall.SIGTERM)

	go func() {
		select {
		case <-sigTerm:
			klog.Info("Interrupt received")
			klog.Info("Stopping services")
			runCancel()
		case <-runCtx.Done():
			klog.Info("Done signal from context received")
		}
	}()

	// Provides a function to call when you want to stop the services
	// because this is a blocking call it serves as a call and wait for
	// all the services to stop.
	stopFunc := func() {
		runCancel()
		select {
		case <-stopped:
		case <-time.After(gracefulShutdownTimeout):
			klog.Infof("Timed out waiting for services to stop")
		}
		klog.Infof("MicroShift Sevices stopped")
	}

	// If Microshift does not become ready by the deadline, we consider that an error
	select {
	case <-ready:
		klog.Infof("MicroShift is ready in minimal state")
	case <-time.After(readyStateTimeout):
		klog.Error("MicroShift never became ready in minimal state")
		stopFunc()
		return nil, fmt.Errorf("services failed to be in ready state by timeout(%s)", readyStateTimeout)
	}

	return stopFunc, nil
}
