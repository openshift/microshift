package cmd

import (
	"fmt"

	"github.com/openshift/microshift/pkg/admin/history"
	"github.com/openshift/microshift/pkg/admin/system"

	"github.com/spf13/cobra"
)

func NewHealthCommand() *cobra.Command {
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
			currentDeployID, err := system.NewSystemInfo().GetCurrentDeploymentID()
			if err != nil {
				return err
			}

			dhm := history.NewHistoryManager(&history.HistoryFileStorage{})
			fmt.Printf("Updating current boot's health to %s\n\n", health)
			return dhm.Update(history.DeploymentBoot{Boot: *currentBoot, DeploymentID: currentDeployID}, history.BootInfo{Health: health})
		},
	}
	setCurrent.Flags().Bool("healthy", false, "")
	setCurrent.Flags().Bool("unhealthy", false, "")
	setCurrent.MarkFlagsMutuallyExclusive("healthy", "unhealthy")
	health.AddCommand(setCurrent)

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
	health.AddCommand(getCurrent)

	getPrevious := &cobra.Command{
		Use:   "get-previous",
		Short: "Get health of the system during previous boot",
		RunE: func(cmd *cobra.Command, args []string) error {
			return get("previous", system.NewSystemInfo().GetPreviousBoot)
		},
	}
	health.AddCommand(getPrevious)

	return health
}
