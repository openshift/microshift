#!/usr/bin/env bash

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

OUTDIR="${1:-${HOME}/MicroShiftReviews}"

if [ -z "${OUTDIR}" ]; then
    echo "Specify an output directory as the first argument to the script." 1>&2
    exit 1
fi

if [ -d "${OUTDIR}" ]; then
    echo "${OUTDIR} already exists" 1>&2
    exit 1
fi

if [ -z "${GITHUB_TOKEN}" ]; then
    echo "Set GITHUB_TOKEN variable" 1>&2
    exit 1
fi

echo "Creating ${OUTDIR} ..."
mkdir -p "${OUTDIR}"
chmod 0700 "${OUTDIR}"

echo "Creating ${OUTDIR}/venv ..."
python3 -m venv "${OUTDIR}/venv"

echo "Installing dinghy ..."
"${OUTDIR}/venv/bin/python3" -m pip install dinghy

echo "Configuring dinghy ..."
cp "${SCRIPTDIR}/dinghy.yaml" "${OUTDIR}/"
cat - >"${OUTDIR}/update.sh" <<EOF
#!/usr/bin/env bash

export GITHUB_TOKEN="${GITHUB_TOKEN}"

cd "${OUTDIR}"

./venv/bin/python3 -m dinghy ./dinghy.yaml
EOF
chmod 0700 "${OUTDIR}/update.sh"
