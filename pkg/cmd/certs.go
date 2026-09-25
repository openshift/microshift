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
}

// NewCertsCommand creates the certificate administration command family.
func NewCertsCommand(ioStreams genericclioptions.IOStreams) *cobra.Command {
	return newCertsCommand(&certStatusOptions{
		IOStreams:     ioStreams,
		now:           time.Now,
		loadConfig:    config.ActiveConfig,
		loadInventory: loadCertificateInventory,
	}, shouldRunPrivileged)
}

func newCertsCommand(options *certStatusOptions, requirePrivileges func() error) *cobra.Command {
	command := &cobra.Command{
		Use:   "certs",
		Short: "Inspect and manage MicroShift certificates",
		Args:  cobra.NoArgs,
		PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
			if err := requirePrivileges(); err != nil {
				return &certificateCommandError{certificatesv1alpha1.ErrorCodeInsufficientPrivileges, err}
			}
			return nil
		},
	}
	command.SetOut(options.Out)
	command.SetErr(options.ErrOut)
	command.AddCommand(newCertsStatusCommand(options))
	return command
}

func newCertsStatusCommand(options *certStatusOptions) *cobra.Command {
	command := &cobra.Command{
		Use:   "status",
		Short: "Report the status of managed MicroShift certificates",
		Args:  cobra.NoArgs,
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

	cfg, err := o.loadConfig()
	if err != nil {
		// Loader errors can include raw configuration and credentials. Do not
		// retain their contents in errors returned by certificate commands.
		return &certificateCommandError{certificatesv1alpha1.ErrorCodeInvalidConfiguration,
			errors.New("failed to load MicroShift configuration; check /etc/microshift/config.yaml and /etc/microshift/config.d")}
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
		zone, err := entry.ZoneAt(now)
		if err != nil {
			return certificatesv1alpha1.CertificateStatusList{}, err
		}
		items = append(items, certificatesv1alpha1.CertificateStatusItem{
			Service:          entry.Service,
			Name:             entry.Name,
			Role:             certificatesv1alpha1.CertificateRole(entry.Role),
			RotationPolicy:   certificatesv1alpha1.RotationPolicy(entry.RotationPolicy),
			Zone:             certificatesv1alpha1.CertificateZone(zone),
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
			ForceRestartOnRedZone: true,
			ServingValidity:       durationInHours(cryptomaterial.ShortLivedCertificateValidity),
			CAValidity:            durationInHours(cryptomaterial.LongLivedCertificateValidity),
		},
		Items:    items,
		Warnings: statusWarnings,
	}, nil
}

func writeCertificateStatusTable(out io.Writer, status certificatesv1alpha1.CertificateStatusList) error {
	w := tabwriter.NewWriter(out, 0, 4, 2, ' ', 0)
	if _, err := fmt.Fprintln(w, "SERVICE\tCERTIFICATE\tSTATUS\tEXPIRY\tREASON\tMESSAGE"); err != nil {
		return err
	}
	for _, item := range status.Items {
		reason, message := humanCertificateStatus(item, status.GeneratedAt.Time)
		if _, err := fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			item.Service,
			item.Name,
			humanCertificateZone(item.Zone),
			item.NotAfter.UTC().Format(time.RFC3339),
			reason,
			message,
		); err != nil {
			return err
		}
	}
	return w.Flush()
}

func humanCertificateStatus(item certificatesv1alpha1.CertificateStatusItem, now time.Time) (string, string) {
	if now.Before(item.NotBefore.Time) {
		return "NotYetValid", fmt.Sprintf("Valid in %d days", daysUntil(item.NotBefore.Sub(now)))
	}
	if !now.Before(item.NotAfter.Time) {
		return "Expired", fmt.Sprintf("Expired %d days ago", daysUntil(now.Sub(item.NotAfter.Time)))
	}
	remainingDays := daysUntil(item.NotAfter.Sub(now))
	if item.Zone == certificatesv1alpha1.CertificateZoneGreen {
		return "NotExpiring", fmt.Sprintf("Valid for %d days", remainingDays)
	}
	return "Expiring", fmt.Sprintf("Expires in %d days", remainingDays)
}

func humanCertificateZone(zone certificatesv1alpha1.CertificateZone) string {
	switch zone {
	case certificatesv1alpha1.CertificateZoneGreen:
		return "Green"
	case certificatesv1alpha1.CertificateZoneYellow:
		return "Yellow"
	case certificatesv1alpha1.CertificateZoneRed:
		return "Red"
	default:
		return string(zone)
	}
}

func daysUntil(duration time.Duration) int64 {
	return int64(math.Ceil(duration.Hours() / 24))
}

func durationInHours(duration time.Duration) string {
	return fmt.Sprintf("%dh", int64(duration.Hours()))
}
