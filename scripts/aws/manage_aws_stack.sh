#!/bin/bash
#
# Script for management of an AWS stack.

set -euo pipefail

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

action_create() {
    local -r stack_name="${1}"
    local -r inst_type="${2}"
    local -r region="${3}"
    local -r pub_key="${4}"
    local -r os="${5}"
    local ami="${6}"

    if [ -z "${stack_name}" ] ; then
        echo "ERROR: stack name not set" 1>&2
        exit 1
    fi
    if [ -z "${inst_type}" ] ; then
        echo "ERROR: instance type not set" 1>&2
        exit 1
    fi
    if [ -z "${region}" ] ; then
        echo "ERROR: region type not set" 1>&2
        exit 1
    fi

    local -r cf_tpl_file="${SCRIPTDIR}/cf-gen.yaml"
    
    local arch="x86_64"
    if [[ "${inst_type%.*}" =~ .+"g".* ]]; then
        arch="arm64"
    fi
    
    local ec2_type="VirtualMachine"
    if [[ "${inst_type}" =~ c[0-9]+[gn].metal ]]; then
        ec2_type="MetalMachine"
    fi

    if [[ "${ami}" == "" ]]; then
        ami=$(get_amis "${os}" "${arch}" "${region}" | head -n 1 | awk '{print $2}')
    fi
    
    echo "Stack name: ${stack_name}"
    echo "Instance type: ${inst_type}" 
    echo "OS: ${os}" 
    echo "Arch: ${arch}" 
    echo "AMI ID: ${ami}" 
    echo "Region: ${region}"

    aws --region "${region}" cloudformation create-stack --stack-name "${stack_name}" \
        --template-body "file://${cf_tpl_file}" \
        --capabilities CAPABILITY_NAMED_IAM \
        --parameters \
            ParameterKey=HostInstanceType,ParameterValue="${inst_type}" \
            ParameterKey=Machinename,ParameterValue="${stack_name}" \
            ParameterKey=AmiId,ParameterValue="${ami}" \
            ParameterKey=EC2Type,ParameterValue="${ec2_type}" \
            ParameterKey=PublicKeyString,ParameterValue="$(cat "${pub_key}")" \
            ParameterKey=StackLaunchTemplate,ParameterValue="${stack_name}-launch-template"
    
    echo "Waiting for stack to be created"
    aws --region "${region}" cloudformation wait stack-create-complete --stack-name "${stack_name}"
    echo "Stack created successfully"

    local instance_id
    # shellcheck disable=SC2016
    instance_id="$(aws --region "${region}" cloudformation describe-stacks --stack-name "${stack_name}" \
        --query 'Stacks[].Outputs[?OutputKey == `InstanceId`].OutputValue' --output text)"
    echo "Waiting for stack status to be OK"
    aws --region "${region}" ec2 wait instance-status-ok --instance-id "${instance_id}"
    echo "Stack status OK"

    local public_ip
    # shellcheck disable=SC2016
    public_ip=$(aws --region "${region}" cloudformation describe-stacks --stack-name "${stack_name}" \
        --query 'Stacks[].Outputs[?OutputKey == `PublicIp`].OutputValue' --output text)

    echo "PUBLIC IP: ${public_ip}"
}

action_delete() {
    local -r stack_name="${1}"
    local -r region="${2}"

    aws --region "${region}" cloudformation delete-stack --stack-name "${stack_name}"
    echo "Waiting for stack to be deleted"
    aws --region "${region}" cloudformation wait stack-delete-complete --stack-name "${stack_name}"
    echo "Stack deleted"
}

action_describe() {
    local -r stack_name="${1}"
    local -r region="${2}"

    aws --region "${region}" cloudformation describe-stack-events --stack-name "${stack_name}" --output json
}

action_logs() {
    local -r stack_name="${1}"
    local -r region="${2}"
    local instance_id
    # shellcheck disable=SC2016
    instance_id="$(aws --region "${region}" cloudformation describe-stacks --stack-name "${stack_name}" \
        --query 'Stacks[].Outputs[?OutputKey == `InstanceId`].OutputValue' --output text)"

    aws ec2 get-console-output --instance-id "${instance_id}" --output text
}

action_ami() {
    local -r os="${1}"
    local -r arch="${2}"
    local -r region="${3}"

    echo "Listing AMIs for:"
    echo "OS: ${os}"
    echo "Arch: ${arch:-"any"}"
    echo "Region: ${region}"
    echo "Sorting latest to oldest"
    get_amis "${os}" "${arch}" "${region}"
}

get_amis() {
    local -r os="${1:-""}"
    local -r arch="${2:-""}"
    local -r region="${3}"
    local ami_table
    ami_table=$(aws ec2 describe-images --region "${region}" \
        --filters "Name=name,Values=$(tr '[:lower:]' '[:upper:]' <<< "${os}")*HVM*Hourly2-GP3" "Name=architecture,Values=${arch}*" \
        --query 'Images[*].[Name,ImageId,Architecture]' --output text)
  
    echo "${ami_table}" | sort -t'-' -k3,3r
}

usage() {
    cat - <<EOF
Script for AWS stack management

Usage:      

    Create

        ${BASH_SOURCE[0]} create \ 
            --stack-name <name> \ 
            --region <region> \ 
            --inst-type <type> \ 
            [--pub-key <path>] \ 
            [--os <os>] \ 
            [--ami <ami>]

    Delete|Describe|Logs

        ${BASH_SOURCE[0]} (delete|describe|logs) \ 
            --stack-name <name> \ 
            --region <region>

    List AMIs

        ${BASH_SOURCE[0]} ami \ 
            --region <region> \ 
            [--os <os>] \ 
            [--arch <arch>]

Arguments:
    --stack-name <name>:    The name of the stack.

    --region <region>:      The region (e.g. eu-west-1).

    --inst-type <type>:     (create only) The type of instance to create 
                            (e.g. t3.large, c5n.metal).

    [--pub-key <path>]:     (create only) The path to a .pub file. Defaults 
                            to ${HOME}/.ssh/id_rsa.pub.

    [--ami <ami>]:          (create only) AMI code to use. If not specified, 
                            the latest build will be selected.

    [--os <os>]:            (create/ami only) specific version of RHEL, 
                            e.g. 'rhel-9.3'. Defaults to rhel-9.4.

    [--arch <arch>]:        (ami only) Filter AMIs based on arch. Must be 
                            one of: x86_64, arm64. Lists all when not set.

EOF
}

if [ $# == 0 ] ; then
    usage
    exit 1
fi

action="$1"
shift

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
    echo "ERROR: region not set" 1>&2
    exit 1
fi

case "${action}" in
    create)
        "action_${action}" "${stack_name}" "${inst_type}" "${region}" "${pub_key}" "${os}" "${ami}"
        ;;

    delete|describe|logs)
        if [ -z "${stack_name}" ] ; then
            echo "ERROR: stack name not set" 1>&2
            exit 1
        fi
        "action_${action}" "${stack_name}" "${region}"
        ;;

    ami)
        "action_${action}" "${os}" "${arch}" "${region}"
        ;;

    *)
        usage
        exit 1
        ;;
esac
