GDP_CONFIG_DROPIN = '''
genericDevicePlugin:
  status: Enabled
  devices:
     - name: fakeserial
       groups:
         - paths:
             - path: /dev/ttyPipeB0
'''

GDP_CONFIG_DROPIN_WITH_MOUNT = '''
genericDevicePlugin:
  status: Enabled
  devices:
     - name: fakeserial
       groups:
         - paths:
             - path: /dev/ttyPipeB0
               mountPath: /dev/myrenamedserial
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

GDP_CONFIG_FUSE_COUNT = '''
genericDevicePlugin:
  status: Enabled
  devices:
     - name: fuse
       groups:
         - count: 10
           paths:
             - path: /dev/fuse
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


def get_ttyusb_pod_definition(name: str, num_devices: int):
    return f"""
apiVersion: v1
kind: Pod
metadata:
  name: {name}
spec:
  containers:
  - name: serialdevice-app-container
    image: registry.access.redhat.com/ubi9/ubi:9.6
    command: ["sleep", "infinity"]
    resources:
      limits:
        device.microshift.io/fakeserial: "{num_devices}"
    securityContext:
      allowPrivilegeEscalation: false
      capabilities:
        drop: ["ALL"]
      runAsNonRoot: true
      seccompProfile:
        type: "RuntimeDefault"
"""


def append_to_preamble(content: str) -> str:
    # Add 4 spaces before each line
    content = "    " + content.replace("\n", "\n    ")
    return CONFIGMAP_PREAMBLE + content
