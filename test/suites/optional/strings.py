GDP_CONFIG_DROPIN = '''
genericDevicePlugin:
  status: Enabled
  devices:
     - name: fakeserial
       groups:
         - paths:
             - path: /dev/ttyPipeB0
'''

GDP_CONFIG_SERIAL_GLOB = '''
genericDevicePlugin:
  status: Enabled
  devices:
     - name: fakeserial
       groups:
         - paths:
             - path: /dev/ttyUSB*
'''

POD_SERIAL_2_DEVICES = '''
apiVersion: v1
kind: Pod
metadata:
  name: serial-test-pod
spec:
  containers:
  - name: serialdevice-app-container
    image: registry.access.redhat.com/ubi9/ubi:9.6
    command: ["sleep", "infinity"]
    resources:
      limits:
        device.microshift.io/fakeserial: "2"
    securityContext:
      allowPrivilegeEscalation: false
      capabilities:
        drop: ["ALL"]
      runAsNonRoot: true
      seccompProfile:
        type: "RuntimeDefault"
'''

POD_SERIAL_1_DEVICE = '''
apiVersion: v1
kind: Pod
metadata:
  name: serial-test-pod1
spec:
  containers:
  - name: serialdevice-app-container
    image: registry.access.redhat.com/ubi9/ubi:9.6
    command: ["sleep", "infinity"]
    resources:
      limits:
        device.microshift.io/fakeserial: "1"
    securityContext:
      allowPrivilegeEscalation: false
      capabilities:
        drop: ["ALL"]
      runAsNonRoot: true
      seccompProfile:
        type: "RuntimeDefault"
'''

POD_SERIAL_2_DEVICES_SECOND = '''
apiVersion: v1
kind: Pod
metadata:
  name: serial-test-pod2
spec:
  containers:
  - name: serialdevice-app-container
    image: registry.access.redhat.com/ubi9/ubi:9.6
    command: ["sleep", "infinity"]
    resources:
      limits:
        device.microshift.io/fakeserial: "2"
    securityContext:
      allowPrivilegeEscalation: false
      capabilities:
        drop: ["ALL"]
      runAsNonRoot: true
      seccompProfile:
        type: "RuntimeDefault"
'''

CONFIGMAP_PREAMBLE = '''
apiVersion: v1
kind: ConfigMap
metadata:
  name: gdp-script
data:
  entrypoint.sh: |
    #!/usr/bin/env bash
    # Fake user homedir for installing pip and pyserial.
    HOME=/tmp python3 -m ensurepip --upgrade
    HOME=/tmp python3 -m pip install pyserial
    HOME=/tmp /script/fake-serial-communication.py pod

  fake-serial-communication.py: |
'''


def append_to_preamble(content: str) -> str:
    # Add 4 spaces before each line
    content = "    " + content.replace("\n", "\n    ")
    return CONFIGMAP_PREAMBLE + content
