#! /usr/bin/env bash

source "$(dirname "${BASH_SOURCE}")/lib/init.sh"

# Generates the .syso file used to add compile-time VERSIONINFO metadata to the
# Windows binary.
function os::build::generate_windows_versioninfo() {
  set -x  
  if [[ "${SOURCE_GIT_TAG}" =~ ^[a-z-]*([0-9]+)\.([0-9]+)\.([0-9]+).* ]] ; then
    local major=${BASH_REMATCH[1]}
    local minor=${BASH_REMATCH[2]}
    local patch=${BASH_REMATCH[3]}
  fi
  local windows_versioninfo_file=`mktemp --suffix=".versioninfo.json"`
  cat <<EOF >"${windows_versioninfo_file}"
{
       "FixedFileInfo":
       {
               "FileVersion": {
                       "Major": ${major},
                       "Minor": ${minor},
                       "Patch": ${patch}
               },
               "ProductVersion": {
                       "Major": ${major},
                       "Minor": ${minor},
                       "Patch": ${patch}
               },
               "FileFlagsMask": "3f",
               "FileFlags ": "00",
               "FileOS": "040004",
               "FileType": "01",
               "FileSubType": "00"
       },
       "StringFileInfo":
       {
               "Comments": "",
               "CompanyName": "Red Hat, Inc.",
               "InternalName": "openshift client",
               "FileVersion": "${SOURCE_GIT_TAG}",
               "InternalName": "oc",
               "LegalCopyright": "© Red Hat, Inc. Licensed under the Apache License, Version 2.0",
               "LegalTrademarks": "",
               "OriginalFilename": "oc.exe",
               "PrivateBuild": "",
               "ProductName": "OpenShift Client",
               "ProductVersion": "${SOURCE_GIT_TAG}",
               "SpecialBuild": ""
       },
       "VarFileInfo":
       {
               "Translation": {
                       "LangID": "0409",
                       "CharsetID": "04B0"
               }
       }
}
EOF
  goversioninfo -o ${OS_ROOT}/cmd/oc/oc.syso ${windows_versioninfo_file}
}
readonly -f os::build::generate_windows_versioninfo

os::build::generate_windows_versioninfo
