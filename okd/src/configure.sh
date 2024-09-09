#!/bin/bash

    cat > /etc/microshift/config.yaml <<EOF
storage:
    driver: "none"
EOF

systemctl enable microshift