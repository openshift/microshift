#!/bin/bash

set -eux

# This file is not needed anymore in 4.7+, but when rolling back to 4.6
# the ovs pod needs it to know ovs is running on the host.
touch /var/run/ovs-config-executed
NM_CONN_PATH="/etc/NetworkManager/system-connections"
# this flag tracks if any config change was made
nm_config_changed=0
MANAGED_NM_CONN_SUFFIX="-slave-ovs-clone"

# when creating the bridge, we use a value lower than NM's ethernet device default route metric
# (we pick 48 and 49 to be lower than anything that NM chooses by default)
BRIDGE_METRIC="48"
BRIDGE1_METRIC="49"

# Configuration paths
MICROSHIFT_CONFIG_FILE_PATH="/etc/microshift/config.yaml"
ovnk_config_dir='/etc/ovnk'
ovnk_var_dir='/var/lib/ovnk'
extra_bridge_file="${ovnk_config_dir}/extra_bridge"
iface_default_hint_file="${ovnk_var_dir}/iface_default_hint"

# Parse gatewayInterface from MicroShift config file
# Returns the value of network.gatewayInterface or empty string
get_gateway_interface_config() {
  if [ ! -f "$MICROSHIFT_CONFIG_FILE_PATH" ]; then
    echo ""
    return
  fi
  # Parse YAML to get network.gatewayInterface value
  # This handles both "gatewayInterface: value" and "gatewayInterface: 'value'" formats
  local value=""
  value=$(awk '
    /^network:/ { in_network=1; next }
    in_network && /^[a-zA-Z]/ && !/^  / { in_network=0 }
    in_network && /gatewayInterface:/ {
      gsub(/.*gatewayInterface:[[:space:]]*/, "")
      gsub(/["\047]/, "")  # Remove quotes
      gsub(/[[:space:]]*$/, "")  # Trim trailing whitespace
      print
      exit
    }
  ' "$MICROSHIFT_CONFIG_FILE_PATH" 2>/dev/null || echo "")
  echo "$value"
}

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

###############################################################################
# Uplink mode functions - used when network.gatewayInterface is set
###############################################################################

# Used to clone a slave connection by uuid, returns new uuid
clone_slave_connection() {
  local uuid="$1"
  local old_name
  old_name="$(nmcli -g connection.id connection show uuid "$uuid")"
  local new_name="${old_name}${MANAGED_NM_CONN_SUFFIX}"
  if nmcli connection show id "${new_name}" &> /dev/null; then
    echo "WARN: existing ovs slave ${new_name} connection profile file found, overwriting..." >&2
    nmcli connection delete id "${new_name}" &> /dev/null
  fi
  nmcli connection clone $uuid "${new_name}" &> /dev/null
  nmcli -g connection.uuid connection show "${new_name}"
}

# Used to replace an old master connection uuid with a new one on all connections
replace_connection_master() {
  local old="$1"
  local new="$2"
  for conn_uuid in $(nmcli -g UUID connection show) ; do
    if [ "$(nmcli -g connection.master connection show uuid "$conn_uuid")" != "$old" ]; then
      continue
    fi
    local active_state=$(nmcli -g GENERAL.STATE connection show "$conn_uuid")
    local autoconnect=$(nmcli -g connection.autoconnect connection show "$conn_uuid")
    if [ "$active_state" != "activated" ] && [ "$autoconnect" != "yes" ]; then
      # Assume that slave profiles intended to be used are those that are:
      # - active
      # - or inactive (which might be due to link being down) but to be autoconnected.
      # Otherwise, ignore them.
      continue
    fi
    # make changes for slave profiles in a new clone
    local new_uuid
    new_uuid=$(clone_slave_connection $conn_uuid)
    nmcli conn mod uuid $new_uuid connection.master "$new"
    nmcli conn mod $new_uuid connection.autoconnect-priority 100
    nmcli conn mod $new_uuid connection.autoconnect no
    echo "Replaced master $old with $new for slave profile $new_uuid"
  done
}

# Add a deactivated connection profile
add_nm_conn() {
  nmcli c add "$@" connection.autoconnect no
}

# Activates an ordered set of NM connection profiles
activate_nm_connections() {
  local connections=("$@")

  # make sure to set bond or team slaves autoconnect, otherwise as we
  # activate one slave, the other slave might get implicitly re-activated
  # with the old profile, activating the old master, interfering and
  # causing the former activation to fail.
  for conn in "${connections[@]}"; do
    local slave_type=$(nmcli -g connection.slave-type connection show "$conn")
    if [ "$slave_type" = "team" ] || [ "$slave_type" = "bond" ]; then
      nmcli c mod "$conn" connection.autoconnect yes
    fi
  done
  # Then activate all the connections
  for conn in "${connections[@]}"; do
    local active_state=$(nmcli -g GENERAL.STATE conn show "$conn")
    if [ "$active_state" != "activated" ]; then
      for i in {1..10}; do
        echo "Attempt $i to bring up connection $conn"
        nmcli conn up "$conn" && s=0 && break || s=$?
        sleep 5
      done
      if [ $s -eq 0 ]; then
        echo "Brought up connection $conn successfully"
      else
        echo "ERROR: Cannot bring up connection $conn after $i attempts"
        return $s
      fi
    else
      echo "Connection $conn already activated"
    fi
    nmcli c mod "$conn" connection.autoconnect yes
  done
}

# Writes content of $iface into $iface_default_hint_file
write_iface_default_hint() {
  local iface_default_hint_file="$1"
  local iface="$2"
  echo "${iface}" >| "${iface_default_hint_file}"
}

# Returns the stored interface default hint if valid
get_iface_default_hint() {
  local iface_default_hint_file=$1
  if [ -f "${iface_default_hint_file}" ]; then
    local iface_default_hint=$(cat "${iface_default_hint_file}")
    if [ "${iface_default_hint}" != "" ] &&
       [ "${iface_default_hint}" != "br-ex" ] &&
       [ "${iface_default_hint}" != "br-ex1" ] &&
       [ -d "/sys/class/net/${iface_default_hint}" ]; then
       echo "${iface_default_hint}"
       return
    fi
  fi
  echo ""
}

# Finds the default interface for uplink mode
get_default_interface() {
  local iface=""
  local counter=0
  local iface_default_hint_file="$1"
  local extra_bridge_file="$2"
  local extra_bridge=""
  if [ -f "${extra_bridge_file}" ]; then
    extra_bridge=$(cat ${extra_bridge_file})
  fi
  # find default interface
  while [ ${counter} -lt 12 ]; do
    # check ipv4
    if [ "${extra_bridge}" != "" ]; then
      iface=$(ip route show default | grep -v "br-ex1" | grep -v "${extra_bridge}" | awk '{ if ($4 == "dev") { print $5; exit } }')
    else
      iface=$(ip route show default | grep -v "br-ex1" | awk '{ if ($4 == "dev") { print $5; exit } }')
    fi
    if [[ -n "${iface}" ]]; then
      break
    fi
    # check ipv6
    if [ "${extra_bridge}" != "" ]; then
      iface=$(ip -6 route show default | grep -v "br-ex1" | grep -v "${extra_bridge}" | awk '{ if ($4 == "dev") { print $5; exit } }')
    else
      iface=$(ip -6 route show default | grep -v "br-ex1" | awk '{ if ($4 == "dev") { print $5; exit } }')
    fi
    if [[ -n "${iface}" ]]; then
      break
    fi

    # Interface may not get default route immediately after boot
    if [ ${counter} -gt 8 ]; then
      if [ "${extra_bridge}" != "" ]; then
        iface=$(ip route show | grep -v "br-ex1" | grep -v "${extra_bridge}" | awk '{ if ($2 == "dev") { print $3; exit } }')
      else
        iface=$(ip route show | grep -v "br-ex1" | awk '{ if ($2 == "dev") { print $3; exit } }')
      fi
      if [[ -n "${iface}" ]]; then
        break
      fi
      if [ "${extra_bridge}" != "" ]; then
        iface=$(ip -6 route show | grep -v "br-ex1" | grep -v "${extra_bridge}" | awk '{ if ($2 == "dev") { print $3; exit } }')
      else
        iface=$(ip -6 route show | grep -v "br-ex1" | grep -w -v "lo" | awk '{ if ($2 == "dev") { print $3; exit } }')
      fi
      if [[ -n "${iface}" ]]; then
        break
      fi
    fi

    counter=$((counter+1))
    sleep 5
  done
  # if the default interface does not point out of br-ex or br-ex1
  if [ "${iface}" != "br-ex" ] && [ "${iface}" != "br-ex1" ]; then
    iface_default_hint=$(get_iface_default_hint "${iface_default_hint_file}")
    if [ "${iface_default_hint}" != "" ] &&
       [ "${iface_default_hint}" != "${iface}" ]; then
      while [ ${counter} -le 12 ]; do
        if [ "$(ip route show default dev "${iface_default_hint}")" != "" ]; then
          iface="${iface_default_hint}"
          break
        fi
        if [ "$(ip -6 route show default dev "${iface_default_hint}")" != "" ]; then
          iface="${iface_default_hint}"
          break
        fi
        if [ "$(ip route show dev "${iface_default_hint}")" != "" ]; then
          iface="${iface_default_hint}"
          break
        fi
        if [ "$(ip -6 route show dev "${iface_default_hint}")" != "" ]; then
          iface="${iface_default_hint}"
          break
        fi
        counter=$((counter+1))
        sleep 5
      done
    fi
    if [ "${iface}" != "" ]; then
      write_iface_default_hint "${iface_default_hint_file}" "${iface}"
    fi
  fi
  echo "${iface}"
}

# Configure driver options for vmxnet3
configure_driver_options() {
  intf=$1
  if [ ! -f "/sys/class/net/${intf}/device/uevent" ]; then
    echo "Device file doesn't exist, skipping setting multicast mode"
  else
    driver=$(cat "/sys/class/net/${intf}/device/uevent" | grep DRIVER | awk -F "=" '{print $2}')
    echo "Driver name is" $driver
    if [ "$driver" = "vmxnet3" ]; then
      ifconfig "$intf" allmulti
    fi
  fi
}

# Given an interface, generates NM configuration to add to an OVS bridge
convert_to_bridge() {
  local iface=${1}
  local bridge_name=${2}
  local port_name=${3}
  local bridge_metric=${4}
  local ovs_port="ovs-port-${bridge_name}"
  local ovs_interface="ovs-if-${bridge_name}"
  local default_port_name="ovs-port-${port_name}" # ovs-port-phys0
  local bridge_interface_name="ovs-if-${port_name}" # ovs-if-phys0
  if [ "$iface" = "$bridge_name" ]; then
    ifaces=$(ovs-vsctl list-ifaces ${iface})
    for intf in $ifaces; do configure_driver_options $intf; done
    echo "Networking already configured and up for ${bridge_name}!"
    return
  fi
  nm_config_changed=1
  if [ -z "$iface" ]; then
    echo "ERROR: Unable to find default gateway interface"
    exit 1
  fi
  # find the MAC from the default interface
  if ! iface_mac=$(<"/sys/class/net/${iface}/address"); then
    echo "Unable to determine default interface MAC"
    exit 1
  fi
  echo "MAC address found for iface: ${iface}: ${iface_mac}"
  # find MTU from original iface
  iface_mtu=$(ip link show "$iface" | awk '{print $5; exit}')
  if [[ -z "$iface_mtu" ]]; then
    echo "Unable to determine default interface MTU, defaulting to 1500"
    iface_mtu=1500
  else
    echo "MTU found for iface: ${iface}: ${iface_mtu}"
  fi
  # store old conn for later
  old_conn=$(nmcli --fields UUID,DEVICE conn show --active | awk "/\s${iface}\s*\$/ {print \$1}")
  # create bridge
  if ! nmcli connection show "$bridge_name" &> /dev/null; then
    ovs-vsctl --timeout=30 --if-exists del-br "$bridge_name"
    add_nm_conn type ovs-bridge con-name "$bridge_name" conn.interface "$bridge_name" 802-3-ethernet.mtu ${iface_mtu}
  fi
  # find default port to add to bridge
  if ! nmcli connection show "$default_port_name" &> /dev/null; then
    ovs-vsctl --timeout=30 --if-exists del-port "$bridge_name" ${iface}
    add_nm_conn type ovs-port conn.interface ${iface} master "$bridge_name" con-name "$default_port_name"
  fi
  if ! nmcli connection show "$ovs_port" &> /dev/null; then
    ovs-vsctl --timeout=30 --if-exists del-port "$bridge_name" "$bridge_name"
    add_nm_conn type ovs-port conn.interface "$bridge_name" master "$bridge_name" con-name "$ovs_port"
  fi
  extra_phys_args=()
  # check if this interface is a vlan, bond, team, or ethernet type
  if [ $(nmcli --get-values connection.type conn show ${old_conn}) == "vlan" ]; then
    iface_type=vlan
    vlan_id=$(nmcli --get-values vlan.id conn show ${old_conn})
    if [ -z "$vlan_id" ]; then
      echo "ERROR: unable to determine vlan_id for vlan connection: ${old_conn}"
      exit 1
    fi
    vlan_parent=$(nmcli --get-values vlan.parent conn show ${old_conn})
    if [ -z "$vlan_parent" ]; then
      echo "ERROR: unable to determine vlan_parent for vlan connection: ${old_conn}"
      exit 1
    fi
    extra_phys_args=( dev "${vlan_parent}" id "${vlan_id}" )
  elif [ $(nmcli --get-values connection.type conn show ${old_conn}) == "bond" ]; then
    iface_type=bond
    bond_opts=$(nmcli --get-values bond.options conn show ${old_conn})
    if [ -n "$bond_opts" ]; then
      extra_phys_args+=( bond.options "${bond_opts}" )
      MODE_REGEX="(^|,)mode=active-backup(,|$)"
      MAC_REGEX="(^|,)fail_over_mac=(1|active|2|follow)(,|$)"
      if [[ $bond_opts =~ $MODE_REGEX ]] && [[ $bond_opts =~ $MAC_REGEX ]]; then
        clone_mac=0
      fi
    fi
  elif [ $(nmcli --get-values connection.type conn show ${old_conn}) == "team" ]; then
    iface_type=team
    team_config_opts=$(nmcli --get-values team.config -e no conn show ${old_conn})
    if [ -n "$team_config_opts" ]; then
      extra_phys_args+=( team.config "${team_config_opts//[[:space:]]/}" )
      team_mode=$(echo "${team_config_opts}" | jq -r ".runner.name // empty")
      team_mac_policy=$(echo "${team_config_opts}" | jq -r ".runner.hwaddr_policy // empty")
      MAC_REGEX="(by_active|only_active)"
      if [ "$team_mode" = "activebackup" ] && [[ "$team_mac_policy" =~ $MAC_REGEX ]]; then
        clone_mac=0
      fi
    fi
  elif [ $(nmcli --get-values connection.type conn show ${old_conn}) == "tun" ]; then
    iface_type=tun
    tun_mode=$(nmcli --get-values tun.mode -e no connection show ${old_conn})
    extra_phys_args+=( tun.mode "${tun_mode}" )
  else
    iface_type=802-3-ethernet
  fi
  if [ ! "${clone_mac:-}" = "0" ]; then
    extra_phys_args+=( 802-3-ethernet.cloned-mac-address "${iface_mac}" )
  fi
  if ! nmcli connection show "$bridge_interface_name" &> /dev/null; then
    ovs-vsctl --timeout=30 --if-exists destroy interface ${iface}
    add_nm_conn type ${iface_type} conn.interface ${iface} master "$default_port_name" con-name "$bridge_interface_name" \
    connection.autoconnect-priority 100 802-3-ethernet.mtu ${iface_mtu} ${extra_phys_args[@]+"${extra_phys_args[@]}"}
  fi
  # Get the new connection uuids
  new_conn=$(nmcli -g connection.uuid conn show "$bridge_interface_name")
  ovs_port_conn=$(nmcli -g connection.uuid conn show "$ovs_port")
  # Update connections with master property set to use the new connection
  replace_connection_master $old_conn $new_conn
  replace_connection_master $iface $new_conn
  if ! nmcli connection show "$ovs_interface" &> /dev/null; then
    ovs-vsctl --timeout=30 --if-exists destroy interface "$bridge_name"
    if nmcli --fields ipv4.method,ipv6.method conn show $old_conn | grep manual; then
      echo "Static IP addressing detected on default gateway connection: ${old_conn}"
      nmcli conn clone "${old_conn}" "${ovs_interface}"
      shopt -s nullglob
      new_conn_files=(${NM_CONN_PATH}/"${ovs_interface}"*)
      shopt -u nullglob
      if [ ${#new_conn_files[@]} -ne 1 ] || [ ! -f "${new_conn_files[0]}" ]; then
        echo "ERROR: could not find ${ovs_interface} conn file after cloning from ${old_conn}"
        exit 1
      fi
      new_conn_file="${new_conn_files[0]}"
      sed -i '/^multi-connect=.*$/d' ${new_conn_file}
      sed -i '/^autoconnect=.*$/d' ${new_conn_file}
      sed -i '/^\[connection\]$/a autoconnect=false' ${new_conn_file}
      sed -i '/^\[connection\]$/,/^\[/ s/^type=.*$/type=ovs-interface/' ${new_conn_file}
      sed -i '/^\[connection\]$/a slave-type=ovs-port' ${new_conn_file}
      sed -i '/^\[connection\]$/a master='"$ovs_port_conn" ${new_conn_file}
      cat <<EOF >> ${new_conn_file}
[ovs-interface]
type=internal
EOF
      nmcli c load ${new_conn_file}
      nmcli c mod "${ovs_interface}" conn.interface "$bridge_name" \
        802-3-ethernet.mtu ${iface_mtu} 802-3-ethernet.cloned-mac-address ${iface_mac} \
        ipv4.route-metric "${bridge_metric}" ipv6.route-metric "${bridge_metric}"
      echo "Loaded new $ovs_interface connection file: ${new_conn_file}"
    else
      extra_if_brex_args=""
      num_ipv4_addrs=$(ip -j a show dev ${iface} | jq ".[0].addr_info | map(. | select(.family == \"inet\")) | length")
      if [ "$num_ipv4_addrs" -gt 0 ]; then
        extra_if_brex_args+="ipv4.may-fail no "
      fi
      num_ip6_addrs=$(ip -j a show dev ${iface} | jq ".[0].addr_info | map(. | select(.family == \"inet6\" and .scope != \"link\")) | length")
      if [ "$num_ip6_addrs" -gt 0 ]; then
        extra_if_brex_args+="ipv6.may-fail no "
      fi
      dhcp_client_id=$(nmcli --get-values ipv4.dhcp-client-id conn show ${old_conn})
      if [ -n "$dhcp_client_id" ]; then
        extra_if_brex_args+="ipv4.dhcp-client-id ${dhcp_client_id} "
      fi
      dhcp6_client_id=$(nmcli --get-values ipv6.dhcp-duid conn show ${old_conn})
      if [ -n "$dhcp6_client_id" ]; then
        extra_if_brex_args+="ipv6.dhcp-duid ${dhcp6_client_id} "
      fi
      ipv6_addr_gen_mode=$(nmcli --get-values ipv6.addr-gen-mode conn show ${old_conn})
      if [ -n "$ipv6_addr_gen_mode" ]; then
        extra_if_brex_args+="ipv6.addr-gen-mode ${ipv6_addr_gen_mode} "
      fi
      ipv4_ignore_auto_dns=$(nmcli --get-values ipv4.ignore-auto-dns conn show ${old_conn})
      if [ -n "$ipv4_ignore_auto_dns" ]; then
        extra_if_brex_args+="ipv4.ignore-auto-dns ${ipv4_ignore_auto_dns} "
      fi
      ipv6_ignore_auto_dns=$(nmcli --get-values ipv6.ignore-auto-dns conn show ${old_conn})
      if [ -n "$ipv6_ignore_auto_dns" ]; then
        extra_if_brex_args+="ipv6.ignore-auto-dns ${ipv6_ignore_auto_dns} "
      fi
      ipv4_dns=$(nmcli --get-values ipv4.dns conn show ${old_conn})
      if [ -n "$ipv4_dns" ]; then
        extra_if_brex_args+="ipv4.dns ${ipv4_dns} "
      fi
      ipv6_dns=$(nmcli --get-values ipv6.dns conn show ${old_conn})
      if [ -n "$ipv6_dns" ]; then
        extra_if_brex_args+="ipv6.dns ${ipv6_dns} "
      fi
      add_nm_conn type ovs-interface slave-type ovs-port conn.interface "$bridge_name" master "$ovs_port_conn" con-name \
        "$ovs_interface" 802-3-ethernet.mtu ${iface_mtu} 802-3-ethernet.cloned-mac-address ${iface_mac} \
        ipv4.route-metric "${bridge_metric}" ipv6.route-metric "${bridge_metric}" ${extra_if_brex_args}
    fi
  fi
  configure_driver_options "${iface}"
  update_nm_conn_files "$bridge_name" "$port_name"
}

# Used to print network state
print_state() {
  echo "Current device, connection, interface and routing state:"
  nmcli -g all device | grep -v unmanaged
  nmcli -g all connection
  ip -d address show
  ip route show
  ip -6 route show
}

# Setup an exit trap to rollback on error (used in uplink mode)
handle_exit() {
  e=$?
  [ $e -eq 0 ] && print_state && exit 0
  echo "ERROR: configure-ovs exited with error: $e"
  print_state
  dir=$(mktemp -d -t "configure-ovs-$(date +%Y-%m-%d-%H-%M-%S)-XXXXXXXXXX")
  update_nm_conn_files br-ex phys0
  copy_nm_conn_files "$dir"
  update_nm_conn_files br-ex1 phys1
  copy_nm_conn_files "$dir"
  echo "Copied OVS configuration to $dir for troubleshooting"
  echo "Attempting to restore previous configuration..."
  rollback_nm
  print_state
  exit $e
}

###############################################################################
# Main script logic
###############################################################################

# Read gateway interface configuration
gateway_interface=$(get_gateway_interface_config)
echo "Gateway interface configuration: '${gateway_interface}'"

# Check for and remove any previous NM configuration (upgrade scenario)
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
  # Remove NM config files created by previous release when doing upgrade
  rollback_nm
fi

if [ -z "$gateway_interface" ]; then
  ###########################################################################
  # No-uplink mode (default): Create br-ex with static MAC and no uplink
  ###########################################################################
  echo "Using no-uplink mode: creating br-ex with static MAC"

  # use a locally administered address as the static MAC of br-ex
  static_mac="0a:59:00:00:00:01"
  ovs-vsctl --timeout=15 --may-exist add-br br-ex -- br-set-external-id br-ex bridge-id br-ex -- set bridge br-ex fail-mode=standalone other_config:hwaddr="${static_mac}"
  nmcli device set br-ex managed no

else
  ###########################################################################
  # Uplink mode: Convert physical interface to OVS bridge
  ###########################################################################
  echo "Using uplink mode with gateway interface: ${gateway_interface}"

  # Set up exit trap for rollback on error
  trap "handle_exit" EXIT

  # Create config directories
  mkdir -p "${ovnk_config_dir}"
  mkdir -p "${ovnk_var_dir}"

  # Handle gateway interface setting
  if [ "$gateway_interface" != "auto" ]; then
    # User specified a specific interface
    write_iface_default_hint "${iface_default_hint_file}" "${gateway_interface}"
  fi

  # For upgrade scenarios, stabilize existing configuration
  iface_default_hint=$(get_iface_default_hint "${iface_default_hint_file}")
  if [ "${iface_default_hint}" == "" ]; then
    current_interface=$(get_bridge_physical_interface ovs-if-phys0)
    if [ "${current_interface}" != "" ]; then
      write_iface_default_hint "${iface_default_hint_file}" "${current_interface}"
    fi
  fi

  # Handle conflicting hint files
  if [ -f "${iface_default_hint_file}" ] &&
     [ -f "${extra_bridge_file}" ] &&
     [ "$(cat "${iface_default_hint_file}")" == "$(cat "${extra_bridge_file}")" ]; then
    echo "${iface_default_hint_file} and ${extra_bridge_file} share the same content"
    echo "Deleting file ${iface_default_hint_file} to choose a different interface"
    rm -f "${iface_default_hint_file}"
    rm -f /run/configure-ovs-boot-done
  fi

  # On every boot, rollback and regenerate configuration
  if [ ! -f /run/configure-ovs-boot-done ]; then
    echo "Running on boot, restoring previous configuration before proceeding..."
    rollback_nm
    print_state
  fi
  touch /run/configure-ovs-boot-done

  # Determine the interface to use
  iface=$(get_default_interface "${iface_default_hint_file}" "$extra_bridge_file")

  if [ "$iface" != "br-ex" ]; then
    if [ -f "$extra_bridge_file" ] || [ -z "$(nmcli connection show --active br-ex 2> /dev/null)" ]; then
      echo "Bridge br-ex is not active, restoring previous configuration before proceeding..."
      rollback_nm
      print_state
    fi
  fi

  # Convert the interface to an OVS bridge
  convert_to_bridge "$iface" "br-ex" "phys0" "${BRIDGE_METRIC}"

  # Remove second bridge if it exists
  if nmcli connection show br-ex1 &> /dev/null || nmcli connection show ovs-if-phys1 &> /dev/null; then
    update_nm_conn_files br-ex1 phys1
    rm_nm_conn_files
  fi

  if [ -f "$extra_bridge_file" ]; then
    rm -f "$extra_bridge_file"
  fi

  # Remove bridges created by openshift-sdn
  ovs-vsctl --timeout=30 --if-exists del-br br0

  # Make sure everything is activated
  connections=()
  while IFS= read -r connection; do
    if [[ $connection == *"$MANAGED_NM_CONN_SUFFIX" ]]; then
      connections+=("$connection")
    fi
  done < <(nmcli -g NAME c)
  connections+=(ovs-if-phys0 ovs-if-br-ex)
  activate_nm_connections "${connections[@]}"
fi
