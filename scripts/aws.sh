#!/bin/bash

set -uo pipefail

export AWS_PAGER=""

err_log="err.log"

trap 'rm -f "$err_log"' EXIT

declare -A ami_map=(
  [us-west-2,x86_64,rhel-9.2]=ami-0378fd0689802d015    # RHEL-9.2.0_HVM-20240229-x86_64-33-Hourly2-GP3
  [us-west-2,x86_64,rhel-9.3]=ami-0c2f1f1137a85327e    # RHEL-9.3.0_HVM-20240229-x86_64-27-Hourly2-GP3
  [us-west-2,x86_64,rhel-9.4]=ami-0423fca164888b941    # RHEL-9.4.0_HVM-20240605-x86_64-82-Hourly2-GP3
  [us-west-2,arm64,rhel-9.2]=ami-0cb125bb261a63f52     # RHEL-9.2.0_HVM-20240229-arm64-33-Hourly2-GP3
  [us-west-2,arm64,rhel-9.3]=ami-04379fa947a959c92     # RHEL-9.3.0_HVM-20240229-arm64-27-Hourly2-GP3
  [us-west-2,arm64,rhel-9.4]=ami-05b40ce1c0e236ef2     # RHEL-9.4.0_HVM-20240605-arm64-82-Hourly2-GP3
  [eu-west-1,x86_64,rhel-9.2]=ami-02c220fcee5dab581    # RHEL-9.2.0_HVM-20240229-x86_64-33-Hourly2-GP3
  [eu-west-1,x86_64,rhel-9.3]=ami-05463a02d11667441    # RHEL-9.3.0_HVM-20240229-x86_64-27-Hourly2-GP3
  [eu-west-1,x86_64,rhel-9.4]=ami-07d4917b6f95f5c2a    # RHEL-9.4.0_HVM-20240605-x86_64-82-Hourly2-GP3
  [eu-west-1,arm64,rhel-9.2]=ami-06f9fb7baa169fdfc     # RHEL-9.2.0_HVM-20240229-arm64-33-Hourly2-GP3
  [eu-west-1,arm64,rhel-9.3]=ami-016e6894567bb7f3e     # RHEL-9.3.0_HVM-20240229-arm64-27-Hourly2-GP3
  [eu-west-1,arm64,rhel-9.4]=ami-02b8573b23fde21aa     # RHEL-9.4.0_HVM-20240605-arm64-82-Hourly2-GP3
)

spin() {
  pid="${1}"
  message_wait="${2}"
  message_ok="${3}"

  spin[0]="-"
  spin[1]="\\"
  spin[2]="|"
  spin[3]="/"

  while kill -0 "${pid}" 2>/dev/null
  do
    for i in "${spin[@]}"
    do
      echo -ne "\r${i} ${message_wait}"
      sleep 0.1
    done
  done

  wait "${pid}"
  exit_status=$?

  if [ "${exit_status}" -ne 0 ] ; then
    cat "${err_log}"
    exit 1
  else
    length=${#message_wait}
    length=$((length + 2))
    printf "\r%${length}s" ""
    echo -e "\r${message_ok}"
  fi

}

action_create() {
  if [ -z "${inst_type}" ] ; then
    echo "ERROR: inst_type not set" 1>&2
    exit 1
  fi

  cf_tpl_file=cf-gen.yaml

  EC2_AMI=""
  MICROSHIFT_OS=rhel-9.4

  ARCH="x86_64"
  if [[ "${inst_type%.*}" =~ .*"g".* ]]; then
    ARCH="arm64"
  fi

  if [[ "${EC2_AMI}" == "" ]]; then
    EC2_AMI="${ami_map["${region},${ARCH},${MICROSHIFT_OS}"]}"
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

  cat >"${cf_tpl_file}" <<EOF
AWSTemplateFormatVersion: 2010-09-09
Description: Template for RHEL machine Launch
Conditions:
  AddSecondaryVolume: !Not [!Equals [!Ref EC2Type, 'MetalMachine']]
Mappings:
 VolumeSize:
   MetalMachine:
     PrimaryVolumeSize: "300"
     SecondaryVolumeSize: "0"
     Throughput: 500
     Iops: 6000
   VirtualMachine:
     PrimaryVolumeSize: "200"
     SecondaryVolumeSize: "10"
     Throughput: 125
     Iops: 3000
Parameters:
  EC2Type:
    Default: 'VirtualMachine'
    Type: String
  VpcCidr:
    AllowedPattern: ^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])(\/(1[6-9]|2[0-4]))$
    ConstraintDescription: CIDR block parameter must be in the form x.x.x.x/16-24.
    Default: 10.192.0.0/16
    Description: CIDR block for VPC.
    Type: String
  PublicSubnetCidr:
    Description: Please enter the IP range (CIDR notation) for the public subnet in the first Availability Zone
    Type: String
    Default: 10.192.10.0/24
  AmiId:
    Description: Current RHEL AMI to use.
    Type: AWS::EC2::Image::Id
  Machinename:
    AllowedPattern: ^([a-zA-Z][a-zA-Z0-9\-]{0,26})$
    MaxLength: 27
    MinLength: 1
    ConstraintDescription: Machinename
    Description: Machinename
    Type: String
    Default: rhel-testbed-ec2-instance
  HostInstanceType:
    Default: t2.medium
    Type: String
  PublicKeyString:
    Type: String
    Description: The public key used to connect to the EC2 instance

Metadata:
  AWS::CloudFormation::Interface:
    ParameterGroups:
    - Label:
        default: "Host Information"
      Parameters:
      - HostInstanceType
    - Label:
        default: "Network Configuration"
      Parameters:
      - PublicSubnet
    ParameterLabels:
      PublicSubnet:
        default: "Worker Subnet"
      HostInstanceType:
        default: "Worker Instance Type"

Resources:
## VPC Creation

  RHELVPC:
    Type: AWS::EC2::VPC
    Properties:
      CidrBlock: !Ref VpcCidr
      EnableDnsHostnames: true
      EnableDnsSupport: true
      Tags:
        - Key: Name
          Value: RHELVPC

## Setup internet access

  RHELInternetGateway:
    Type: AWS::EC2::InternetGateway
    Properties:
      Tags:
        - Key: Name
          Value: RHELInternetGateway

  RHELGatewayAttachment:
    Type: AWS::EC2::VPCGatewayAttachment
    Properties:
      VpcId: !Ref RHELVPC
      InternetGatewayId: !Ref RHELInternetGateway

  RHELPublicSubnet:
    Type: AWS::EC2::Subnet
    Properties:
      VpcId: !Ref RHELVPC
      CidrBlock: !Ref PublicSubnetCidr
      MapPublicIpOnLaunch: true
      Tags:
        - Key: Name
          Value: RHELPublicSubnet

  RHELNatGatewayEIP:
    Type: AWS::EC2::EIP
    DependsOn: RHELGatewayAttachment
    Properties:
      Domain: vpc
  RHELNatGateway:
    Type: AWS::EC2::NatGateway
    Properties:
      AllocationId: !GetAtt RHELNatGatewayEIP.AllocationId
      SubnetId: !Ref RHELPublicSubnet

  RHELRouteTable:
    Type: AWS::EC2::RouteTable
    Properties:
      VpcId: !Ref RHELVPC
      Tags:
        - Key: Name
          Value: RHELRouteTable

  RHELPublicRoute:
    Type: AWS::EC2::Route
    DependsOn: RHELGatewayAttachment
    Properties:
      RouteTableId: !Ref RHELRouteTable
      DestinationCidrBlock: "0.0.0.0/0"
      GatewayId: !Ref RHELInternetGateway

  RHELPublicSubnetRouteTableAssociation:
    Type: AWS::EC2::SubnetRouteTableAssociation
    Properties:
      RouteTableId: !Ref RHELRouteTable
      SubnetId: !Ref RHELPublicSubnet

## Setup EC2 Roles and security

  RHELIamRole:
    Type: AWS::IAM::Role
    Properties:
      AssumeRolePolicyDocument:
        Version: "2012-10-17"
        Statement:
        - Effect: "Allow"
          Principal:
            Service:
            - "ec2.amazonaws.com"
          Action:
          - "sts:AssumeRole"
      Path: "/"

  RHELInstanceProfile:
    Type: "AWS::IAM::InstanceProfile"
    Properties:
      Path: "/"
      Roles:
      - Ref: "RHELIamRole"

  RHELSecurityGroup:
    Type: AWS::EC2::SecurityGroup
    Properties:
      GroupDescription: RHEL Host Security Group
      SecurityGroupIngress:
      - IpProtocol: icmp
        FromPort: -1
        ToPort: -1
        CidrIp: 0.0.0.0/0
      - IpProtocol: tcp
        FromPort: 22
        ToPort: 22
        CidrIp: 0.0.0.0/0
      - IpProtocol: tcp
        FromPort: 80
        ToPort: 80
        CidrIp: 0.0.0.0/0
      - IpProtocol: tcp
        FromPort: 443
        ToPort: 443
        CidrIp: 0.0.0.0/0
      - IpProtocol: tcp
        FromPort: 5353
        ToPort: 5353
        CidrIp: 0.0.0.0/0
      - IpProtocol: tcp
        FromPort: 5678
        ToPort: 5678
        CidrIp: 0.0.0.0/0
      - IpProtocol: tcp
        FromPort: 6443
        ToPort: 6443
        CidrIp: 0.0.0.0/0
      - IpProtocol: tcp
        FromPort: 30000
        ToPort: 32767
        CidrIp: 0.0.0.0/0
      - IpProtocol: udp
        FromPort: 30000
        ToPort: 32767
        CidrIp: 0.0.0.0/0
      VpcId: !Ref RHELVPC

  rhelLaunchTemplate:
    Type: AWS::EC2::LaunchTemplate
    Properties:
      LaunchTemplateName: ${stack_name}-launch-template
      LaunchTemplateData:
        BlockDeviceMappings:
        - DeviceName: /dev/sda1
          Ebs:
            VolumeSize: !FindInMap [VolumeSize, !Ref EC2Type, PrimaryVolumeSize]
            VolumeType: gp3
            Throughput: !FindInMap [VolumeSize, !Ref EC2Type, Throughput]
            Iops: !FindInMap [VolumeSize, !Ref EC2Type, Iops]
        - !If
          - AddSecondaryVolume
          - DeviceName: /dev/sdc
            Ebs:
              VolumeSize: !FindInMap [VolumeSize, !Ref EC2Type, SecondaryVolumeSize]
              VolumeType: gp3
          - !Ref AWS::NoValue

  RHELInstance:
    Type: AWS::EC2::Instance
    Properties:
      ImageId: !Ref AmiId
      LaunchTemplate:
        LaunchTemplateName: ${stack_name}-launch-template
        Version: !GetAtt rhelLaunchTemplate.LatestVersionNumber
      IamInstanceProfile: !Ref RHELInstanceProfile
      InstanceType: !Ref HostInstanceType
      NetworkInterfaces:
      - AssociatePublicIpAddress: "True"
        DeviceIndex: "0"
        GroupSet:
        - !GetAtt RHELSecurityGroup.GroupId
        SubnetId: !Ref RHELPublicSubnet
      Tags:
      - Key: Name
        Value: !Join ["", [!Ref Machinename]]
      PrivateDnsNameOptions:
        EnableResourceNameDnsARecord: true
        HostnameType: resource-name
      UserData:
        Fn::Base64: !Sub |
          #!/bin/bash -xe
          echo "====== Authorizing public key ======" | tee -a /tmp/init_output.txt
          echo "\${PublicKeyString}" >> /home/ec2-user/.ssh/authorized_keys
          # Use the same defaults as OCP to avoid failing requests to apiserver, such as
          # requesting logs.
          echo "====== Updating inotify =====" | tee -a /tmp/init_output.txt
          echo "fs.inotify.max_user_watches = 65536" >> /etc/sysctl.conf
          echo "fs.inotify.max_user_instances = 8192" >> /etc/sysctl.conf
          sysctl --system |& tee -a /tmp/init_output.txt
          sysctl -a |& tee -a /tmp/init_output.txt

Outputs:
  InstanceId:
    Description: RHEL Host Instance ID
    Value: !Ref RHELInstance
  PrivateIp:
    Description: The bastion host Private DNS, will be used for cluster install pulling release image
    Value: !GetAtt RHELInstance.PrivateIp
  PublicIp:
    Description: The bastion host Public IP, will be used for registering minIO server DNS
    Value: !GetAtt RHELInstance.PublicIp
EOF

  aws --region "${region}" cloudformation create-stack --stack-name "${stack_name}" \
    --template-body "file://${cf_tpl_file}" \
    --capabilities CAPABILITY_NAMED_IAM \
    --parameters \
    ParameterKey=HostInstanceType,ParameterValue="${inst_type}" \
    ParameterKey=Machinename,ParameterValue="${stack_name}" \
    ParameterKey=AmiId,ParameterValue="${ami_id}" \
    ParameterKey=EC2Type,ParameterValue="${ec2Type}" \
    ParameterKey=PublicKeyString,ParameterValue="$(cat "${pub_key}")"
    
  #echo "Waiting for stack to be created"
  aws --region "${region}" cloudformation wait stack-create-complete --stack-name "${stack_name}" 2> "${err_log}" &
  pid=$!
  spin "${pid}" "Waiting for stack to be created" "Stack created successfully"

  # shellcheck disable=SC2016
  instance_id="$(aws --region "${region}" cloudformation describe-stacks --stack-name "${stack_name}" \
    --query 'Stacks[].Outputs[?OutputKey == `InstanceId`].OutputValue' --output text)"

  aws --region "${region}" ec2 wait instance-status-ok --instance-id "${instance_id}" 2> "${err_log}" &
  pid=$!
  spin "${pid}" "Waiting for stack status to be OK" "Stack status OK"

  # shellcheck disable=SC2016
  public_ip=$(aws --region "${region}" cloudformation describe-stacks --stack-name "${stack_name}" \
    --query 'Stacks[].Outputs[?OutputKey == `PublicIp`].OutputValue' --output text)

  rm -f "${cf_tpl_file}"
  
  echo "PUBLIC IP: ${public_ip}"
}

action_delete() {
  aws --region "${region}" cloudformation delete-stack --stack-name "${stack_name}"
	aws --region "${region}" cloudformation wait stack-delete-complete --stack-name "${stack_name}" 2> "${err_log}" &
  pid=$!
  spin "${pid}" "Waiting for stack to delete" "Stack deleted"
	exit 0
}

action_describe() {
  aws --region "${region}" cloudformation describe-stack-events --stack-name "${stack_name}" --output json | less
	exit 0
}

action_logs() {
  # shellcheck disable=SC2016
  instance_id="$(aws --region "${region}" cloudformation describe-stacks --stack-name "${stack_name}" \
    --query 'Stacks[].Outputs[?OutputKey == `InstanceId`].OutputValue' --output text)"
  
  aws ec2 get-console-output --instance-id "${instance_id}" --output text
}

usage() {
  cat - <<EOF
Script for AWS stack management

Usage:

aws.sh (create|delete|describe|logs) \ 
            --stack-name <name> \ 
            [--inst-type <type>] \ 
            [--region <region>] \ 
            [--pub-key <path>]       

Create

  aws.sh create --stack-name <name> --inst-type <type> [--region <region>] [--pub-key <path>]

Delete|Describe|Logs

  aws.sh (delete|describe|logs) --stack-name <name> [--region <region>]

Arguments:
  --stack-name <name>:  The name of the stack.
  [--inst-type <type>]: (create only) The type of instance to create (e.g. t3.large, c5n.metal).
  [--region <region>]:  (create only) The region where the stack should be created. Defaults to eu-west-1.
  [--pub-key <path>]:   (create only) The path to a .pub file. Defaults to /home/${USER}/.ssh/id_rsa.pub.

EOF
}

action="$1"
shift

case "${action}" in
  create|delete|describe|logs)
    ;;
  *)
    usage
    exit 1
    ;;
esac

stack_name=""
inst_type=""
region=eu-west-1
pub_key="/home/${USER}/.ssh/id_rsa.pub"

while [ $# -gt 0 ]; do
  case "$1" in
    --inst-type|--region|--stack-name|--pub-key)
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

if [ -z "${stack_name}" ] ; then
  usage
  exit 1
fi

"action_${action}"
