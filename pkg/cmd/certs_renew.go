package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"slices"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/openshift/microshift/pkg/admin/certificates"
	certificatesv1alpha1 "github.com/openshift/microshift/pkg/apis/certificates/v1alpha1"
	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/util/cryptomaterial/certchains"
)

type certRenewOptions struct {
	genericclioptions.IOStreams

	serving, ca, dryRun bool
	output              string
	now                 func() time.Time
	loadConfig          func() (*config.Config, error)
	plan                func(*config.Config, bool, time.Time) ([]certchains.CertificateRenewalPlanEntry, error)
	apply               func(*config.Config, bool) (certchains.CertificateInventory, error)
	prepare             func(bool) (func(), error)
}

func newCertsRenewCommand(options *certRenewOptions) *cobra.Command {
	command := &cobra.Command{
		Use:   "renew",
		Short: "Renew managed certificates while MicroShift is stopped",
		Long: "Renew all managed leaf certificates with --serving, or all CAs and descendants with --ca. " +
			"Use --dry-run to validate and preview renewal without changing files, including while MicroShift is running.",
		Args: cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error { return options.run() },
	}
	command.Flags().BoolVar(&options.serving, "serving", false, "Renew all serving, client, and peer certificates without changing CAs")
	command.Flags().BoolVar(&options.ca, "ca", false, "Renew all CAs and their descendant certificates")
	command.Flags().BoolVar(&options.dryRun, "dry-run", false, "Validate and preview renewal without changing files")
	command.Flags().StringVarP(&options.output, "output", "o", "", "Output format. One of: json, yaml.")
	return command
}

func (o *certRenewOptions) run() error {
	if o.ca == o.serving {
		return &certificateCommandError{certificatesv1alpha1.ErrorCodeInvalidArguments,
			fmt.Errorf("specify exactly one of --serving or --ca")}
	}
	if o.output != "" && o.output != certificateOutputJSON && o.output != certificateOutputYAML {
		return &certificateCommandError{certificatesv1alpha1.ErrorCodeInvalidArguments,
			fmt.Errorf("unsupported output format %q; supported formats: json, yaml", o.output)}
	}
	if o.prepare != nil {
		release, err := o.prepare(!o.dryRun)
		if err != nil {
			return err
		}
		defer release()
	}
	cfg, err := o.loadConfig()
	if err != nil {
		return invalidCertificateConfiguration()
	}
	plan, err := o.plan(cfg, o.ca, o.now())
	if err != nil {
		return certificateRenewalError(err)
	}
	var renewed certchains.CertificateInventory
	if !o.dryRun {
		renewed, err = o.apply(cfg, o.ca)
		if err != nil {
			return certificateRenewalError(err)
		}
	}
	result, err := newCertificateRenewalResult(plan, renewed, cfg.Warnings, o.ca, o.dryRun, o.now())
	if err != nil {
		return &certificateCommandError{certificatesv1alpha1.ErrorCodeRenewalFailed, err}
	}
	if o.output != "" {
		err = writeCertificateObject(o.Out, result, o.output)
	} else {
		var out bytes.Buffer
		err = writeCertificateRenewalTable(&out, result, renewed)
		if err == nil {
			for _, warning := range result.Warnings {
				if _, err = fmt.Fprintf(o.ErrOut, "WARNING: %s\n", warning); err != nil {
					return &certificateCommandError{certificatesv1alpha1.ErrorCodeInternalError, err}
				}
			}
			_, err = o.Out.Write(out.Bytes())
		}
	}
	if err != nil {
		return &certificateCommandError{certificatesv1alpha1.ErrorCodeInternalError, err}
	}
	return nil
}

func certificateRenewalError(err error) error {
	var commandError *certificateCommandError
	if errors.As(err, &commandError) {
		return err
	}
	var recoveryError *certificates.RecoveryError
	if errors.As(err, &recoveryError) {
		return &certificateCommandError{certificatesv1alpha1.ErrorCodeRecoveryFailed, err}
	}
	var inventoryError *certchains.InventoryReadError
	if errors.As(err, &inventoryError) {
		return &certificateCommandError{certificatesv1alpha1.ErrorCodeCertificateInventoryFailed, err}
	}
	return &certificateCommandError{certificatesv1alpha1.ErrorCodeRenewalFailed, err}
}

func newCertificateRenewalResult(plan []certchains.CertificateRenewalPlanEntry, renewed certchains.CertificateInventory,
	warnings []string, renewCAs, dryRun bool, now time.Time,
) (*certificatesv1alpha1.CertificateRenewalResult, error) {
	if len(plan) == 0 {
		return nil, fmt.Errorf("no certificates selected for renewal")
	}
	mode, state := certificatesv1alpha1.RenewalModeServing, certificatesv1alpha1.RenewalStatusCompleted
	if renewCAs {
		mode = certificatesv1alpha1.RenewalModeCA
	}
	if dryRun {
		state = certificatesv1alpha1.RenewalStatusValidated
	}
	result := &certificatesv1alpha1.CertificateRenewalResult{
		TypeMeta:    metav1.TypeMeta{APIVersion: certificatesv1alpha1.APIVersion, Kind: certificatesv1alpha1.CertificateRenewalResultKind},
		GeneratedAt: metav1.NewTime(now.UTC().Truncate(time.Second)),
		Mode:        mode, Status: state, DryRun: dryRun,
		Items: make([]certificatesv1alpha1.CertificateRenewalItem, 0, len(plan)),
		Impact: certificatesv1alpha1.CertificateRenewalImpact{
			ServiceRestartRequired: true, KubeconfigRedistributionRequired: renewCAs, ApplicationReloadMayBeRequired: true,
		},
		Warnings: append([]string{}, warnings...),
	}
	for _, entry := range plan {
		expiry := entry.NewNotAfter
		if !dryRun {
			found := false
			for _, actual := range renewed {
				if slices.Equal(actual.Path, entry.Path) {
					expiry, found = actual.Certificate.NotAfter, true
					break
				}
			}
			if !found {
				return nil, fmt.Errorf("renewed certificate %q is missing", entry.Name)
			}
		}
		var parent *string
		if entry.ParentCA != "" {
			parent = &entry.ParentCA
		}
		result.Items = append(result.Items, certificatesv1alpha1.CertificateRenewalItem{
			Service: entry.Service, Name: entry.Name, Role: certificatesv1alpha1.CertificateRole(entry.Role),
			ParentCA: parent, CurrentNotAfter: metav1.NewTime(entry.Certificate.NotAfter.UTC()),
			NewNotAfter: metav1.NewTime(expiry.UTC()), Changed: !dryRun,
		})
	}
	if renewCAs {
		result.Warnings = append(result.Warnings, "Kubeconfigs stored outside the MicroShift data directory must be copied again after renewal.")
	}
	result.Warnings = append(result.Warnings, "Applications that cache certificates or CA bundles may need to be reloaded or restarted after renewal.")
	return result, nil
}

func writeCertificateRenewalTable(out io.Writer, result *certificatesv1alpha1.CertificateRenewalResult, renewed certchains.CertificateInventory) error {
	if result.DryRun {
		if err := writeCertificateRenewalPlan(out, result); err != nil {
			return err
		}
	} else {
		if err := writeCompletedCertificateRenewal(out, result, renewed); err != nil {
			return err
		}
	}
	if result.Mode == certificatesv1alpha1.RenewalModeServing {
		if _, err := fmt.Fprintln(out, "CA certificates are unchanged by leaf renewal."); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintln(out, "MicroShift must be stopped for renewal. Start microshift.service after renewal completes."); err != nil {
		return err
	}
	return nil
}

func writeCompletedCertificateRenewal(out io.Writer, result *certificatesv1alpha1.CertificateRenewalResult, renewed certchains.CertificateInventory) error {
	if _, err := fmt.Fprintf(out, "Renewed %d certificates.\n", len(result.Items)); err != nil {
		return err
	}
	status, err := newCertificateStatusList(renewed, nil, result.GeneratedAt.Time)
	if err != nil {
		return err
	}
	return writeCertificateStatusTable(out, status)
}

func writeCertificateRenewalPlan(out io.Writer, result *certificatesv1alpha1.CertificateRenewalResult) error {
	if _, err := fmt.Fprintln(out, "DRY RUN: Validated certificate renewal; no certificates or kubeconfigs were changed."); err != nil {
		return err
	}
	w := tabwriter.NewWriter(out, 0, 4, 2, ' ', 0)
	if _, err := fmt.Fprintln(w, "SERVICE\tROLE\tPARENT CA\tCERTIFICATE\tCURRENT EXPIRY\tPROPOSED EXPIRY"); err != nil {
		return err
	}
	for _, item := range result.Items {
		parent := "-"
		if item.ParentCA != nil {
			parent = *item.ParentCA
		}
		if _, err := fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n", item.Service, item.Role, parent, item.Name,
			item.CurrentNotAfter.Format(time.RFC3339), item.NewNotAfter.Format(time.RFC3339)); err != nil {
			return err
		}
	}
	return w.Flush()
}
