package healthcheck

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/openshift/microshift/pkg/config"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	restclient "k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/rawhttp"
)

func apiServerShouldBeLiveAndReady(ctx context.Context, timeout time.Duration) error {
	client, err := buildClient()
	if err != nil {
		return err
	}

	// If API Server is not ready/live, raw get call generates an error.
	// But for healthcheck, we only care about waiting until timeout for "ok",
	// so we can discard the errors.

	ready := false
	live := false

	klog.Info("Waiting for API Server to be Live and Ready")

	if err := wait.PollUntilContextTimeout(ctx, time.Second, timeout, true, func(ctx context.Context) (done bool, err error) {
		if !live {
			if err := rawCallApiServer(client, "/livez"); err == nil {
				klog.Infof("API Server is Live")
				live = true
			}
		}

		if !ready {
			if err := rawCallApiServer(client, "/readyz"); err == nil {
				klog.Infof("API Server is Ready")
				ready = true
			}
		}

		return live && ready, nil
	}); err != nil {
		return err
	}

	return nil
}

func rawCallApiServer(client *restclient.RESTClient, endpoint string) error {
	stdout := strings.Builder{}
	stderr := strings.Builder{}
	ioStreams := genericclioptions.IOStreams{In: os.Stdin, Out: &stdout, ErrOut: &stderr}
	err := rawhttp.RawGet(client, ioStreams, endpoint)
	klog.V(2).Infof("Raw GET to %q: stdout:%q stderr:%q err:%v", endpoint, stdout.String(), stderr.String(), err)
	if err != nil {
		return err
	}
	return nil
}

func buildClient() (*restclient.RESTClient, error) {
	kubeconfig := filepath.Join(config.DataDir, "resources", string(config.KubeAdmin), "kubeconfig")
	kubeConfigFlags := genericclioptions.NewConfigFlags(true).WithDeprecatedPasswordFlag()
	kubeConfigFlags.KubeConfig = &kubeconfig
	matchVersionKubeConfigFlags := cmdutil.NewMatchVersionFlags(kubeConfigFlags)
	f := cmdutil.NewFactory(matchVersionKubeConfigFlags)
	restClient, err := f.RESTClient()
	if err != nil {
		return nil, err
	}
	return restClient, nil
}
