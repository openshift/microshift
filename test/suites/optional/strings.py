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
