#!/bin/bash
#
# Script for management of an AWS stack.

set -euo pipefail

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

action_create() {
  if [ -z "${inst_type}" ] ; then
    echo "ERROR: Instance type was not set" 1>&2
    exit 1
  fi

  local -r cf_tpl_file="${SCRIPTDIR}/cf-gen.yaml"

  EC2_AMI="${ami}"
  MICROSHIFT_OS="${os:-"rhel-9.4"}"

  ARCH="x86_64"
  if [[ "${inst_type%.*}" =~ .*"g".* ]]; then
    ARCH="arm64"
  fi

  if [[ "${EC2_AMI}" == "" ]]; then
    EC2_AMI=$(get_amis "${MICROSHIFT_OS}" "${ARCH}" | head -n 1 | awk '{print $2}')
  fi

  ec2Type="VirtualMachine"
  if [[ "${inst_type}" =~ c[0-9]+[gn].metal ]]; then
    ec2Type="MetalMachine"
  fi

  ami_id=${EC2_AMI}

  echo "Stack name: ${stack_name}"
  echo "Instance type: ${inst_type}" 
  echo "OS: ${MICROSHIFT_OS}" 
  echo "ARCH: ${ARCH}" 
  echo "AMI ID: ${EC2_AMI}" 
  echo "region: ${region}" 

  aws --region "${region}" cloudformation create-stack --stack-name "${stack_name}" \
    --template-body "file://${cf_tpl_file}" \
    --capabilities CAPABILITY_NAMED_IAM \
    --parameters \
    ParameterKey=HostInstanceType,ParameterValue="${inst_type}" \
    ParameterKey=Machinename,ParameterValue="${stack_name}" \
    ParameterKey=AmiId,ParameterValue="${ami_id}" \
    ParameterKey=EC2Type,ParameterValue="${ec2Type}" \
    ParameterKey=PublicKeyString,ParameterValue="$(cat "${pub_key}")" \
    ParameterKey=StackLaunchTemplate,ParameterValue="${stack_name}-launch-template"
    
  echo "Waiting for stack to be created"
  aws --region "${region}" cloudformation wait stack-create-complete --stack-name "${stack_name}"
  echo "Stack created successfully"

  # shellcheck disable=SC2016
  instance_id="$(aws --region "${region}" cloudformation describe-stacks --stack-name "${stack_name}" \
    --query 'Stacks[].Outputs[?OutputKey == `InstanceId`].OutputValue' --output text)"

  echo "Waiting for stack status to be OK"
  aws --region "${region}" ec2 wait instance-status-ok --instance-id "${instance_id}"
  echo "Stack status OK"


  # shellcheck disable=SC2016
  public_ip=$(aws --region "${region}" cloudformation describe-stacks --stack-name "${stack_name}" \
    --query 'Stacks[].Outputs[?OutputKey == `PublicIp`].OutputValue' --output text)
  
  echo "PUBLIC IP: ${public_ip}"
}

action_delete() {
  aws --region "${region}" cloudformation delete-stack --stack-name "${stack_name}"
  echo "Waiting for stack to be deleted"
  aws --region "${region}" cloudformation wait stack-delete-complete --stack-name "${stack_name}"
  echo "Stack deleted"
  
  exit 0
}

action_describe() {
  aws --region "${region}" cloudformation describe-stack-events --stack-name "${stack_name}" --output json
  exit 0
}

action_logs() {
  # shellcheck disable=SC2016
  instance_id="$(aws --region "${region}" cloudformation describe-stacks --stack-name "${stack_name}" \
    --query 'Stacks[].Outputs[?OutputKey == `InstanceId`].OutputValue' --output text)"
  
  aws ec2 get-console-output --instance-id "${instance_id}" --output text
}

action_ami() {
  echo "Listing AMIs for:"
  echo "OS: ${os}"
  echo "ARCH: ${arch:-"any"}"
  echo "REGION: ${region}"
  echo "Sorting latest to oldest"
  get_amis "${os}" "${arch}"
}

get_amis() {
  local -r os_variant="${1:-""}"
  local -r architecture="${2:-""}"
  local ami_table
  ami_table=$(aws ec2 describe-images --region "${region}" \
    --filters "Name=name,Values=${os_variant^^}*HVM-*Hourly2-GP3" "Name=architecture,Values=${architecture}*" \
    --query 'Images[*].[Name,ImageId,Architecture]' --output text)
  
  echo "${ami_table}" | sort -t'-' -k3,3r 

}

usage() {
  cat - <<EOF
Script for AWS stack management

Usage:      

  Create

    ${BASH_SOURCE[0]} create --stack-name <name> --inst-type <type> --region <region> [--pub-key <path>] [--os <os>]

  Delete|Describe|Logs

    ${BASH_SOURCE[0]} (delete|describe|logs) --stack-name <name> --region <region>

  List AMIs

    ${BASH_SOURCE[0]} ami --region <region> [--os <os>] [--arch <arch>]

Arguments:
  --stack-name <name>:  The name of the stack.
  --inst-type <type>:   (create only) The type of instance to create (e.g. t3.large, c5n.metal).
  --region <region>:    (create only) The region where the stack should be created, e.g. eu-west-1.
  [--pub-key <path>]:   (create only) The path to a .pub file. Defaults to ${HOME}/.ssh/id_rsa.pub.
  [--os <os>]:          (create/ami only) specific version of RHEL, e.g. 'rhel-9.3'. Defaults to rhel-9.4.
  [--arch <arch>]:      (ami only) Filter for listing only x86_64/arm64 AMIs. Lists all when not set.

EOF
}

action="$1"
shift

case "${action}" in
  create|delete|describe|logs|ami)
    ;;
  *)
    usage
    exit 1
    ;;
esac

stack_name=""
inst_type=""
region=""
pub_key="${HOME}/.ssh/id_rsa.pub"
ami=""
os="rhel-9.4"
arch=""

while [ $# -gt 0 ]; do
  case "$1" in
    --inst-type|--region|--stack-name|--pub-key|--ami|--os|--arch)
      var="${1/--/}"
      var="${var/-/_}"
      if [ -n "$2" ] && [ "${2:0:1}" != "-" ]; then 
          declare "${var}=$2"
          shift 2
      else
          usage
          exit 1
      fi
      ;;
    *)
      
      usage
      exit 1
      ;;
  esac
done

if [ -z "${region}" ] ; then
  usage
  exit 1
fi

if [ "${action}" == "ami" ] ; then
  action_ami "${os}" "${arch}"
  exit 0
fi

if [ -z "${stack_name}" ] || [ -z "${region}" ] ; then
  usage
  exit 1
fi

"action_${action}"

