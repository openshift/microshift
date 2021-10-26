package cmd

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/coreos/go-systemd/daemon"
	"github.com/openshift/microshift/pkg/components"
	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/controllers"
	"github.com/openshift/microshift/pkg/kustomize"
	"github.com/openshift/microshift/pkg/node"
	"github.com/openshift/microshift/pkg/servicemanager"
	"github.com/openshift/microshift/pkg/util"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	gracefulShutdownTimeout = 60
)

func NewRunMicroshiftCommand() *cobra.Command {
	cfg := config.NewMicroshiftConfig()

	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run MicroShift",
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunMicroshift(cfg, cmd.Flags())
		},
	}

	flags := cmd.Flags()
	// Read the config flag directly into the struct, so it's immediately available.
	flags.StringVar(&cfg.ConfigFile, "config", cfg.ConfigFile, "File to read configuration from.")
	cmd.MarkFlagFilename("config", "yaml", "yml")
	// All other flags will be read after reading both config file and env vars.
	flags.String("data-dir", cfg.DataDir, "Directory for storing runtime data.")
	flags.StringSlice("roles", cfg.Roles, "Roles of this MicroShift instance.")

	return cmd
}

func RunMicroshift(cfg *config.MicroshiftConfig, flags *pflag.FlagSet) error {
	if err := cfg.ReadAndValidate(flags); err != nil {
		logrus.Fatal(err)
	}

	// fail early if we don't have enough privileges
	if config.StringInList("node", cfg.Roles) && os.Geteuid() > 0 {
		logrus.Fatalf("MicroShift must be run privileged for role 'node'")
	}

	os.MkdirAll(cfg.DataDir, 0700)
	os.MkdirAll(cfg.LogDir, 0700)

	// TODO: change to only initialize what is strictly necessary for the selected role(s)
	if _, err := os.Stat(filepath.Join(cfg.DataDir, "certs")); errors.Is(err, os.ErrNotExist) {
		initAll(cfg)
	}

	m := servicemanager.NewServiceManager()
	if config.StringInList("controlplane", cfg.Roles) {
		util.Must(m.AddService(controllers.NewEtcd(cfg)))
		util.Must(m.AddService(controllers.NewKubeAPIServer(cfg)))
		util.Must(m.AddService(controllers.NewKubeScheduler(cfg)))
		util.Must(m.AddService(controllers.NewKubeControllerManager(cfg)))
		// util.Must(m.AddService(controllers.NewOpenShiftPrepJob()))
		// util.Must(m.AddService(controllers.NewOpenShiftAPIServer()))
		util.Must(m.AddService(controllers.NewOpenShiftControllerManager(cfg)))
		// util.Must(m.AddService(controllers.NewOpenShiftAPIComponents()))
		util.Must(m.AddService(controllers.NewOpenShiftOAuth(cfg)))
		// util.Must(m.AddService(controllers.NewInfrastructureServices()))

		util.Must(m.AddService(servicemanager.NewGenericService(
			"other-controlplane",
			[]string{"kube-apiserver"},
			func(ctx context.Context, ready chan<- struct{}, stopped chan<- struct{}) error {
				defer close(stopped)
				defer close(ready)

				startControllerOnly(cfg)

				return nil
			},
		)))
		util.Must(m.AddService(kustomize.NewKustomizer(cfg)))
	}

	if config.StringInList("node", cfg.Roles) {
		util.Must(m.AddService(node.NewKubeletServer(cfg)))
		util.Must(m.AddService(node.NewKubeProxyServer(cfg)))
	}

	logrus.Info("Starting MicroShift")

	ctx, cancel := context.WithCancel(context.Background())
	ready, stopped := make(chan struct{}), make(chan struct{})
	go func() {
		logrus.Infof("Starting %s", m.Name())
		if err := m.Run(ctx, ready, stopped); err != nil {
			logrus.Infof("%s stopped: %s", m.Name(), err)
		} else {
			logrus.Infof("%s completed", m.Name())
		}
	}()

	sigTerm := make(chan os.Signal, 1)
	signal.Notify(sigTerm, os.Interrupt, syscall.SIGTERM)

	select {
	case <-ready:
		logrus.Info("MicroShift is ready.")
		daemon.SdNotify(false, daemon.SdNotifyReady)

		<-sigTerm
	case <-sigTerm:
	}
	logrus.Info("Interrupt received. Stopping services.")
	cancel()

	select {
	case <-stopped:
	case <-sigTerm:
		logrus.Info("Another interrupt received. Force terminating services.")
	case <-time.After(time.Duration(gracefulShutdownTimeout) * time.Second):
		logrus.Info("Timed out waiting for services to stop.")
	}
	logrus.Info("MicroShift stopped.")
	return nil
}

func startControllerOnly(cfg *config.MicroshiftConfig) error {
	if err := controllers.PrepareOCP(cfg); err != nil {
		return err
	}

	logrus.Infof("starting openshift-apiserver")
	controllers.OCPAPIServer(cfg)

	//TODO: cloud provider
	// controllers.OCPControllerManager(cfg)

	if err := controllers.StartOCPAPIComponents(cfg); err != nil {
		return err
	}

	if err := components.StartComponents(cfg); err != nil {
		return err
	}
	return nil
}
