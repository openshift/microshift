package cmd

import (
	"errors"
	"os"

	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/node"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func NewRunMicroshiftCommand() *cobra.Command {
	cfg := config.NewMicroshiftConfig()

	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run Microshift",
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunMicroshift(cfg, cmd.Flags())
		},
	}

	flags := cmd.Flags()

	// get the log level
        flags.StringVar(&logLevel, "log-level", "info", "Set the log level for Microshift (default: info)")
        if logLevel == "trace" {
                logrus.SetLevel(logrus.TraceLevel)
        } else if logLevel == "debug" {
                logrus.SetLevel(logrus.DebugLevel)
        } else if logLevel == "warn" {
                logrus.SetLevel(logrus.WarnLevel)
        } else if logLevel == "error" {
                logrus.SetLevel(logrus.ErrorLevel)
        } else if logLevel == "fatal" {
                logrus.SetLevel(logrus.FatalLevel)
        } else if logLevel == "panic" {
                logrus.SetLevel(logrus.PanicLevel)
        } else {
                // default to info
                logrus.SetLevel(logrus.InfoLevel)
        }

	// Read the config flag directly into the struct, so it's immediately available.
	flags.StringVar(&cfg.ConfigFile, "config", cfg.ConfigFile, "File to read configuration from.")
	cmd.MarkFlagFilename("config", "yaml", "yml")
	// All other flags will be read after reading both config file and env vars.
	flags.String("data-dir", cfg.DataDir, "Directory for storing runtime data.")
	flags.StringSlice("roles", cfg.Roles, "Roles of this Microshift instance.")

	return cmd
}

func RunMicroshift(cfg *config.MicroshiftConfig, flags *pflag.FlagSet) error {
	if err := cfg.ReadAndValidate(flags); err != nil {
		logrus.Fatal(err)
	}

	// fail early if we don't have enough privileges
	if config.StringInList("node", cfg.Roles) && os.Geteuid() > 0 {
		logrus.Fatalf("Microshift must be run privileged for role 'node'")
	}

	// if data dir is missing, create and initialize it
	// TODO: change to only initialize what is strictly necessary for the selected role(s)
	if _, err := os.Stat(cfg.DataDir); errors.Is(err, os.ErrNotExist) {
		os.MkdirAll(cfg.DataDir, 0700)
		initAll(cfg)
	}
	// if log dir is missing, create it
	os.MkdirAll(cfg.LogDir, 0700)

	if config.StringInList("controlplane", cfg.Roles) {
		if err := startControllerOnly(cfg); err != nil {
			return err
		}
	}

	if config.StringInList("node", cfg.Roles) {
		if err := node.StartKubelet(cfg); err != nil {
			return err
		}
		if err := node.StartKubeProxy(cfg); err != nil {
			return err
		}
	}

	select {}
}
