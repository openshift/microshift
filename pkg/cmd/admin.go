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

func newAdminOstreeCommand() *cobra.Command {
	setHealth := &cobra.Command{
		Use:       "set-health { healthy | unhealthy }",
		Short:     "Persist health of current deployment",
		Args:      cobra.ExactArgs(1),
		ValidArgs: []string{"healthy", "unhealthy"},
		RunE: func(cmd *cobra.Command, args []string) error {
			health, err := history.StringToHealth(args[0])
			if err != nil {
				cmd.PrintErrf("Error: %v\n\n", err)
				return cmd.Help()
			}
			currentBoot, err := system.NewSystemInfo().GetCurrentBoot()
			if err != nil {
				return err
			}

			dhm := history.NewHistoryManager(&history.HistoryFileStorage{})
			return dhm.Update(*currentBoot, history.BootInfo{Health: health})
		},
	}

	getHealth := &cobra.Command{
		Use:       "get-health { current | previous }",
		Short:     "Get health of current or previous boot",
		Args:      cobra.ExactArgs(1),
		ValidArgs: []string{"current", "previous"},
		RunE: func(cmd *cobra.Command, args []string) error {
			var bootInfo *system.Boot
			found := true
			var err error

			sys := system.NewSystemInfo()
			switch args[0] {
			case "current":
				bootInfo, err = sys.GetCurrentBoot()
			case "previous":
				bootInfo, err = sys.GetPreviousBoot()
			default:
				cmd.PrintErrf("Error: invalid argument: %s\n\n", args[0])
				return cmd.Help()
			}
			if err != nil {
				return err
			}
			if !found {
				return fmt.Errorf("no %s boot", args[0])
			}

			dh := history.NewHistoryManager(&history.HistoryFileStorage{})
			history, err := dh.Get()
			if err != nil {
				return err
			}

			bh, found := history.GetBootByID(bootInfo.ID)
			if !found {
				return fmt.Errorf("health for %s boot is not present in history", args[0])
			}
			fmt.Println(bh.Health)
			return nil
		},
	}

	ostree := &cobra.Command{
		Use:   "ostree",
		Short: "Internal commands used in ostree-based systems",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			sys := system.NewSystemInfo()
			if isOstree, err := sys.IsOSTree(); err != nil {
				return err
			} else if !isOstree {
				return fmt.Errorf("this command is only for ostree-based systems")
			}
			return nil
		},
	}
	ostree.AddCommand(getHealth)
	ostree.AddCommand(setHealth)
	return ostree
}

func NewAdminCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "admin",
		Short: "Commands for managing MicroShift",
	}
	cmd.AddCommand(newAdminDataCommand())
	cmd.AddCommand(newAdminOstreeCommand())

	return cmd
}
