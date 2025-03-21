#!/usr/bin/env bash
set -xeuo pipefail

OUTPUT=$1
PAYLOAD=/tmp/bee.jpeg

# Download payload
curl -o "${PAYLOAD}" https://raw.githubusercontent.com/openvinotoolkit/model_server/main/demos/common/static/images/bee.jpeg

# Add an inference header (len=63)
echo -n '{"inputs" : [{"name": "0", "shape": [1], "datatype": "BYTES"}]}' > "${OUTPUT}"

# Add size of the data (image) in binary format (4 bytes, little endian)
printf "%08X" "$(stat --format=%s "${PAYLOAD}")" | sed 's/\(..\)/\1\n/g' | tac | tr -d '\n' | xxd -r -p >> "${OUTPUT}"

# Add the data, i.e. the image
cat "${PAYLOAD}" >> "${OUTPUT}"
