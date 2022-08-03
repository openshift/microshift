#!/bin/sh
set -eux

OVS_HANDLER_THREADS=${OVS_HANDLER_THREADS:-1}
OVS_REVALIDATOR_THREADS=${OVS_REVALIDATOR_THREADS:-1}

# limit the number of threads in ovs-vswitchd to limit the memory consumption
if [[ "${OVS_HANDLER_THREADS}" != "auto" ]]; then
    # this is not effective as of OpenvSwitch 2.17, it will depend on the CPUAffinity that we set.
    # TODO: talk to the ovs team to see if we can obey this config again, and avoid
    # setting up CPUAffinity (then we don't need separate service definitions)
    ovs-vsctl set open . other_config:n-handler-threads=${OVS_HANDLER_THREADS}
fi

if [[ "${OVS_REVALIDATOR_THREADS}" != "auto" ]]; then
    ovs-vsctl set open . other_config:n-revalidator-threads=${OVS_REVALIDATOR_THREADS}
fi
