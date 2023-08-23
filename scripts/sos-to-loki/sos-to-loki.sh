#!/bin/bash
set -euo pipefail

TMP_ROOT_DIR=/tmp/sos-to-loki

get_mountpoint() {
	local input="${1}"

	# Already existing, extracted sos report directory
	if [ -d "${input}" ]; then
		echo "${input}"
		return
	fi

	mkdir -p "${TMP_ROOT_DIR}"

	# Remote url to sos-report archive: download & extract
	if [[ $1 = http* ]]; then
		(
			filename="$(basename "${input}")"
			without_ext="${filename%%.*}"
			out="${TMP_ROOT_DIR}/${without_ext}"
			if [[ -d "${out}" ]]; then
				echo "${out}"
				return
			fi

			cd "${TMP_ROOT_DIR}"
			curl --silent --remote-name "${input}"

			tar xf "${filename}"
			echo "${out}"
		)
		return
	fi

	# Local path to sos-report archive: copy & extract
	(
		filename="$(basename "${input}")"
		without_ext="${filename%%.*}"
		out="${TMP_ROOT_DIR}/${without_ext}"
		if [[ -d "${out}" ]]; then
			echo "${out}"
			return
		fi

		cd "${TMP_ROOT_DIR}"
		cp "${input}" .

		tar xf "${filename}"
		echo "${out}"
	)
}

if [ "$#" -ne 1 ]; then
	echo "./sos-to-loki.sh <path to sos-report>"
	exit 1
fi

echo "Fetching SOS report"
sos="$(get_mountpoint "${1}")"
echo "Path to local SOS report: ${sos}"

# TODO: Rename localhost

# $ grep "Hostname set to" ./journalctl_--no-pager
# Aug 18 10:25:04 el92-src-fdo-host1 systemd[1]: Hostname set to <el92-src-fdo-host1>.

# localhost kernel -> el92-src-fdo-host1 kernel

echo "Starting Grafana+Loki+Promtail stack"
podman pod exists sos-to-loki && podman kube down ./pod.yaml >/dev/null
SOS_REPORT_INPUT="${sos}" envsubst <./pod.yaml | podman kube play - >/dev/null

sos_journal="${sos}/sos_commands/logs/journalctl_--no-pager"

timezone=$(grep 'Local time' "${sos}/date" | tr -s ' ' | rev | cut -d ' ' -f1 | rev)
local_timezone=$(date +%Z)

first_entry=$(head -n1 "${sos_journal}" | cut -d ' ' -f1-3)
last_entry=$(tail -n1 "${sos_journal}" | cut -d ' ' -f1-3)

local_first_entry=$(date --date "${first_entry} ${timezone}" +"%b %d %H:%M:%S")
local_last_entry=$(date --date "${last_entry} ${timezone}" +"%b %d %H:%M:%S")

from=$(date --date "${local_first_entry} ${local_timezone} -5 minutes" +%s%N | cut -b1-13)
to=$(date --date "${local_last_entry} ${local_timezone} +5 minutes" +%s%N | cut -b1-13)

echo ""
printf "                 Timezone     First journal entry       Last entry\n"
printf "SOS              %-12s %-25s %-20s\n" "${timezone}" "${first_entry}" "${last_entry}"
printf "Local            %-12s %-25s %-20s\n" "${local_timezone}" "${local_first_entry}" "${local_last_entry}"
printf "URL timestamps   %-12s %-25s %-20s\n" "${local_timezone}" "${from}" "${to}"

echo -e "\nLink to query with journal:\nhttp://localhost:3000/explore?orgId=1&left=%7B%22datasource%22:%22P8E80F9AEF21F6940%22,%22queries%22:%5B%7B%22refId%22:%22A%22,%22expr%22:%22%7Bfilename%3D%5C%22%2Flogs%2Fsos_commands%2Flogs%2Fjournalctl_--no-pager%5C%22%7D%20%7C%3D%20%60%60%22,%22queryType%22:%22range%22,%22datasource%22:%7B%22type%22:%22loki%22,%22uid%22:%22P8E80F9AEF21F6940%22%7D,%22editorMode%22:%22builder%22%7D%5D,%22range%22:%7B%22from%22:%22${from}%22,%22to%22:%22${to}%22%7D%7D"
