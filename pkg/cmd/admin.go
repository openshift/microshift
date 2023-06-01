package cmd

import (
	"fmt"
	"time"

	"github.com/openshift/microshift/pkg/admin/data"
	"github.com/openshift/microshift/pkg/admin/history"
	"github.com/openshift/microshift/pkg/admin/system"
	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/version"

	"github.com/spf13/cobra"
)

func backup(cmd *cobra.Command, args []string) error {
	storage := data.StoragePath(cmd.Flag("storage").Value.String())
	name := data.BackupName(cmd.Flag("name").Value.String())

	dataManager, err := data.NewManager(storage)
	if err != nil {
		return err
	}

	if exists, err := dataManager.BackupExists(name); err != nil {
		return err
	} else if exists {
		return fmt.Errorf("backup %s already exists", dataManager.GetBackupPath(name))
	}

	return dataManager.Backup(name)
}

func newAdminDataCommand() *cobra.Command {
	backup := &cobra.Command{
		Use:   "backup",
		Short: "Backup MicroShift data",
		RunE:  backup,
	}

	data := &cobra.Command{
		Use:   "data",
		Short: "Commands for managing MicroShift data",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if cmd.Flag("storage").Value.String() == "" {
				return fmt.Errorf("--storage must not be empty")
			}

			if cmd.Flag("name").Value.String() == "" {
				return fmt.Errorf("--name must not be empty")
			}

			if err := data.MicroShiftIsNotRunning(); err != nil {
				return fmt.Errorf("microshift must not be running: %w", err)
			}

			return nil
		},
	}
	v := version.Get()
	data.PersistentFlags().String("storage", config.BackupsDir, "Directory with backups")
	data.PersistentFlags().String("name",
		fmt.Sprintf("%s.%s__%s", v.Major, v.Minor, time.Now().UTC().Format("20060102_150405")),
		"Backup name")

	data.AddCommand(backup)
	return data
}

func newAdminHealthCommand() *cobra.Command {
	setCurrent := &cobra.Command{
		Use:   "set-current",
		Short: "Store system health of current boot",
		RunE: func(cmd *cobra.Command, args []string) error {
			healthy, err := cmd.Flags().GetBool("healthy")
			if err != nil {
				return err
			}
			unhealthy, err := cmd.Flags().GetBool("unhealthy")
			if err != nil {
				return err
			}

			var health history.Health
			if healthy {
				health = history.Healthy
			} else if unhealthy {
				health = history.Unhealthy
			} else {
				return fmt.Errorf("command requires either --healthy or --unhealthy")
			}

			currentBoot, err := system.NewSystemInfo().GetCurrentBoot()
			if err != nil {
				return err
			}

			dhm := history.NewHistoryManager(&history.HistoryFileStorage{})
			fmt.Printf("Updating current boot's health to %s\n\n", health)
			return dhm.Update(*currentBoot, history.BootInfo{Health: health})
		},
	}
	setCurrent.Flags().Bool("healthy", false, "")
	setCurrent.Flags().Bool("unhealthy", false, "")
	setCurrent.MarkFlagsMutuallyExclusive("healthy", "unhealthy")

	get := func(friendlyName string, getter func() (*system.Boot, error)) error {
		bootInfo, err := getter()
		if err != nil {
			return err
		}
		if bootInfo == nil {
			return fmt.Errorf("no %s boot info", friendlyName)
		}

		history, err := history.NewHistoryManager(&history.HistoryFileStorage{}).Get()
		if err != nil {
			return err
		}

		boot, found := history.GetBootByID(bootInfo.ID)
		if !found {
			return fmt.Errorf("health for %s boot is not present in history", friendlyName)
		}

		fmt.Println(boot.Health)
		return nil
	}

	getCurrent := &cobra.Command{
		Use:   "get-current",
		Short: "Get health of the system during current boot",
		RunE: func(cmd *cobra.Command, args []string) error {
			return get("current", system.NewSystemInfo().GetCurrentBoot)
		},
	}

	getPrevious := &cobra.Command{
		Use:   "get-previous",
		Short: "Get health of the system during previous boot",
		RunE: func(cmd *cobra.Command, args []string) error {
			return get("previous", system.NewSystemInfo().GetPreviousBoot)
		},
	}

	health := &cobra.Command{
		Use:    "health",
		Hidden: true,
		Short:  "Internal commands used for managing system health information (ostree-based systems only)",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if isOstree, err := system.NewSystemInfo().IsOSTree(); err != nil {
				return err
			} else if !isOstree {
				return fmt.Errorf("this command is only for ostree-based systems")
			}
			return nil
		},
	}
	health.AddCommand(setCurrent)
	health.AddCommand(getCurrent)
	health.AddCommand(getPrevious)
	return health
}

func NewAdminCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "admin",
		Short: "Commands for managing MicroShift",
	}
	cmd.AddCommand(newAdminDataCommand())
	cmd.AddCommand(newAdminHealthCommand())

	return cmd
}
