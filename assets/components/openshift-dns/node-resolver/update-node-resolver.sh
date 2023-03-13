#!/bin/bash
set -uo pipefail

trap 'jobs -p | xargs kill || true; wait; exit 0' TERM

OPENSHIFT_MARKER="openshift-generated-node-resolver"
HOSTS_FILE="/etc/hosts"
TEMP_FILE="/etc/hosts.tmp"

IFS=', ' read -r -a services <<< "${SERVICES}"

# Make a temporary file with the old hosts file's attributes.
if ! cp -f --attributes-only "${HOSTS_FILE}" "${TEMP_FILE}"; then
  echo "Failed to preserve hosts file. Exiting."
  exit 1
fi

while true; do
  declare -A svc_ips
  for svc in "${services[@]}"; do
    # Fetch service IP from cluster dns if present. We make several tries
    # to do it: IPv4, IPv6, IPv4 over TCP and IPv6 over TCP. The two last ones
    # are for deployments with Kuryr on older OpenStack (OSP13) - those do not
    # support UDP loadbalancers and require reaching DNS through TCP.
    cmds=('dig -t A @"${NAMESERVER}" +short "${svc}.${CLUSTER_DOMAIN}"|grep -v "^;"'
          'dig -t AAAA @"${NAMESERVER}" +short "${svc}.${CLUSTER_DOMAIN}"|grep -v "^;"'
          'dig -t A +tcp +retry=0 @"${NAMESERVER}" +short "${svc}.${CLUSTER_DOMAIN}"|grep -v "^;"'
          'dig -t AAAA +tcp +retry=0 @"${NAMESERVER}" +short "${svc}.${CLUSTER_DOMAIN}"|grep -v "^;"')
    for i in ${!cmds[*]}
    do
      ips=($(eval "${cmds[i]}"))
      if [[ "$?" -eq 0 && "${#ips[@]}" -ne 0 ]]; then
        svc_ips["${svc}"]="${ips[@]}"
        break
      fi
    done
  done

  # Update /etc/hosts only if we get valid service IPs
  # We will not update /etc/hosts when there is coredns service outage or api unavailability
  # Stale entries could exist in /etc/hosts if the service is deleted
  if [[ -n "${svc_ips[*]-}" ]]; then
    # Build a new hosts file from /etc/hosts with our custom entries filtered out
    if ! sed --silent "/# ${OPENSHIFT_MARKER}/d; w ${TEMP_FILE}" "${HOSTS_FILE}"; then
      # Only continue rebuilding the hosts entries if its original content is preserved
      sleep 60 & wait
      continue
    fi

    # Append resolver entries for services
    rc=0
    for svc in "${!svc_ips[@]}"; do
      for ip in ${svc_ips[${svc}]}; do
        echo "${ip} ${svc} ${svc}.${CLUSTER_DOMAIN} # ${OPENSHIFT_MARKER}" >> "${TEMP_FILE}" || rc=$?
      done
    done
    if [[ $rc -ne 0 ]]; then
      sleep 60 & wait
      continue
    fi


    # TODO: Update /etc/hosts atomically to avoid any inconsistent behavior
    # Replace /etc/hosts with our modified version if needed
    cmp "${TEMP_FILE}" "${HOSTS_FILE}" || cp -f "${TEMP_FILE}" "${HOSTS_FILE}"
    # TEMP_FILE is not removed to avoid file create/delete and attributes copy churn
  fi
  sleep 60 & wait
  unset svc_ips
done
