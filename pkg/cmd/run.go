package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/coreos/go-systemd/daemon"
	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/controllers"
	"github.com/openshift/microshift/pkg/kustomize"
	"github.com/openshift/microshift/pkg/mdns"
	"github.com/openshift/microshift/pkg/node"
	"github.com/openshift/microshift/pkg/servicemanager"
	"github.com/openshift/microshift/pkg/sysconfwatch"
	"github.com/openshift/microshift/pkg/util"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"k8s.io/klog/v2"
)

const (
	gracefulShutdownTimeout = 60
)

func addRunFlags(cmd *cobra.Command, cfg *config.MicroshiftConfig) {
	flags := cmd.Flags()
	// All other flags will be read after reading both config file and env vars.
	flags.String("node-name", cfg.NodeName, "The hostname of the node.")
	flags.String("node-ip", cfg.NodeIP, "The IP address of the node.")
	flags.String("url", cfg.Cluster.URL, "The URL of the API server.")
	flags.String("cluster-cidr", cfg.Cluster.ClusterCIDR, "The IP range in CIDR notation for pods in the cluster.")
	flags.String("service-cidr", cfg.Cluster.ServiceCIDR, "The IP range in CIDR notation for services in the cluster.")
	flags.String("service-node-port-range", cfg.Cluster.ServiceNodePortRange, "The port range to reserve for services with NodePort visibility. This must not overlap with the ephemeral port range on nodes.")
	flags.String("cluster-domain", cfg.Cluster.Domain, "Domain for this cluster.")
}

func NewRunMicroshiftCommand() *cobra.Command {
	cfg := config.NewMicroshiftConfig()

	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run MicroShift",
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunMicroshift(cfg, cmd.Flags())
		},
	}

	addRunFlags(cmd, cfg)

	return cmd
}

func RunMicroshift(cfg *config.MicroshiftConfig, flags *pflag.FlagSet) error {
	if err := cfg.ReadAndValidate(config.GetConfigFile(), flags); err != nil {
		klog.Fatalf("Error in reading and validating flags: %v", err)
	}

	// fail early if we don't have enough privileges
	if os.Geteuid() > 0 {
		klog.Fatalf("MicroShift must be run privileged")
	}

	// TO-DO: When multi-node is ready, we need to add the controller host-name/mDNS hostname
	//        or VIP to this list on start
	//        see https://github.com/openshift/microshift/pull/471

	if err := util.AddToNoProxyEnv(
		cfg.NodeIP,
		cfg.NodeName,
		cfg.Cluster.ClusterCIDR,
		cfg.Cluster.ServiceCIDR,
		".svc",
		"."+cfg.Cluster.Domain); err != nil {
		klog.Fatal(err)
	}

	os.MkdirAll(microshiftDataDir, 0700)

	// TODO: change to only initialize what is strictly necessary for the selected role(s)
	if err := initAll(cfg); err != nil {
		klog.Fatalf("failed to retrieve the necessary certificates: %v", err)
	}

	m := servicemanager.NewServiceManager()
	util.Must(m.AddService(controllers.NewEtcd(cfg)))
	util.Must(m.AddService(sysconfwatch.NewSysConfWatchController(cfg)))
	util.Must(m.AddService(controllers.NewKubeAPIServer(cfg)))
	util.Must(m.AddService(controllers.NewKubeScheduler(cfg)))
	util.Must(m.AddService(controllers.NewKubeControllerManager(cfg)))
	util.Must(m.AddService(controllers.NewOpenShiftCRDManager(cfg)))
	util.Must(m.AddService(controllers.NewRouteControllerManager(cfg)))
	util.Must(m.AddService(controllers.NewClusterPolicyController(cfg)))
	util.Must(m.AddService(controllers.NewOpenShiftDefaultSCCManager(cfg)))
	util.Must(m.AddService(mdns.NewMicroShiftmDNSController(cfg)))
	util.Must(m.AddService(controllers.NewInfrastructureServices(cfg)))
	util.Must(m.AddService((controllers.NewVersionManager((cfg)))))
	util.Must(m.AddService(kustomize.NewKustomizer(cfg)))
	util.Must(m.AddService(node.NewKubeletServer(cfg)))

	// Storing and clearing the env, so other components don't send the READY=1 until MicroShift is fully ready
	notifySocket := os.Getenv("NOTIFY_SOCKET")
	os.Unsetenv("NOTIFY_SOCKET")

	klog.Infof("Starting MicroShift")

	ctx, cancel := context.WithCancel(context.Background())
	ready, stopped := make(chan struct{}), make(chan struct{})
	go func() {
		klog.Infof("Started %s", m.Name())
		if err := m.Run(ctx, ready, stopped); err != nil {
			klog.Errorf("Stopped %s: %v", m.Name(), err)
		} else {
			klog.Infof("%s completed", m.Name())

		}
	}()

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

		<-sigTerm
	case <-sigTerm:
	}
	klog.Infof("Interrupt received. Stopping services")
	cancel()

	select {
	case <-stopped:
	case <-sigTerm:
		klog.Infof("Another interrupt received. Force terminating services")
	case <-time.After(time.Duration(gracefulShutdownTimeout) * time.Second):
		klog.Infof("Timed out waiting for services to stop")
	}
	klog.Infof("MicroShift stopped")
	return nil
}
