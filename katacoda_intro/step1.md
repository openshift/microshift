# Install MicroShift

As provided in the documentation at the repository, we can install MicroShift by running:

`curl -sfL https://raw.githubusercontent.com/redhat-et/microshift/main/install.sh | bash`{{execute}}

We will see some commands output like package installation and configuration
# Examining the environment

Once MicroShift has been installed we can run some example commands like:

`kubectl get all -A --context microshift`{{execute}}
