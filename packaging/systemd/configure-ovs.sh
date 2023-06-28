#!/bin/bash

set -eux

# This file is not needed anymore in 4.7+, but when rolling back to 4.6
# the ovs pod needs it to know ovs is running on the host.
touch /var/run/ovs-config-executed
NM_CONN_PATH="/etc/NetworkManager/system-connections"
# this flag tracks if any config change was made
nm_config_changed=0
MANAGED_NM_CONN_SUFFIX="-slave-ovs-clone"
# Workaround to ensure OVS is installed due to bug in systemd Requires:
# https://bugzilla.redhat.com/show_bug.cgi?id=1888017
copy_nm_conn_files() {
  local dst_path="$1"
  for src in "${MANAGED_NM_CONN_FILES[@]}"; do
    src_path=$(dirname "$src")
    file=$(basename "$src")
    if [ -f "$src_path/$file" ]; then
      if [ ! -f "$dst_path/$file" ]; then
        echo "Copying configuration $file"
        cp "$src_path/$file" "$dst_path/$file"
      elif ! cmp --silent "$src_path/$file" "$dst_path/$file"; then
        echo "Copying updated configuration $file"
        cp -f "$src_path/$file" "$dst_path/$file"
      else
        echo "Skipping $file since it's equal at destination"
      fi
    else
      echo "Skipping $file since it does not exist at source"
    fi
  done
}
update_nm_conn_files() {
  bridge_name=${1}
  port_name=${2}
  ovs_port="ovs-port-${bridge_name}"
  ovs_interface="ovs-if-${bridge_name}"
  default_port_name="ovs-port-${port_name}" # ovs-port-phys0
  bridge_interface_name="ovs-if-${port_name}" # ovs-if-phys0
  # In RHEL7 files in /{etc,run}/NetworkManager/system-connections end without the suffix '.nmconnection', whereas in RHCOS they end with the suffix.
  MANAGED_NM_CONN_FILES=($(echo "${NM_CONN_PATH}"/{"$bridge_name","$ovs_interface","$ovs_port","$bridge_interface_name","$default_port_name"}{,.nmconnection}))
  shopt -s nullglob
  MANAGED_NM_CONN_FILES+=(${NM_CONN_PATH}/*${MANAGED_NM_CONN_SUFFIX}.nmconnection ${NM_CONN_PATH}/*${MANAGED_NM_CONN_SUFFIX})
  shopt -u nullglob
}
# Used to remove files managed by configure-ovs
rm_nm_conn_files() {
  for file in "${MANAGED_NM_CONN_FILES[@]}"; do
    if [ -f "$file" ]; then
      rm -f "$file"
      echo "Removed nmconnection file $file"
      nm_config_changed=1
    fi
  done
}
# Used to remove a bridge
remove_ovn_bridges() {
  bridge_name=${1}
  port_name=${2}
  # Reload configuration, after reload the preferred connection profile
  # should be auto-activated
  update_nm_conn_files ${bridge_name} ${port_name}
  rm_nm_conn_files
  nmcli connection reload
  # NetworkManager will not remove ${bridge_name} if it has the patch port created by ovn-kubernetes
  # so remove explicitly
  ovs-vsctl --timeout=30 --if-exists del-br ${bridge_name}
}
# Removes any previous ovs configuration
remove_all_ovn_bridges() {
  echo "Reverting any previous OVS configuration"
  
  remove_ovn_bridges br-ex phys0
  if [ -d "/sys/class/net/br-ex1" ]; then
    remove_ovn_bridges br-ex1 phys1
  fi
  
  echo "OVS configuration successfully reverted"
}
# Reloads NM NetworkManager profiles if any configuration change was done.
# Accepts a list of devices that should be re-connect after reload.
reload_profiles_nm() {
  if [ $nm_config_changed -eq 0 ]; then
    # no config was changed, no need to reload
    return
  fi
  # reload profiles
  nmcli connection reload
  # precautionary sleep of 10s (default timeout of NM to bring down devices)
  sleep 10

  # After reload, devices that were already connected should connect again
  # if any profile is available. If no profile is available, a device can
  # remain disconnected and we have to explicitly connect it so that a
  # profile is generated. This can happen for physical devices but should
  # not happen for software devices as those always require a profile.
  for dev in $@; do
    # Only attempt to connect a disconnected device
    local connected_state=$(nmcli -g GENERAL.STATE device show "$dev" || echo "")
    if [[ "$connected_state" =~ "disconnected" ]]; then
      # keep track if a profile by the same name as the device existed 
      # before we attempt activation
      local named_profile_existed=$([ -f "${NM_CONN_PATH}/${dev}" ] || [ -f "${NM_CONN_PATH}/${dev}.nmconnection" ] && echo "yes")
      
      for i in {1..10}; do
          echo "Attempt $i to connect device $dev"
          nmcli device connect "$dev" && break
          sleep 5
      done
      # if a profile did not exist before but does now, it was generated
      # but we want it to be ephemeral, so move it back to /run
      if [ ! "$named_profile_existed" = "yes" ]; then
        local dst="/run/NetworkManager/system-connections/"
        MANAGED_NM_CONN_FILES=("${NM_CONN_PATH}/${dev}" "${NM_CONN_PATH}/${dev}.nmconnection")
        copy_nm_conn_files "${dst}"
        rm_nm_conn_files
        # reload profiles so that NM notices that some might have been moved
        nmcli connection reload
      fi
    fi
    echo "Waiting for interface $dev to activate..."
    if ! timeout 60 bash -c "while ! nmcli -g DEVICE,STATE c | grep "'"'"$dev":activated'"'"; do sleep 5; done"; then
      echo "Warning: $dev did not activate"
    fi
  done
  nm_config_changed=0
}
# Removes all configuration and reloads NM if necessary
rollback_nm() {
  phys0=$(get_bridge_physical_interface ovs-if-phys0)
  phys1=$(get_bridge_physical_interface ovs-if-phys1)
  
  # Revert changes made by /usr/local/bin/configure-ovs.sh during SDN migration.
  remove_all_ovn_bridges
  
  # reload profiles so that NM notices that some were removed
  reload_profiles_nm "$phys0" "$phys1"
}
# Accepts parameters $bridge_interface (e.g. ovs-port-phys0)
# Returns the physical interface name if $bridge_interface exists, "" otherwise
get_bridge_physical_interface() {
  local bridge_interface="$1"
  local physical_interface=""
  physical_interface=$(nmcli -g connection.interface-name conn show "${bridge_interface}" 2>/dev/null || echo "")
  echo "${physical_interface}"
}

nm_need_rollback=false
update_nm_conn_files br-ex phys0
if [ -d "/sys/class/net/br-ex1" ]; then
  update_nm_conn_files br-ex1 phys1
fi
for file in "${MANAGED_NM_CONN_FILES[@]}"; do
  if [ -f "$file" ]; then
    nm_need_rollback=true
  fi
done

if [ "${nm_need_rollback}" = "true" ]; then
  # Remove NM config files create by previous release when doing upgrade
  rollback_nm
fi

# use a locally administered addresse as the static MAC of br-ex
static_mac="0a:59:00:00:00:01"
ovs-vsctl --timeout=15 --may-exist add-br br-ex -- br-set-external-id br-ex bridge-id br-ex  -- set bridge br-ex fail-mode=standalone other_config:hwaddr="${static_mac}"
nmcli device set br-ex managed no
