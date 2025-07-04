GDP_CONFIG_DROPIN = '''
genericDevicePlugin:
  status: Enabled
  devices:
     - name: fakeserial
       groups:
         - paths:
             - path: /dev/ttyPipeB0
'''

CONFIGMAP_PREAMBLE = '''
apiVersion: v1
kind: ConfigMap
metadata:
  name: gdp-script
data:
  entrypoint.sh: |
    #!/usr/bin/env bash
    pip install pyserial
    /script/fake-serial-communication.py pod

  fake-serial-communication.py: |
'''


def append_to_preamble(content: str) -> str:
    # Add 4 spaces before each line
    content = "    " + content.replace("\n", "\n    ")
    return CONFIGMAP_PREAMBLE + content
