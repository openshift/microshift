#! /usr/bin/env bash

# Usage:
# ./otelcol-jaeger-microshift.sh [run, cleanup]

set -eo pipefail

cleanup() {
  echo "DATA LOSS WARNING: Do you wish to stop and cleanup ALL MicroShift data AND cri-o container workloads?"
  select yn in "Yes" "No"; do
    case "${yn}" in
        Yes ) break ;;
        * ) echo "aborting cleanup; " ; exit;;
    esac
  done
  rm -rf ~/.kube/config openshift-client-linux.tar.gz
  sudo podman stop jaeger || true
  sudo podman stop microshift-bin || true
  sudo systemctl stop otelcol
  sudo dnf rm -y otelcol

  # crictl redirect STDOUT.  When no objects (pod, image, container) are present, crictl dump the help menu instead.  This may be confusing to users.
  sudo bash -c '
    echo "Stopping microshift"
    set +e
    systemctl stop --now microshift 2>/dev/null
    systemctl disable microshift 2>/dev/null

    echo "Removing crio pods"
    until crictl rmp --all --force 1>/dev/null; do sleep 1; done

    echo "Removing crio containers"
    crictl rm --all --force 1>/dev/null

    echo "Removing crio images"
    crictl rmi --all --prune 1>/dev/null

    echo "Killing conmon, pause processes"
    pkill -9 conmon
    pkill -9 pause

    echo "Removing MicroShift package"
    dnf rm -y microshift

    echo "Removing MicroShift local storage storage"
    rm -rf /var/lib/microshift

    echo "Cleanup succeeded"
  '
}

trap cleanup SIGINT
if [ "$1" = "run" ]
then
  # get and start opentelemetry collector service
  wget https://github.com/open-telemetry/opentelemetry-collector-releases/releases/download/v0.44.0/otelcol_0.44.0_linux_amd64.rpm
  sudo rpm -ivh otelcol_0.44.0_linux_amd64.rpm

  # start microshift otel-bin container to extract artifacts from
  sudo podman run --rm -d --name microshift-bin quay.io/sallyom/microshift:otel-bin sleep 100

  # extract opentelemetry-collector config  
  sudo podman cp microshift-bin:/opt/otelcol-config.yaml /etc/otelcol/config.yaml
  # uncomment the following if running from microshift checkout, otherwise follow the README to cp otelcol-service/config.yaml from podman image
  # sudo cp sample-opentelemetry/otelcol-service/config.yaml /etc/otelcol/config.yaml
  sudo systemctl daemon-reload && sudo systemctl restart otelcol
  sudo rm otelcol_0.44.0_linux_amd64.rpm

  # start jaeger
  sudo podman run -d --privileged --security-opt label=disable --rm --network=host --name jaeger  jaegertracing/all-in-one:1.34 --collector.grpc-server.max-message-size=9999999

  # start microshift from latest rpm, then replace binary with custom binary from quay.io/sallyom/microshift:otel-bin
  # this is from sallyom/microshift branch otel-tracing

  sudo dnf copr enable -y @redhat-et/microshift
  sudo dnf install -y microshift

  sudo podman cp microshift-bin:/usr/bin/microshift /usr/bin/
  sudo podman stop microshift-bin

  sudo firewall-cmd --zone=trusted --add-source=10.42.0.0/16 --permanent
  sudo firewall-cmd --zone=public --add-port=80/tcp --permanent
  sudo firewall-cmd --zone=public --add-port=443/tcp --permanent
  sudo firewall-cmd --zone=public --add-port=5353/udp --permanent
  sudo firewall-cmd --reload

  sudo systemctl start microshift
  echo "pause 5s for microshift data"
  sleep 5

  # install openshift & k8s clients
  curl -O https://mirror.openshift.com/pub/openshift-v4/$(uname -m)/clients/ocp/stable/openshift-client-linux.tar.gz
  sudo tar -xf openshift-client-linux.tar.gz -C /usr/local/bin oc kubectl
  rm openshift-client-linux.tar.gz

  # cp KUBECONFIG from microshift root storage
  mkdir -p ~/.kube
  sudo cp /var/lib/microshift/resources/kubeadmin/kubeconfig ~/.kube/config
  sudo chown `whoami`: ~/.kube/config
else
  if [ "$1" = "cleanup" ]
  then
    cleanup
  fi
fi
