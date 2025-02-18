FROM registry.stage.redhat.io/rhel9-eus/rhel-9.6-bootc:9.6

# Build a container image to be used by Bootc Image Builder in the subsequent
# stages for producing ISO images.
#
# This is a workaround until https://issues.redhat.com/browse/RHEL-34807 is fixed.
