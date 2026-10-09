package cmd

import (
	"errors"
	"fmt"
	"io"
	"math"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	certificatesv1alpha1 "github.com/openshift/microshift/pkg/apis/certificates/v1alpha1"
	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/util/cryptomaterial"
	"github.com/openshift/microshift/pkg/util/cryptomaterial/certchains"
)

const (
	certificateOutputJSON = "json"
	certificateOutputYAML = "yaml"
)

type certStatusOptions struct {
	genericclioptions.IOStreams

	output        string
	now           func() time.Time
	loadConfig    func() (*config.Config, error)
	loadInventory func(*config.Config) (certchains.CertificateInventory, error)
	loadPending   func() (*certificatesv1alpha1.CertificateRenewalResult, error)
	prepare       func(bool) (func(), error)
}

// NewCertsCommand creates the certificate administration command family.
func NewCertsCommand(ioStreams genericclioptions.IOStreams) *cobra.Command {
	command := newCertsCommand(&certStatusOptions{
		IOStreams:     ioStreams,
		now:           time.Now,
		loadConfig:    config.ActiveConfig,
		loadInventory: loadCertificateInventory,
		prepare:       prepareCertificateAccess,
		loadPending: func() (*certificatesv1alpha1.CertificateRenewalResult, error) {
			return loadPendingCertificateRenewal(config.DataDir)
		},
	}, shouldRunPrivileged)
	command.AddCommand(newCertsRenewCommand(&certRenewOptions{
		IOStreams: ioStreams, now: time.Now, loadConfig: config.ActiveConfig,
		prepare: prepareCertificateAccess,
		plan: func(cfg *config.Config, renewCAs bool, now time.Time) ([]certchains.CertificateRenewalPlanEntry, error) {
			builder, err := certificateChainsSetup(cfg, config.DataDir)
			if err != nil {
				return nil, invalidCertificateConfiguration()
			}
			return builder.PlanRenewal(renewCAs, now)
		},
		apply: func(cfg *config.Config, renewCAs bool) (certchains.CertificateInventory, error) {
			return applyCertificateRenewal(cfg, config.DataDir, renewCAs)
		},
	}, shouldRunPrivileged))
	return command
}

func newCertsCommand(options *certStatusOptions, requirePrivileges func() error) *cobra.Command {
	command := &cobra.Command{
		Use:   "certs",
		Short: "Inspect and manage MicroShift certificates",
		Args:  cobra.NoArgs,
	}
	command.SetOut(options.Out)
	command.SetErr(options.ErrOut)
	command.AddCommand(newCertsStatusCommand(options, requirePrivileges))
	return command
}

// certificatePreRun checks privileges without shadowing root initialization.
// Attach it to each executable certificate subcommand as PreRunE.
func certificatePreRun(requirePrivileges func() error) func(*cobra.Command, []string) error {
	return func(_ *cobra.Command, _ []string) error {
		if err := requirePrivileges(); err != nil {
			return &certificateCommandError{certificatesv1alpha1.ErrorCodeInsufficientPrivileges, err}
		}
		return nil
	}
}

func newCertsStatusCommand(options *certStatusOptions, requirePrivileges func() error) *cobra.Command {
	command := &cobra.Command{
		Use:     "status",
		Short:   "Report the status of managed MicroShift certificates",
		Args:    cobra.NoArgs,
		PreRunE: certificatePreRun(requirePrivileges),
		RunE: func(_ *cobra.Command, _ []string) error {
			return options.run()
		},
	}
	command.Flags().StringVarP(&options.output, "output", "o", "", "Output format. One of: json, yaml.")
	return command
}

func (o *certStatusOptions) run() error {
	if o.output != "" && o.output != certificateOutputJSON && o.output != certificateOutputYAML {
		return &certificateCommandError{certificatesv1alpha1.ErrorCodeInvalidArguments,
			fmt.Errorf("unsupported output format %q; supported formats: json, yaml", o.output)}
	}
	if o.prepare != nil {
		release, err := o.prepare(false)
		if err != nil {
			return err
		}
		defer release()
	}

	cfg, err := o.loadConfig()
	if err != nil {
		// Loader errors can include raw configuration and credentials. Do not
		// retain their contents in errors returned by certificate commands.
		return invalidCertificateConfiguration()
	}
	inventory, err := o.loadInventory(cfg)
	if err != nil {
		return &certificateCommandError{certificatesv1alpha1.ErrorCodeCertificateInventoryFailed,
			fmt.Errorf("failed to load certificate inventory: %w", err)}
	}

	status, err := newCertificateStatusList(inventory, cfg.Warnings, o.now())
	if err != nil {
		return &certificateCommandError{certificatesv1alpha1.ErrorCodeCertificateInventoryFailed,
			fmt.Errorf("failed to build certificate status: %w", err)}
	}
	if o.loadPending != nil {
		status.PendingRenewal, err = o.loadPending()
		if err != nil {
			return &certificateCommandError{certificatesv1alpha1.ErrorCodeCertificateInventoryFailed, err}
		}
		if status.PendingRenewal != nil {
			status.Warnings = append(status.Warnings, "Renewed certificates are pending activation. Status items describe active files; restart MicroShift to activate the pending renewal.")
		}
	}
	switch c := o.output; c {
	case certificateOutputJSON, certificateOutputYAML:
		err = writeCertificateObject(o.Out, &status, c)
	default:
		for _, warning := range status.Warnings {
			if _, err := fmt.Fprintf(o.ErrOut, "WARNING: %s\n", warning); err != nil {
				return &certificateCommandError{certificatesv1alpha1.ErrorCodeInternalError, err}
			}
		}
		err = writeCertificateStatusTable(o.Out, status)
	}
	if err != nil {
		return &certificateCommandError{certificatesv1alpha1.ErrorCodeInternalError, err}
	}
	return nil
}

func invalidCertificateConfiguration() error {
	return &certificateCommandError{certificatesv1alpha1.ErrorCodeInvalidConfiguration,
		errors.New("failed to load MicroShift configuration; check /etc/microshift/config.yaml and /etc/microshift/config.d")}
}

func loadCertificateInventory(cfg *config.Config) (certchains.CertificateInventory, error) {
	builder, err := certificateChainsSetup(cfg, config.DataDir)
	if err != nil {
		return nil, err
	}
	return builder.LoadInventory()
}

func newCertificateStatusList(inventory certchains.CertificateInventory, warnings []string, now time.Time) (certificatesv1alpha1.CertificateStatusList, error) {
	now = now.UTC().Truncate(time.Second)
	// Sort a copy so the inventory retains its chain traversal order.
	inventory = append(certchains.CertificateInventory(nil), inventory...)
	sort.Slice(inventory, func(i, j int) bool {
		if inventory[i].Service != inventory[j].Service {
			return inventory[i].Service < inventory[j].Service
		}
		if inventory[i].Name != inventory[j].Name {
			return inventory[i].Name < inventory[j].Name
		}
		return strings.Join(inventory[i].Path, "/") < strings.Join(inventory[j].Path, "/")
	})
	items := make([]certificatesv1alpha1.CertificateStatusItem, 0, len(inventory))
	for _, entry := range inventory {
		if entry.Name == "" {
			return certificatesv1alpha1.CertificateStatusList{}, fmt.Errorf("certificate inventory entry has no name")
		}
		if entry.Service == "" {
			return certificatesv1alpha1.CertificateStatusList{}, fmt.Errorf("certificate %q has no owning service", entry.Name)
		}
		switch entry.Role {
		case certchains.CertificateRoleCA, certchains.CertificateRoleServing, certchains.CertificateRoleClient, certchains.CertificateRolePeer:
		case certchains.CertificateRoleUnknown:
			return certificatesv1alpha1.CertificateStatusList{}, fmt.Errorf("certificate %q has an unknown role", entry.Name)
		default:
			return certificatesv1alpha1.CertificateStatusList{}, fmt.Errorf("certificate %q has invalid role %q", entry.Name, entry.Role)
		}
		state, err := entry.StatusAt(now)
		if err != nil {
			return certificatesv1alpha1.CertificateStatusList{}, err
		}
		items = append(items, certificatesv1alpha1.CertificateStatusItem{
			Service:          entry.Service,
			Name:             entry.Name,
			Role:             certificatesv1alpha1.CertificateRole(entry.Role),
			RotationPolicy:   certificatesv1alpha1.RotationPolicy(entry.RotationPolicy),
			Status:           certificatesv1alpha1.CertificateStatus(state),
			NotBefore:        metav1.NewTime(entry.Certificate.NotBefore.UTC()),
			NotAfter:         metav1.NewTime(entry.Certificate.NotAfter.UTC()),
			RemainingSeconds: int64(entry.Certificate.NotAfter.Sub(now).Seconds()),
		})
	}

	statusWarnings := append([]string(nil), warnings...)
	if statusWarnings == nil {
		statusWarnings = []string{}
	}
	return certificatesv1alpha1.CertificateStatusList{
		TypeMeta: metav1.TypeMeta{
			APIVersion: certificatesv1alpha1.APIVersion,
			Kind:       certificatesv1alpha1.CertificateStatusListKind,
		},
		GeneratedAt: metav1.NewTime(now),
		Config: certificatesv1alpha1.CertificateStatusConfig{
			ForceRestartOnExpirationImminent: true,
			ServingValidity:                  durationInHours(cryptomaterial.ShortLivedCertificateValidity),
			CAValidity:                       durationInHours(cryptomaterial.LongLivedCertificateValidity),
		},
		Items:    items,
		Warnings: statusWarnings,
	}, nil
}

func writeCertificateStatusTable(out io.Writer, status certificatesv1alpha1.CertificateStatusList) error {
	w := tabwriter.NewWriter(out, 0, 4, 2, ' ', 0)
	if _, err := fmt.Fprintln(w, "SERVICE\tCERTIFICATE\tSTATUS\tEXPIRY\tMESSAGE"); err != nil {
		return err
	}
	for _, item := range status.Items {
		message, err := humanCertificateStatus(item, status.GeneratedAt.Time)
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			item.Service,
			item.Name,
			item.Status,
			item.NotAfter.UTC().Format(time.RFC3339),
			message,
		); err != nil {
			return err
		}
	}
	return w.Flush()
}

func humanCertificateStatus(item certificatesv1alpha1.CertificateStatusItem, now time.Time) (string, error) {
	if !now.Before(item.NotAfter.Time) {
		return fmt.Sprintf("Expired %d days ago", daysUntil(now.Sub(item.NotAfter.Time))), nil
	}
	if now.Before(item.NotBefore.Time) {
		return fmt.Sprintf("Valid in %d days", daysUntil(item.NotBefore.Sub(now))), nil
	}

	remainingDays := daysUntil(item.NotAfter.Sub(now))
	if item.Status == certificatesv1alpha1.CertificateStatusHealthy {
		return fmt.Sprintf("Valid for %d days", remainingDays), nil
	}

	warning, critical, err := certchains.RotationPolicy(item.RotationPolicy).StatusThresholds()
	if err != nil {
		return "", fmt.Errorf("certificate %q has %w", item.Name, err)
	}

	validity := item.NotAfter.Sub(item.NotBefore.Time)
	if item.Status == certificatesv1alpha1.CertificateStatusExpirationImminent {
		criticalDays := daysUntil(time.Duration(float64(validity) * critical))
		return fmt.Sprintf("Expires in %d days, which is at or below the critical threshold of %d days", remainingDays, criticalDays), nil
	}
	warningDays := daysUntil(time.Duration(float64(validity) * warning))
	return fmt.Sprintf("Expires in %d days, which is at or below the warning threshold of %d days", remainingDays, warningDays), nil
}

func daysUntil(duration time.Duration) int64 {
	return int64(math.Ceil(duration.Hours() / 24))
}

func durationInHours(duration time.Duration) string {
	return fmt.Sprintf("%dh", int64(duration.Hours()))
}
