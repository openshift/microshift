# nvidia-gpu-setup

This Ansible role configures NVIDIA GPU support for MicroShift on Red Hat Enterprise Linux, enabling GPU-accelerated workloads.

## Features

- Installs NVIDIA GPU drivers
- Configures NVIDIA Container Toolkit for CRI-O
- Deploys NVIDIA Device Plugin for Kubernetes
- Validates GPU availability and functionality
- Provides test workloads for verification

## Requirements

- RHEL 9.x with MicroShift installed
- NVIDIA GPU hardware present on the system
- Internet connectivity for downloading packages and container images
- SELinux in enforcing mode (fully supported)

## Role Variables

### Default Variables (defaults/main.yml)

| Variable | Default | Description |
|----------|---------|-------------|
| `nvidia_driver_version` | `575-open` | NVIDIA driver version (supports open-source kernel modules with "-open" suffix, use "latest" for auto-detection) |
| `rhel_version` | `ansible_distribution_major_version` | RHEL major version |
| `validate_gpu` | `true` | Run validation tests after installation |
| `device_plugin_manifest_dir` | `/etc/microshift/manifests.d/nvidia-gpu` | Directory for MicroShift manifests |
| `cuda_repo_url` | Auto-generated | CUDA repository URL for your RHEL version and architecture |
| `container_toolkit_repo_url` | NVIDIA stable repo | Container toolkit repository URL |

### Configurable Parameters

- **Driver Version**: Set `nvidia_driver_version` to install a specific driver version
- **Repository URLs**: Can be overridden for air-gapped environments
- **Validation**: Set `validate_gpu: false` to skip validation steps

## Dependencies

This role requires MicroShift to be installed and running. Use in conjunction with:
- `install-microshift` role
- `setup-microshift-host` role
- `configure-firewall` role

## Installation Steps

The role performs the following operations:

1. **GPU Detection**: Verifies NVIDIA GPU hardware presence
2. **Driver Installation**:
   - Adds CUDA repository
   - Installs NVIDIA driver module and fabric manager
   - Blacklists nouveau driver
   - Rebuilds initramfs
3. **Container Runtime Configuration**:
   - Installs NVIDIA Container Toolkit
   - Creates custom SELinux policy module for NVIDIA containers
   - Runs nvidia-ctk to configure CRI-O runtime
   - Configures CRI-O for NVIDIA runtime
   - Sets SELinux booleans for device access
4. **Device Plugin Deployment**:
   - Downloads NVIDIA Device Plugin manifest
   - Creates kustomization.yaml for namespace configuration
   - Deploys to MicroShift manifests directory
5. **System Reboot**: Reboots if driver changes require it
6. **Validation**:
   - Verifies nvidia-smi functionality
   - Checks Device Plugin pod status
   - Confirms GPU resources in node capacity

## Example Playbook

### Basic Usage

```yaml
- hosts: microshift
  become: yes
  roles:
    - role: nvidia-gpu-setup
```

### With Custom Driver Version

```yaml
- hosts: microshift
  become: yes
  vars:
    nvidia_driver_version: "535"
    validate_gpu: true
  roles:
    - role: nvidia-gpu-setup
```

### With Auto-detected Latest Driver

```yaml
- hosts: microshift
  become: yes
  vars:
    nvidia_driver_version: "latest"
    validate_gpu: true
  roles:
    - role: nvidia-gpu-setup
```

### Integration with setup-node.yml

This role is integrated into the main `setup-node.yml` playbook and is executed conditionally based on the `enable_gpu` variable. The complete setup is handled automatically when you run:

```bash
ansible-playbook -i inventory/inventory setup-node.yml \
  -e enable_gpu_arg=true
```

This will execute all necessary roles in the correct order, including GPU setup after MicroShift is running.

## Usage

### Running the Playbook

```bash
# Using the main setup playbook with GPU support
ansible-playbook -i inventory/inventory setup-node.yml \
  -e enable_gpu_arg=true \
  -e deploy_gpu_test_arg=true

# With custom driver version
ansible-playbook -i inventory/inventory setup-node.yml \
  -e enable_gpu_arg=true \
  -e nvidia_driver_version=535

# With auto-detected latest driver version
ansible-playbook -i inventory/inventory setup-node.yml \
  -e enable_gpu_arg=true \
  -e nvidia_driver_version=latest

# GPU support without running test workload
ansible-playbook -i inventory/inventory setup-node.yml \
  -e enable_gpu_arg=true \
  -e deploy_gpu_test_arg=false
```

### Verification

After successful installation, verify GPU support:

```bash
# Check GPU availability
nvidia-smi

# Check Device Plugin pods
export KUBECONFIG=/var/lib/microshift/resources/kubeadmin/kubeconfig
oc get pods -n nvidia-device-plugin | grep nvidia

# Check node GPU resources
oc get nodes -o json | jq '.items[0].status.capacity."nvidia.com/gpu"'

# The nvidia-gpu-test role will automatically deploy and validate test workloads
# if deploy_gpu_test_arg=true (default)
```

## Test Workloads

Test workloads are handled by the companion `nvidia-gpu-test` role, which is automatically executed when `deploy_gpu_test_arg=true` (default). The test role deploys:

- **nvidia-smi**: Simple GPU verification using CUDA base image
- **cuda-vector-add**: CUDA computation test using the vectoradd sample

Test workloads are deployed to the `gpu-test` namespace and logs are saved to `./gpu-test-logs/` directory.

To manually check test results:

```bash
export KUBECONFIG=/var/lib/microshift/resources/kubeadmin/kubeconfig
oc get pods -n gpu-test
oc logs -n gpu-test nvidia-smi-all-gpus
oc logs -n gpu-test cuda-vector-add
```

## Troubleshooting

### Common Issues

1. **No GPU Detected**: Ensure GPU is properly installed and visible with `lspci -nnv | grep -i nvidia`

2. **Driver Installation Fails**: 
   - Check RHEL version compatibility
   - Verify repository accessibility
   - Review `/var/log/nvidia-installer.log`

3. **Device Plugin Not Running**:
   - Check MicroShift status: `systemctl status microshift`
   - Review pod logs: `oc logs -n nvidia-device-plugin <device-plugin-pod>`

4. **GPU Not Available to Pods**:
   - Verify CRI-O configuration: `cat /etc/crio/crio.conf.d/99-nvidia-runtime.conf`
   - Check SELinux: `getsebool container_use_devices`

### Manual Cleanup

If needed to reset GPU configuration:

```bash
# Remove NVIDIA packages
sudo dnf remove nvidia-* libnvidia-*

# Remove configuration files
sudo rm -f /etc/modprobe.d/nouveau-blacklist.conf
sudo rm -f /etc/crio/crio.conf.d/99-nvidia-runtime.conf
sudo rm -rf /etc/microshift/manifests.d/nvidia-gpu/

# Delete NVIDIA device plugin namespace and all its resources
kubectl delete namespace nvidia-device-plugin

# Rebuild initramfs
sudo dracut --force

# Reboot
sudo reboot
```

## References

- [NVIDIA GPU with Red Hat Device Edge](https://docs.nvidia.com/datacenter/cloud-native/edge/latest/nvidia-gpu-with-device-edge.html)
- [NVIDIA Container Toolkit](https://docs.nvidia.com/datacenter/cloud-native/container-toolkit/overview.html)
- [NVIDIA Device Plugin for Kubernetes](https://github.com/NVIDIA/k8s-device-plugin)
- [MicroShift Documentation](https://access.redhat.com/documentation/en-us/red_hat_build_of_microshift/)