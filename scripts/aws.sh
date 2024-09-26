#!/bin/bash

PS4='+ $(date "+%T.%N")\011 '

set -euo pipefail
#set -x

export AWS_PAGER=""

# | Instance Type | Arch   | vCPUs | GiB | Cost  |
# |---------------|--------|-------|-----|-------|
# | t3.large      | x86_64 | 2     | 8   | $0.12 |
# | t3.xlarge     | x86_64 | 4     | 16  | $0.24 |
# | t3.2xlarge    | x86_64 | 8     | 32  | $0.48 |
# | c5n.2xlarge   | x86_64 | 8     | 21  | $0.60 | not good for low lat, hwlatdetect poor
# | c5n.metal     | x86_64 | 72    | 192 | $5.17 |
# | t4g.large     | arm64  | 2     | 8   | $0.10 |
# | t4g.xlarge    | arm64  | 4     | 16  | $0.20 |
# | c6g.metal     | arm64  | 6     | 128 | $3.02 |

#EC2_INSTANCE_TYPE=c7i.metal-24xl
#EC2_INSTANCE_TYPE=c5n.metal # x86
#EC2_INSTANCE_TYPE=t3.2xlarge
#EC2_INSTANCE_TYPE=t3.large
EC2_INSTANCE_TYPE=t3.nano

#EC2_INSTANCE_TYPE=c6g.metal # ARM
#EC2_INSTANCE_TYPE=t4g.medium

# aws ec2 describe-images --region us-west-2 --filters 'Name=name,Values=RHEL-9.*' --query 'Images[*].[Name,ImageId,Architecture]' --output text | sort --reverse
# aws ec2 describe-images --region eu-west-1 --filters 'Name=name,Values=RHEL-9.*' --query 'Images[*].[Name,ImageId,Architecture]' --output text | sort --reverse
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

# OPTIONS: instance type, region, stack name, pub key path
# ACTIONS: create, delete, describe, logs

#REGION=us-west-2
REGION=eu-west-1
stack_name=microshift-pmatusza-stack
cf_tpl_file=cf-gen.yaml

act="${1:-}"

if [[ "${act}" == "delete" ]]; then
	aws --region $REGION cloudformation delete-stack --stack-name $stack_name
    echo "Waiting for stack to delete"
	aws --region $REGION cloudformation wait stack-delete-complete --stack-name $stack_name
	exit 0
fi

if [[ "${act}" == "describe" ]]; then
	aws --region $REGION cloudformation describe-stack-events --stack-name $stack_name --output json | less
	exit 0
fi


EC2_AMI=""
MICROSHIFT_OS=rhel-9.4

ARCH="x86_64"
if [[ "${EC2_INSTANCE_TYPE%.*}" =~ .*"g".* ]]; then
	ARCH="arm64"
fi

if [[ "${EC2_AMI}" == "" ]]; then
	EC2_AMI="${ami_map["${REGION},${ARCH},${MICROSHIFT_OS}"]}"
fi

ec2Type="VirtualMachine"
if [[ "$EC2_INSTANCE_TYPE" =~ c[0-9]+[gn].metal ]]; then
	ec2Type="MetalMachine"
fi

ami_id=${EC2_AMI}
instance_type=${EC2_INSTANCE_TYPE}


echo "Instance type: ${EC2_INSTANCE_TYPE}" 
echo "OS: ${MICROSHIFT_OS}" 
echo "ARCH: ${ARCH}" 
echo "AMI ID: ${EC2_AMI}" 
echo "Region: ${REGION}" 

# TODO default id_rsa.pub, option to give path to different key
public_key="/home/${USER}/.ssh/id_ed25519.pub"

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

aws --region "$REGION" cloudformation create-stack --stack-name "${stack_name}" \
	--template-body "file://${cf_tpl_file}" \
	--capabilities CAPABILITY_NAMED_IAM \
	--parameters \
	ParameterKey=HostInstanceType,ParameterValue="${instance_type}" \
	ParameterKey=Machinename,ParameterValue="${stack_name}" \
	ParameterKey=AmiId,ParameterValue="${ami_id}" \
	ParameterKey=PublicKeyString,ParameterValue="$(cat ${public_key})"

aws --region "${REGION}" cloudformation wait stack-create-complete --stack-name "${stack_name}"

#aws --region "${REGION}" cloudformation describe-stacks --stack-name "${stack_name}"

INSTANCE_ID="$(aws --region "${REGION}" cloudformation describe-stacks --stack-name "${stack_name}" \
	--query 'Stacks[].Outputs[?OutputKey == `InstanceId`].OutputValue' --output text)"

aws --region "${REGION}" ec2 wait instance-status-ok --instance-id "${INSTANCE_ID}"

PUBLIC_IP=$(aws --region "${REGION}" cloudformation describe-stacks --stack-name "${stack_name}" \
	--query 'Stacks[].Outputs[?OutputKey == `PublicIp`].OutputValue' --output text)

echo "PUBLIC IP: ${PUBLIC_IP}"

# cp ~/.ssh/config ~/.ssh/config.bak
# sed -i '/Host aws/{N;N;d;}' ~/.ssh/config
# tee -a ~/.ssh/config << EOF
# Host aws
#   HostName ${PUBLIC_IP}
#   User ec2-user
# EOF
# 
# ssh_cmd() {
#   echo "> ${1}"
#   ssh -o "StrictHostKeyChecking no" aws "$1"
# }
# 
# ssh_cmd "sudo subscription-manager register --org 11009103 --activationkey microshift-rhsm-creds"
# ssh_cmd "sudo subscription-manager config --rhsm.manage_repos=1"
# ssh_cmd "sudo subscription-manager repos --enable rhel-9-for-\$(uname -m)-baseos-rpms --enable rhel-9-for-\$(uname -m)-appstream-rpms"
# 
# ssh_cmd "sudo dnf install -y git vim tmux rsync"
# scp -o "StrictHostKeyChecking no" $HOME/dev/pull-secret-big ec2-user@${PUBLIC_IP}:~/.pull-secret.json
# ssh_cmd "git clone https://github.com/pmtk/microshift.git --branch 4.16/profiling"
# ssh_cmd "git clone https://github.com/pmtk/microshift.git --branch low-latency/warning-job"

# rsync --exclude '.git' --exclude '_output' --exclude 'test/scenario_settings.sh' -avz ~/dev/microshift/ ec2-user@${PUBLIC_IP}:~/microshift

# ssh_cmd "bash -x ./microshift/scripts/devenv-builder/configure-vm.sh --optional-rpms --force-firewall ~/.pull-secret.json"
#ssh_cmd "sudo systemctl disable --now microshift"
#ssh_cmd "sudo microshift-cleanup-data --all --keep-images <<< 1"
#ssh_cmd "sudo rm -rf /var/lib/kubelet"

#ssh_cmd "sudo subscription-manager repos --enable rhel-9-for-x86_64-rt-rpms"
#ssh_cmd "sudo dnf install kernel-rt tuna realtime-setup tuned tuned-profiles* realtime-tests -y"
#ssh_cmd "sudo grubby --set-default=\$(ls /boot/vmlinuz*rt)"

# ssh_cmd "bash -x ./microshift/scripts/image-builder/configure.sh"
# ssh_cmd "bash -x ./microshift/scripts/devenv-builder/manage-vm.sh config"
# ssh_cmd "bash -x ./scripts/image-builder/cleanup.sh -full"

# set +x
# echo
# echo "PUBLIC IP:   ${PUBLIC_IP}"
# echo