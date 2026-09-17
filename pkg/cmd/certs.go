package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/util/cryptomaterial"
	"github.com/openshift/microshift/pkg/util/cryptomaterial/certchains"
)

const (
	certificateOutputJSON       = "json"
	certificateStatusAPIVersion = "microshift.openshift.io/v1alpha1"
)

type certificateStatusConfig struct {
	ForceRestartOnRedZone bool   `json:"forceRestartOnRedZone"`
	ServingValidity       string `json:"servingValidity"`
	CAValidity            string `json:"caValidity"`
}

type certificateStatusItem struct {
	Service          string                     `json:"service"`
	Name             string                     `json:"name"`
	Role             certchains.CertificateRole `json:"role"`
	RotationPolicy   certchains.RotationPolicy  `json:"rotationPolicy"`
	Zone             certchains.CertificateZone `json:"zone"`
	NotBefore        string                     `json:"notBefore"`
	NotAfter         string                     `json:"notAfter"`
	RemainingSeconds int64                      `json:"remainingSeconds"`
	certificate      certchains.CertificateInventoryEntry
}

type certificateStatusList struct {
	APIVersion  string                  `json:"apiVersion"`
	Kind        string                  `json:"kind"`
	GeneratedAt string                  `json:"generatedAt"`
	Config      certificateStatusConfig `json:"config"`
	Items       []certificateStatusItem `json:"items"`
	Warnings    []string                `json:"warnings"`
	generatedAt time.Time
}

type certStatusOptions struct {
	genericclioptions.IOStreams

	output        string
	now           func() time.Time
	loadConfig    func() (*config.Config, error)
	loadInventory func(*config.Config) (certchains.CertificateInventory, error)
}

// NewCertsCommand creates the certificate administration command family.
func NewCertsCommand(ioStreams genericclioptions.IOStreams) *cobra.Command {
	return newCertsCommand(ioStreams, shouldRunPrivileged)
}

func newCertsCommand(ioStreams genericclioptions.IOStreams, requirePrivileges func() error) *cobra.Command {
	command := &cobra.Command{
		Use:   "certs",
		Short: "Inspect and manage MicroShift certificates",
		Args:  cobra.NoArgs,
		PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
			return requirePrivileges()
		},
	}

	options := &certStatusOptions{
		IOStreams:     ioStreams,
		now:           time.Now,
		loadConfig:    config.ActiveConfig,
		loadInventory: loadCertificateInventory,
	}
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
	command.Flags().StringVarP(&options.output, "output", "o", "", "Output format. One of: json.")
	return command
}

func (o *certStatusOptions) run() error {
	if o.output != "" && o.output != certificateOutputJSON {
		return fmt.Errorf("unsupported output format %q; supported formats: json", o.output)
	}

	cfg, err := o.loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load MicroShift configuration: %w", err)
	}
	inventory, err := o.loadInventory(cfg)
	if err != nil {
		return fmt.Errorf("failed to load certificate inventory: %w", err)
	}

	status, err := newCertificateStatusList(inventory, cfg.Warnings, o.now())
	if err != nil {
		return fmt.Errorf("failed to build certificate status: %w", err)
	}
	if o.output == certificateOutputJSON {
		return writeCertificateStatusJSON(o.Out, status)
	}
	for _, warning := range status.Warnings {
		if _, err := fmt.Fprintf(o.ErrOut, "WARNING: %s\n", warning); err != nil {
			return err
		}
	}
	return writeCertificateStatusTable(o.Out, status)
}

func loadCertificateInventory(cfg *config.Config) (certchains.CertificateInventory, error) {
	builder, err := certificateChainsSetup(cfg)
	if err != nil {
		return nil, err
	}
	return builder.LoadInventory()
}

func newCertificateStatusList(inventory certchains.CertificateInventory, warnings []string, now time.Time) (certificateStatusList, error) {
	now = now.UTC().Truncate(time.Second)
	items := make([]certificateStatusItem, 0, len(inventory))
	for _, entry := range inventory {
		if entry.Name == "" {
			return certificateStatusList{}, fmt.Errorf("certificate inventory entry has no name")
		}
		if entry.Service == "" {
			return certificateStatusList{}, fmt.Errorf("certificate %q has no owning service", entry.Name)
		}
		switch entry.Role {
		case certchains.CertificateRoleCA, certchains.CertificateRoleServing, certchains.CertificateRoleClient, certchains.CertificateRolePeer:
		case certchains.CertificateRoleUnknown:
			return certificateStatusList{}, fmt.Errorf("certificate %q has an unknown role", entry.Name)
		default:
			return certificateStatusList{}, fmt.Errorf("certificate %q has invalid role %q", entry.Name, entry.Role)
		}
		zone, err := entry.ZoneAt(now)
		if err != nil {
			return certificateStatusList{}, err
		}
		items = append(items, certificateStatusItem{
			Service:          entry.Service,
			Name:             entry.Name,
			Role:             entry.Role,
			RotationPolicy:   entry.RotationPolicy,
			Zone:             zone,
			NotBefore:        entry.Certificate.NotBefore.UTC().Format(time.RFC3339),
			NotAfter:         entry.Certificate.NotAfter.UTC().Format(time.RFC3339),
			RemainingSeconds: int64(entry.Certificate.NotAfter.Sub(now).Seconds()),
			certificate:      entry,
		})
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].Service != items[j].Service {
			return items[i].Service < items[j].Service
		}
		if items[i].Name != items[j].Name {
			return items[i].Name < items[j].Name
		}
		return strings.Join(items[i].certificate.Path, "/") < strings.Join(items[j].certificate.Path, "/")
	})

	statusWarnings := append([]string(nil), warnings...)
	if statusWarnings == nil {
		statusWarnings = []string{}
	}
	return certificateStatusList{
		APIVersion:  certificateStatusAPIVersion,
		Kind:        "CertificateStatusList",
		GeneratedAt: now.Format(time.RFC3339),
		Config: certificateStatusConfig{
			ForceRestartOnRedZone: true,
			ServingValidity:       durationInHours(cryptomaterial.ShortLivedCertificateValidity),
			CAValidity:            durationInHours(cryptomaterial.LongLivedCertificateValidity),
		},
		Items:       items,
		Warnings:    statusWarnings,
		generatedAt: now,
	}, nil
}

func writeCertificateStatusJSON(out io.Writer, status certificateStatusList) error {
	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	return encoder.Encode(status)
}

func writeCertificateStatusTable(out io.Writer, status certificateStatusList) error {
	w := tabwriter.NewWriter(out, 0, 4, 2, ' ', 0)
	if _, err := fmt.Fprintln(w, "SERVICE\tCERTIFICATE\tSTATUS\tEXPIRY\tREASON\tMESSAGE"); err != nil {
		return err
	}
	for _, item := range status.Items {
		reason, message := humanCertificateStatus(item.certificate, item.Zone, status.generatedAt)
		if _, err := fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			item.Service,
			item.Name,
			humanCertificateZone(item.Zone),
			item.NotAfter,
			reason,
			message,
		); err != nil {
			return err
		}
	}
	return w.Flush()
}

func humanCertificateStatus(certificate certchains.CertificateInventoryEntry, zone certchains.CertificateZone, now time.Time) (string, string) {
	if now.Before(certificate.Certificate.NotBefore) {
		return "NotYetValid", fmt.Sprintf("Valid in %d days", daysUntil(certificate.Certificate.NotBefore.Sub(now)))
	}
	if !now.Before(certificate.Certificate.NotAfter) {
		return "Expired", fmt.Sprintf("Expired %d days ago", daysUntil(now.Sub(certificate.Certificate.NotAfter)))
	}
	remainingDays := daysUntil(certificate.Certificate.NotAfter.Sub(now))
	if zone == certchains.CertificateZoneGreen {
		return "NotExpiring", fmt.Sprintf("Valid for %d days", remainingDays)
	}
	return "Expiring", fmt.Sprintf("Expires in %d days", remainingDays)
}

func humanCertificateZone(zone certchains.CertificateZone) string {
	switch zone {
	case certchains.CertificateZoneGreen:
		return "Green"
	case certchains.CertificateZoneYellow:
		return "Yellow"
	case certchains.CertificateZoneRed:
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
