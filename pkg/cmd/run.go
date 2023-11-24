package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/coreos/go-systemd/daemon"
	"github.com/openshift/microshift/pkg/admin/data"
	"github.com/openshift/microshift/pkg/admin/prerun"
	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/controllers"
	"github.com/openshift/microshift/pkg/kustomize"
	"github.com/openshift/microshift/pkg/loadbalancerservice"
	"github.com/openshift/microshift/pkg/mdns"
	"github.com/openshift/microshift/pkg/node"
	"github.com/openshift/microshift/pkg/release"
	"github.com/openshift/microshift/pkg/servicemanager"
	"github.com/openshift/microshift/pkg/sysconfwatch"
	"github.com/openshift/microshift/pkg/util"
	"github.com/openshift/microshift/pkg/util/cryptomaterial/certchains"
	"github.com/openshift/microshift/pkg/version"
	"github.com/spf13/cobra"

	logsAPIV1 "k8s.io/component-base/logs/api/v1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/yaml"
)

const (
	gracefulShutdownTimeout = 15
)

var (
	preRunFailedLogPath = util.LogFilePath(filepath.Join(config.BackupsDir, "prerun_failed.log"))
	cleanUpFileLogPaths = []util.LogFilePath{
		preRunFailedLogPath,
	}
)

func NewRunMicroshiftCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run MicroShift",
	}

	var multinode bool

	flags := cmd.Flags()
	flags.BoolVar(&multinode, "multinode", false, "enable multinode mode")
	err := flags.MarkHidden("multinode")
	if err != nil {
		panic(err)
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		versionInfo := version.Get()
		klog.InfoS("Version", "microshift", versionInfo.String(), "base", release.Base)

		cfg, err := config.ActiveConfig()
		if err != nil {
			return err
		}

		// `v` is a flag registered in klog's init()
		if vFlag := flags.Lookup("v"); vFlag != nil {
			verbosity := strconv.Itoa(cfg.GetVerbosity())
			if err := vFlag.Value.Set(verbosity); err != nil {
				klog.Errorf("Failed to set log verbosity: %w", err)
			}
		}

		cfg = config.ConfigMultiNode(cfg, multinode)

		for _, w := range cfg.Warnings {
			klog.Warningf("Configuration warning: %s", w)
		}

		// Things to very badly if the node's name has changed
		// since the last time the server started.
		err = cfg.EnsureNodeNameHasNotChanged()
		if err != nil {
			return err
		}
		return RunMicroshift(cfg)
	}

	return cmd
}

func cleanUpPreviousLogFiles() {
	for _, p := range cleanUpFileLogPaths {
		if errLog := p.Remove(); errLog != nil {
			klog.ErrorS(errLog, "Failed to remove log file", "path", p)
		}
	}
}

func writeLogFileError(l util.LogFilePath, err error) {
	if errLog := l.Write([]byte(err.Error())); errLog != nil {
		klog.ErrorS(errLog, "Failed to write log file", "path", l)
	}
}

func logConfig(cfg *config.Config) {
	marshalled, err := yaml.Marshal(cfg)
	if err != nil {
		klog.Fatal(err)
	}
	klog.Info("Effective configuration:")
	for _, line := range strings.Split(string(marshalled), "\n") {
		klog.Info(line)
	}
}

func prerunDataManagement() error {
	dataManager, err := data.NewManager(config.BackupsDir)
	if err != nil {
		return fmt.Errorf("failed to create data manager: %w", err)
	}

	return prerun.DataManagement(dataManager)
}

func RunMicroshift(cfg *config.Config) error {
	// fail early if we don't have enough privileges
	if os.Geteuid() > 0 {
		klog.Fatalf("MicroShift must be run privileged")
	}

	// Tell the logging code that it's OK to receive reconfiguration
	// instructions unless those instructions are different. This
	// overrides the default behavior of erroring out if any
	// instructions are received a second time. That default behavior
	// causes issues with embedded components that all use the same
	// logging code and are not expecting to be compiled together into
	// the same process, as they are in MicroShift. See comments in
	// k8s.io/component-base/logs/api/v1/options.go for details.
	logsAPIV1.ReapplyHandling = logsAPIV1.ReapplyHandlingIgnoreUnchanged

	cleanUpPreviousLogFiles()

	if err := prerunDataManagement(); err != nil {
		writeLogFileError(preRunFailedLogPath, err)
		return err
	}

	logConfig(cfg)

	// TO-DO: When multi-node is ready, we need to add the controller host-name/mDNS hostname
	//        or VIP to this list on start
	//        see https://github.com/openshift/microshift/pull/471

	if err := util.AddToNoProxyEnv(
		cfg.Node.NodeIP,
		cfg.Node.HostnameOverride,
		cfg.Network.ClusterNetwork[0],
		cfg.Network.ServiceNetwork[0],
		".svc",
		".cluster.local",
		"."+cfg.DNS.BaseDomain); err != nil {
		klog.Fatal(err)
	}

	if err := util.MakeDir(config.DataDir); err != nil {
		return fmt.Errorf("failed to create dir %q: %w", config.DataDir, err)
	}

	if err := prerun.VersionMetadataManagement(); err != nil {
		writeLogFileError(preRunFailedLogPath, err)
		return err
	}

	// TODO: change to only initialize what is strictly necessary for the selected role(s)
	certChains, err := initCerts(cfg)
	if err != nil {
		klog.Fatalf("failed to retrieve the necessary certificates: %v", err)
	}

	// create kubeconfig for kube-scheduler, kubelet,controller-manager
	if err := initKubeconfigs(cfg, certChains); err != nil {
		klog.Fatalf("failed to create the necessary kubeconfigs for internal components: %v", err)
	}

	// Establish the context we will use to control execution
	runCtx, runCancel := context.WithCancel(context.Background())

	m := servicemanager.NewServiceManager()
	util.Must(m.AddService(node.NewNetworkConfiguration(cfg)))
	util.Must(m.AddService(controllers.NewEtcd(cfg)))
	util.Must(m.AddService(sysconfwatch.NewSysConfWatchController(cfg)))
	util.Must(m.AddService(controllers.NewKubeAPIServer(cfg)))
	util.Must(m.AddService(controllers.NewKubeScheduler(cfg)))
	util.Must(m.AddService(controllers.NewKubeControllerManager(runCtx, cfg)))
	util.Must(m.AddService(controllers.NewOpenShiftCRDManager(cfg)))
	util.Must(m.AddService(controllers.NewRouteControllerManager(cfg)))
	util.Must(m.AddService(controllers.NewOpenShiftDefaultSCCManager(cfg)))
	util.Must(m.AddService(mdns.NewMicroShiftmDNSController(cfg)))
	util.Must(m.AddService(controllers.NewInfrastructureServices(cfg)))
	util.Must(m.AddService(controllers.NewClusterPolicyController(cfg)))
	util.Must(m.AddService(controllers.NewVersionManager(cfg)))
	util.Must(m.AddService(kustomize.NewKustomizer(cfg)))
	util.Must(m.AddService(node.NewKubeletServer(cfg)))
	util.Must(m.AddService(loadbalancerservice.NewLoadbalancerServiceController(cfg)))
	util.Must(m.AddService(controllers.NewKubeStorageVersionMigrator(cfg)))

	// Storing and clearing the env, so other components don't send the READY=1 until MicroShift is fully ready
	notifySocket := os.Getenv("NOTIFY_SOCKET")
	os.Unsetenv("NOTIFY_SOCKET")

	klog.Infof("Starting MicroShift")

	_, rotationDate, err := certchains.WhenToRotateAtEarliest(certChains)
	if err != nil {
		klog.Fatalf("failed to determine when to rotate certificates: %v", err)
	}

	// Establish a deadline for restarting to rotate the certificates.
	certCtx, certCancel := context.WithDeadline(context.Background(), rotationDate)

	// Watch for the certificate deadline context to be done, log a
	// message, and cancel the run context to propagate the shutdown.
	go func() {
		select {
		case <-certCtx.Done():
			klog.Info("Stopping services for certificate rotation")
			runCancel()
			return
		case <-runCtx.Done():
			klog.Info("Certificate watcher exiting")
			certCancel()
			return
		}
	}()

	// Start everything up
	ready, stopped := make(chan struct{}), make(chan struct{})
	go func() {
		klog.Infof("Started %s", m.Name())
		if err := m.Run(runCtx, ready, stopped); err != nil {
			klog.Errorf("Stopped %s: %v", m.Name(), err)
		} else {
			klog.Infof("%s completed", m.Name())
		}
	}()

	// Connect signal handler
	sigTerm := make(chan os.Signal, 1)
	signal.Notify(sigTerm, os.Interrupt, syscall.SIGTERM)

	select {
	case <-ready:
		klog.Infof("MicroShift is ready")
		os.Setenv("NOTIFY_SOCKET", notifySocket)
		if supported, err := daemon.SdNotify(false, daemon.SdNotifyReady); err != nil {
			klog.Warningf("error sending sd_notify readiness message: %v", err)
		} else if supported {
			klog.Info("sent sd_notify readiness message")
		} else {
			klog.Info("service does not support sd_notify readiness messages")
		}

		// Watch for SIGTERM to exit, now that we are ready.
		<-sigTerm
		klog.Info("Interrupt received")
	case <-sigTerm:
		// A signal that comes in before we are ready is handled here.
		klog.Info("Interrupt received")
	case <-runCtx.Done():
		// We might end up here if the certificate rotation is
		// triggered and we exit on our own, instead of via a signal.
	}
	klog.Info("Stopping services")
	runCancel()

	select {
	case <-stopped:
	case <-time.After(time.Duration(gracefulShutdownTimeout) * time.Second):
		klog.Infof("Timed out waiting for services to stop")
	}
	klog.Infof("MicroShift stopped")
	return nil
}
