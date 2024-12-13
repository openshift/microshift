package healthcheck

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/openshift/microshift/pkg/config"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/util/homedir"
	"k8s.io/klog/v2"
	"k8s.io/kubectl/pkg/cmd/get"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/utils/ptr"
)

func logPodsAndEvents() {
	cliOptions := genericclioptions.NewConfigFlags(true).WithDeprecatedPasswordFlag()
	cliOptions.KubeConfig = ptr.To(filepath.Join(config.DataDir, "resources", string(config.KubeAdmin), "kubeconfig"))
	if homedir.HomeDir() == "" {
		// By default client writes cache to $HOME/.kube/cache.
		// However, when healthcheck is executed by greenboot, the $HOME is empty,
		// so discovery client wants to write to /.kube which is immutable on ostre
		// causing flood of warnings (and is not elegant to create new root level directory).
		cliOptions.CacheDir = ptr.To(filepath.Join("tmp", ".kube", "cache"))
	}
	matchVersionKubeConfigFlags := cmdutil.NewMatchVersionFlags(cliOptions)
	f := cmdutil.NewFactory(matchVersionKubeConfigFlags)

	output := strings.Builder{}
	ioStreams := genericclioptions.IOStreams{In: os.Stdin, Out: &output, ErrOut: &output}

	cmdGet := get.NewCmdGet("", f, ioStreams)
	opts := get.NewGetOptions("", ioStreams)
	opts.AllNamespaces = true
	opts.PrintFlags.OutputFormat = ptr.To("wide")
	if err := opts.Complete(f, cmdGet, []string{"DUMMY"}); err != nil {
		klog.Errorf("Failed to complete get's options: %v", err)
		return
	}

	if err := opts.Validate(); err != nil {
		klog.Errorf("Failed to validate get's options: %v", err)
		return
	}

	output.WriteString("---------- PODS:\n")
	if err := opts.Run(f, []string{"pods"}); err != nil {
		klog.Errorf("Failed to run 'get pods': %v", err)
		return
	}
	output.WriteString("\n---------- EVENTS:\n")
	opts.SortBy = ".metadata.creationTimestamp"
	if err := opts.Run(f, []string{"events"}); err != nil {
		klog.Errorf("Failed to run 'get events': %v", err)
		return
	}

	klog.Infof("DEBUG INFORMATION\n%s", output.String())
}
