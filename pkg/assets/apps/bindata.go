// Code generated for package assets by go-bindata DO NOT EDIT. (@generated)
// sources:
// assets/apps/0000_00_cluster-version-operator_03_deployment.yaml
// assets/apps/0000_50_service-ca-operator_05_deploy.yaml
// assets/apps/0000_60_service-ca_05_deploy.yaml
// assets/apps/0000_70_dns-operator_02-deployment.yaml
// assets/apps/0000_70_dns_01-daemonset.yaml
// assets/apps/0000_80_openshift-router-deployment.yaml
// assets/apps/000_80_hostpath-provisioner-daemonset.yaml
// assets/apps/ovs-ds.yaml
// assets/apps/sdn-controller-ds.yaml
// assets/apps/sdn-ds.yaml
package assets

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type asset struct {
	bytes []byte
	info  os.FileInfo
}

type bindataFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
}

// Name return file name
func (fi bindataFileInfo) Name() string {
	return fi.name
}

// Size return file size
func (fi bindataFileInfo) Size() int64 {
	return fi.size
}

// Mode return file mode
func (fi bindataFileInfo) Mode() os.FileMode {
	return fi.mode
}

// Mode return file modify time
func (fi bindataFileInfo) ModTime() time.Time {
	return fi.modTime
}

// IsDir return file whether a directory
func (fi bindataFileInfo) IsDir() bool {
	return fi.mode&os.ModeDir != 0
}

// Sys return file is sys mode
func (fi bindataFileInfo) Sys() interface{} {
	return nil
}

var _assetsApps0000_00_clusterVersionOperator_03_deploymentYaml = []byte(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: cluster-version-operator
  namespace: openshift-cluster-version
  annotations:
    exclude.release.openshift.io/internal-openshift-hosted: "true"
spec:
  selector:
    matchLabels:
      k8s-app: cluster-version-operator
  strategy:
    type: Recreate
  template:
    metadata:
      name: cluster-version-operator
      labels:
        k8s-app: cluster-version-operator
    spec:
      containers:
      - name: cluster-version-operator
        image: {{.ReleaseImage}}
        imagePullPolicy: IfNotPresent
        args:
          - "start"
          - "--release-image={{.ReleaseImage}}"
          - "--enable-auto-update=false"
          - "--enable-default-cluster-version=true"
          - "--v=4"
        resources:
          requests:
            cpu: 20m
            memory: 50Mi
        terminationMessagePolicy: FallbackToLogsOnError
        volumeMounts:
          - mountPath: /etc/ssl/certs
            name: etc-ssl-certs
            readOnly: true
          - mountPath: /etc/cvo/updatepayloads
            name: etc-cvo-updatepayloads
            readOnly: true
        env:
          - name: KUBERNETES_SERVICE_PORT # allows CVO to communicate with apiserver directly on same host.
            value: "6443"
          - name: KUBERNETES_SERVICE_HOST # allows CVO to communicate with apiserver directly on same host.
            value: "127.0.0.1"
          - name: NODE_NAME
            valueFrom:
              fieldRef:
                fieldPath: spec.nodeName
      hostNetwork: true
      #nodeSelector:
      # node-role.kubernetes.io/master: ""
      priorityClassName: "system-cluster-critical"
      terminationGracePeriodSeconds: 130
      tolerations:
      - key: "node-role.kubernetes.io/master"
        operator: Exists
        effect: "NoSchedule"
      - key: "node.kubernetes.io/unschedulable"
        operator: Exists
        effect: "NoSchedule"
      - key: "node.kubernetes.io/network-unavailable"
        operator: Exists
        effect: "NoSchedule"
      - key: "node.kubernetes.io/not-ready"
        operator: "Exists"
        effect: "NoSchedule" 
      - key: "node.kubernetes.io/unreachable"
        operator: "Exists"
        effect: "NoExecute"
        tolerationSeconds: 120 
      - key: "node.kubernetes.io/not-ready"
        operator: "Exists"
        effect: "NoExecute"
        tolerationSeconds: 120 
      volumes:
        - name: etc-ssl-certs
          hostPath:
            path: /etc/ssl/certs
        - name: etc-cvo-updatepayloads
          hostPath:
            path: /etc/cvo/updatepayloads
`)

func assetsApps0000_00_clusterVersionOperator_03_deploymentYamlBytes() ([]byte, error) {
	return _assetsApps0000_00_clusterVersionOperator_03_deploymentYaml, nil
}

func assetsApps0000_00_clusterVersionOperator_03_deploymentYaml() (*asset, error) {
	bytes, err := assetsApps0000_00_clusterVersionOperator_03_deploymentYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/apps/0000_00_cluster-version-operator_03_deployment.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsApps0000_50_serviceCaOperator_05_deployYaml = []byte(`apiVersion: apps/v1
kind: Deployment
metadata:
  namespace: openshift-service-ca-operator
  name: service-ca-operator
  labels:
    app: service-ca-operator
spec:
  replicas: 1
  selector:
    matchLabels:
      app: service-ca-operator
  template:
    metadata:
      name: service-ca-operator
      labels:
        app: service-ca-operator
    spec:
      serviceAccountName: service-ca-operator
      containers:
      - name: service-ca-operator
        image: {{.ImageSCOperator}}
        imagePullPolicy: IfNotPresent
        command: ["service-ca-operator", "operator"]
        args:
        - "--config=/var/run/configmaps/config/operator-config.yaml"
        - "-v=4"
        resources:
          requests:
            memory: 80Mi
            cpu: 10m
        env:
        - name: CONTROLLER_IMAGE
          value: {{.ImageSCOperator}}
        - name: OPERATOR_IMAGE_VERSION
          value: {{.VersionSCOperator}}
        volumeMounts:
        - mountPath: /var/run/configmaps/config
          name: config
        - mountPath: /var/run/secrets/serving-cert
          name: serving-cert
      volumes:
      - name: serving-cert
        hostPath:
          path: {{.CertDir}}
      - name: config
        configMap:
          defaultMode: 440
          name: service-ca-operator-config
      #nodeSelector:
      #  node-role.kubernetes.io/master: ""
      priorityClassName: "system-cluster-critical"
      tolerations:
      - key: node-role.kubernetes.io/master
        operator: Exists
        effect: "NoSchedule"
      - key: "node.kubernetes.io/unreachable"
        operator: "Exists"
        effect: "NoExecute"
        tolerationSeconds: 120
      - key: "node.kubernetes.io/not-ready"
        operator: "Exists"
        effect: "NoExecute"
        tolerationSeconds: 120
`)

func assetsApps0000_50_serviceCaOperator_05_deployYamlBytes() ([]byte, error) {
	return _assetsApps0000_50_serviceCaOperator_05_deployYaml, nil
}

func assetsApps0000_50_serviceCaOperator_05_deployYaml() (*asset, error) {
	bytes, err := assetsApps0000_50_serviceCaOperator_05_deployYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/apps/0000_50_service-ca-operator_05_deploy.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsApps0000_60_serviceCa_05_deployYaml = []byte(`apiVersion: apps/v1
kind: Deployment
metadata:
  namespace: openshift-service-ca
  name: service-ca
  labels:
    app: service-ca
spec:
  replicas: 1
  selector:
    matchLabels:
      app: service-ca
  template:
    metadata:
      name: service-ca
      labels:
        app: service-ca
    spec:
      serviceAccountName: service-ca
      containers:
      - name: service-ca-controller
        image: quay.io/openshift/okd-content@sha256:d5ab863a154efd4014b0e1d9f753705b97a3f3232bd600c0ed9bde71293c462e
        imagePullPolicy: IfNotPresent
        command: ["service-ca-operator", "controller"]
        args:
        - "-v=4"
        ports:
          - containerPort: 8443
            protocol: TCP
        resources:
          requests:
            memory: 120Mi
            cpu: 10m
        volumeMounts:
          - mountPath: /var/run/secrets/signing-key
            name: signing-key
          - mountPath: /var/run/configmaps/signing-cabundle
            name: signing-cabundle
      volumes:
        - name: signing-key
          hostPath:
            path: {{.KeyDir}}
        - name: signing-cabundle
          hostPath:
            path: {{.CADir}}
      #nodeSelector:
      #  node-role.kubernetes.io/master: ""
      priorityClassName: "system-cluster-critical"
      tolerations:
      - key: node-role.kubernetes.io/master
        operator: Exists
        effect: "NoSchedule"
      - key: "node.kubernetes.io/unreachable"
        operator: "Exists"
        effect: "NoExecute"
        tolerationSeconds: 120
      - key: "node.kubernetes.io/not-ready"
        operator: "Exists"
        effect: "NoExecute"
        tolerationSeconds: 120
`)

func assetsApps0000_60_serviceCa_05_deployYamlBytes() ([]byte, error) {
	return _assetsApps0000_60_serviceCa_05_deployYaml, nil
}

func assetsApps0000_60_serviceCa_05_deployYaml() (*asset, error) {
	bytes, err := assetsApps0000_60_serviceCa_05_deployYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/apps/0000_60_service-ca_05_deploy.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsApps0000_70_dnsOperator_02DeploymentYaml = []byte(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: dns-operator
  namespace: openshift-dns-operator
spec:
  replicas: 1
  selector:
    matchLabels:
      name: dns-operator
  template:
    metadata:
      labels:
        name: dns-operator
    spec:
      dnsPolicy: Default
      nodeSelector:
        kubernetes.io/os: linux
        #node-role.kubernetes.io/master: ''
      restartPolicy: Always
      priorityClassName: system-cluster-critical
      serviceAccountName: dns-operator
      containers:
      - name: dns-operator
        terminationMessagePolicy: FallbackToLogsOnError
        image: {{.ImageDNSOperator}}
        command:
        - dns-operator
        env:
        - name: RELEASE_VERSION
          value: {{.VersionDNSOperator}}
        - name: IMAGE
          value: {{.ImageCoreDNS}}
        - name: OPENSHIFT_CLI_IMAGE
          value: {{.ImageOC}}
        - name: KUBE_RBAC_PROXY_IMAGE
          value: {{.ImageKubeRbacProxy}}
        resources:
          requests:
            cpu: 10m
      - name: kube-rbac-proxy
        image: {{.ImageKubeRbacProxy}}
        args:
        - --logtostderr
        - --insecure-listen-address=:9393
        - --upstream=http://127.0.0.1:60000/
        ports:
        - containerPort: 9393
          name: metrics
        resources:
          requests:
            cpu: 10m
            memory: 40Mi
      terminationGracePeriodSeconds: 2
      tolerations:
      - key: "node-role.kubernetes.io/master"
        operator: "Exists"
        effect: "NoSchedule"
      - key: "node.kubernetes.io/unreachable"
        operator: "Exists"
        effect: "NoExecute"
        tolerationSeconds: 120
      - key: "node.kubernetes.io/not-ready"
        operator: "Exists"
        effect: "NoExecute"
        tolerationSeconds: 120
      
`)

func assetsApps0000_70_dnsOperator_02DeploymentYamlBytes() ([]byte, error) {
	return _assetsApps0000_70_dnsOperator_02DeploymentYaml, nil
}

func assetsApps0000_70_dnsOperator_02DeploymentYaml() (*asset, error) {
	bytes, err := assetsApps0000_70_dnsOperator_02DeploymentYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/apps/0000_70_dns-operator_02-deployment.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsApps0000_70_dns_01DaemonsetYaml = []byte(`kind: DaemonSet
apiVersion: apps/v1
metadata:
  labels:
    dns.operator.openshift.io/owning-dns: default
  name: dns-default
  namespace: openshift-dns    
spec:
  selector:
    matchLabels:
      dns.operator.openshift.io/daemonset-dns: default
  template:
    metadata:
      labels:
        dns.operator.openshift.io/daemonset-dns: default
    spec:
      serviceAccountName: dns
      priorityClassName: system-node-critical
      containers:
      - name: dns
        image: quay.io/openshift/okd-content@sha256:fb7eafdcb7989575119e1807e4adc2eb29f8165dec5c148b9c3a44d48458d8a7
        imagePullPolicy: IfNotPresent
        terminationMessagePolicy: FallbackToLogsOnError
        command: [ "coredns" ]
        args: [ "-conf", "/etc/coredns/Corefile" ]
        volumeMounts:
        - name: config-volume
          mountPath: /etc/coredns
          readOnly: true
        ports:
        - containerPort: 5353
          name: dns
          protocol: UDP
        - containerPort: 5353
          name: dns-tcp
          protocol: TCP
        readinessProbe:
          failureThreshold: 3
          httpGet:
            path: /health
            port: 8080
            scheme: HTTP
          initialDelaySeconds: 10
          periodSeconds: 10
          successThreshold: 1
          timeoutSeconds: 10
        livenessProbe:
          failureThreshold: 5
          httpGet:
            path: /health
            port: 8080
            scheme: HTTP
          initialDelaySeconds: 60
          timeoutSeconds: 5
          successThreshold: 1
        resources:
          requests:
            cpu: 50m
            memory: 70Mi
      - name: kube-rbac-proxy
        image: quay.io/openshift/okd-content@sha256:1aa5bb03d0485ec2db2c7871a1eeaef83e9eabf7e9f1bc2c841cf1a759817c99
        args:
        - --logtostderr
        - --secure-listen-address=:9154
        - --tls-cipher-suites=TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,TLS_RSA_WITH_AES_128_CBC_SHA256,TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256,TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256
        - --upstream=http://127.0.0.1:9153/
        - --tls-cert-file=/etc/tls/private/tls.crt
        - --tls-private-key-file=/etc/tls/private/tls.key
        ports:
        - containerPort: 9154
          name: metrics
        resources:
          requests:
            cpu: 10m
            memory: 40Mi
        volumeMounts:
        - mountPath: /etc/tls/private
          name: metrics-tls
          readOnly: true
      - name: dns-node-resolver
        image: quay.io/openshift/okd-content@sha256:b20d195c721cd3b6215e5716b5569cbabbe861559af7dce07b5f8f3d38e6d701
        imagePullPolicy: IfNotPresent
        terminationMessagePolicy: FallbackToLogsOnError
        securityContext:
          privileged: true
        volumeMounts:
        - name: hosts-file
          mountPath: /etc/hosts
        env:
        - name: SERVICES
          value: "image-registry.openshift-image-registry.svc"
        - name: CLUSTER_DOMAIN
          value: cluster.local        
        command:
        - /bin/bash
        - -c
        - |
          #!/bin/bash
          set -uo pipefail
          NAMESERVER=${DNS_DEFAULT_SERVICE_HOST}

          trap 'jobs -p | xargs kill || true; wait; exit 0' TERM

          OPENSHIFT_MARKER="openshift-generated-node-resolver"
          HOSTS_FILE="/etc/hosts"
          TEMP_FILE="/etc/hosts.tmp"

          IFS=', ' read -r -a services <<< "${SERVICES}"

          # Make a temporary file with the old hosts file's attributes.
          cp -f --attributes-only "${HOSTS_FILE}" "${TEMP_FILE}"

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
              grep -v "# ${OPENSHIFT_MARKER}" "${HOSTS_FILE}" > "${TEMP_FILE}"

              # Append resolver entries for services
              for svc in "${!svc_ips[@]}"; do
                for ip in ${svc_ips[${svc}]}; do
                  echo "${ip} ${svc} ${svc}.${CLUSTER_DOMAIN} # ${OPENSHIFT_MARKER}" >> "${TEMP_FILE}"
                done
              done

              # TODO: Update /etc/hosts atomically to avoid any inconsistent behavior
              # Replace /etc/hosts with our modified version if needed
              cmp "${TEMP_FILE}" "${HOSTS_FILE}" || cp -f "${TEMP_FILE}" "${HOSTS_FILE}"
              # TEMP_FILE is not removed to avoid file create/delete and attributes copy churn
            fi
            sleep 60 & wait
            unset svc_ips
          done
        resources:
          requests:
            cpu: 5m
            memory: 21Mi
      dnsPolicy: Default
      nodeSelector:
        kubernetes.io/os: linux      
      volumes:
      - name: config-volume
        configMap:
          defaultMode: 420
          items:
          - key: Corefile
            path: Corefile
          name: dns-default
      - name: hosts-file
        hostPath:
          path: /etc/hosts
          type: File
      - name: metrics-tls
        secret:
          defaultMode: 420
          secretName: dns-default-metrics-tls
      tolerations:
      # DNS needs to run everywhere. Tolerate all taints
      - operator: Exists
  updateStrategy:
    type: RollingUpdate
    rollingUpdate:
      # Note: The daemon controller rounds the percentage up
      # (unlike the deployment controller, which rounds down).
      maxUnavailable: 10%
`)

func assetsApps0000_70_dns_01DaemonsetYamlBytes() ([]byte, error) {
	return _assetsApps0000_70_dns_01DaemonsetYaml, nil
}

func assetsApps0000_70_dns_01DaemonsetYaml() (*asset, error) {
	bytes, err := assetsApps0000_70_dns_01DaemonsetYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/apps/0000_70_dns_01-daemonset.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsApps0000_80_openshiftRouterDeploymentYaml = []byte(`kind: Deployment
apiVersion: apps/v1
metadata:
  name: router-default
  namespace: openshift-ingress
  labels:
    ingresscontroller.operator.openshift.io/deployment-ingresscontroller: default
spec:
  progressDeadlineSeconds: 600
  selector:
    matchLabels:
      ingresscontroller.operator.openshift.io/deployment-ingresscontroller: default
  template:
    metadata:
      labels:
        ingresscontroller.operator.openshift.io/deployment-ingresscontroller: default
    spec:
      serviceAccountName: router
      # nodeSelector is set at runtime.
      priorityClassName: system-cluster-critical
      containers:
        - name: router
          image: quay.io/openshift/okd-content@sha256:5908265eb0041cea9a9ec36ad7b2bc82dd45346fc9e0f1b34b0e38a0f43f9f18
          imagePullPolicy: IfNotPresent
          terminationMessagePolicy: FallbackToLogsOnError
          ports:
          - name: http
            containerPort: 80
            protocol: TCP
          - name: https
            containerPort: 443
            protocol: TCP
          - name: metrics
            containerPort: 1936
            protocol: TCP
          # Merged at runtime.
          env:
          # stats username and password are generated at runtime
          - name: STATS_PORT
            value: "1936"
          - name: ROUTER_SERVICE_NAMESPACE
            value: openshift-ingress
          - name: DEFAULT_CERTIFICATE_DIR
            value: /etc/pki/tls/private
          - name: DEFAULT_DESTINATION_CA_PATH
            value: /var/run/configmaps/service-ca/service-ca.crt
          - name: ROUTER_CIPHERS
            value: TLS_AES_128_GCM_SHA256:TLS_AES_256_GCM_SHA384:TLS_CHACHA20_POLY1305_SHA256:ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-ECDSA-CHACHA20-POLY1305:ECDHE-RSA-CHACHA20-POLY1305:DHE-RSA-AES128-GCM-SHA256:DHE-RSA-AES256-GCM-SHA384
          - name: ROUTER_DISABLE_HTTP2
            value: "true"
          - name: ROUTER_DISABLE_NAMESPACE_OWNERSHIP_CHECK
            value: "false"
          #FIXME: use metrics tls
          - name: ROUTER_METRICS_TLS_CERT_FILE
            value: /etc/pki/tls/private/tls.crt
          - name: ROUTER_METRICS_TLS_KEY_FILE
            value: /etc/pki/tls/private/tls.key
          - name: ROUTER_METRICS_TYPE
            value: haproxy
          - name: ROUTER_SERVICE_NAME
            value: default
          - name: ROUTER_SET_FORWARDED_HEADERS
            value: append
          - name: ROUTER_THREADS
            value: "4"
          - name: SSL_MIN_VERSION
            value: TLSv1.2            
          livenessProbe:
            failureThreshold: 3
            httpGet:
              path: /healthz
              port: 1936
              scheme: HTTP
            initialDelaySeconds: 10
            periodSeconds: 10
            successThreshold: 1
            timeoutSeconds: 1
          readinessProbe:
             failureThreshold: 3
             httpGet:
               path: /healthz/ready
               port: 1936
               scheme: HTTP
             initialDelaySeconds: 10
             periodSeconds: 10
             successThreshold: 1
             timeoutSeconds: 1
          resources:
            requests:
              cpu: 100m
              memory: 256Mi
          securityContext:
            privileged: true              
          volumeMounts:
          - mountPath: /etc/pki/tls/private
            name: default-certificate
            readOnly: true
          - mountPath: /var/run/configmaps/service-ca
            name: service-ca-bundle
            readOnly: true
      volumes:
      - name: default-certificate
        secret:
          secretName: router-certs-default
      - name: service-ca-bundle
        configMap:
          items:
          - key: service-ca.crt
            path: service-ca.crt
          name: service-ca-bundle
          optional: false
`)

func assetsApps0000_80_openshiftRouterDeploymentYamlBytes() ([]byte, error) {
	return _assetsApps0000_80_openshiftRouterDeploymentYaml, nil
}

func assetsApps0000_80_openshiftRouterDeploymentYaml() (*asset, error) {
	bytes, err := assetsApps0000_80_openshiftRouterDeploymentYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/apps/0000_80_openshift-router-deployment.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsApps000_80_hostpathProvisionerDaemonsetYaml = []byte(`apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: kubevirt-hostpath-provisioner
  labels:
    k8s-app: kubevirt-hostpath-provisioner
  namespace: kubevirt-hostpath-provisioner
spec:
  selector:
    matchLabels:
      k8s-app: kubevirt-hostpath-provisioner
  template:
    metadata:
      labels:
        k8s-app: kubevirt-hostpath-provisioner
    spec:
      serviceAccountName: kubevirt-hostpath-provisioner-admin
      containers:
        - name: kubevirt-hostpath-provisioner
          image: quay.io/kubevirt/hostpath-provisioner
          imagePullPolicy: Always
          env:
            - name: USE_NAMING_PREFIX
              value: "false" # change to true, to have the name of the pvc be part of the directory
            - name: NODE_NAME
              valueFrom:
                fieldRef:
                  fieldPath: spec.nodeName
            - name: PV_DIR
              value: /var/hpvolumes
          volumeMounts:
            - name: pv-volume # root dir where your bind mounts will be on the node
              mountPath: /var/hpvolumes
              #nodeSelector:
              #- name: xxxxxx
      volumes:
        - name: pv-volume
          hostPath:
            path: /var/hpvolumes`)

func assetsApps000_80_hostpathProvisionerDaemonsetYamlBytes() ([]byte, error) {
	return _assetsApps000_80_hostpathProvisionerDaemonsetYaml, nil
}

func assetsApps000_80_hostpathProvisionerDaemonsetYaml() (*asset, error) {
	bytes, err := assetsApps000_80_hostpathProvisionerDaemonsetYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/apps/000_80_hostpath-provisioner-daemonset.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsAppsOvsDsYaml = []byte(`apiVersion: apps/v1
kind: DaemonSet
metadata:
  labels:
    app: ovs
  name: ovs
  namespace: openshift-sdn
spec:
  selector:
    matchLabels:
      app: ovs
  template:
    metadata:
      creationTimestamp: null
      labels:
        app: ovs
        component: network
        openshift.io/component: network
        type: infra
    spec:
      affinity:
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
            - matchExpressions:
              - key: network.operator.openshift.io/external-openvswitch
                operator: DoesNotExist
      containers:
      - command:
        - /bin/bash
        - -c
        - |
          #!/bin/bash
          set -euo pipefail
          export SYSTEMD_IGNORE_CHROOT=yes

          # systemctl cannot be used in a separate PID namespace to reach
          # the systemd running in PID 1. Therefore we need to use the dbus API
          systemctl_restart(){
            gdbus call \
              --system \
              --dest org.freedesktop.systemd1 \
              --object-path /org/freedesktop/systemd1/unit/"$(svc_encode_name ${1})"_2eservice \
              --method org.freedesktop.systemd1.Unit.Restart "replace"
          }
          svc_encode_name(){
            # systemd encodes some characters, so far we only need to encode
            # the character "-" but there may be more in the future.
            echo "${1//-/_2d}"
          }

            # In some very strange corner cases, the owner for /run/openvswitch
            # can be wrong, so we need to clean up and restart.
            ovs_uid=$(chroot /host id -u openvswitch)
            ovs_gid=$(chroot /host id -g openvswitch)
            chown -R "${ovs_uid}:${ovs_gid}" /run/openvswitch
            if [[ ! -S /run/openvswitch/db.sock ]]; then
              systemctl_restart ovsdb-server
            fi
            # We need to explicitly exit on SIGTERM, see https://github.com/openshift/cluster-dns-operator/issues/65
            function quit {
                exit 0
            }
            trap quit SIGTERM
            # Don't need to worry about restoring flows; this can only change if we've rebooted
            tail --pid=$BASHPID -F /host/var/log/openvswitch/ovs-vswitchd.log /host/var/log/openvswitch/ovsdb-server.log &
            wait
        image: quay.io/openshift/okd-content@sha256:71dbab00e9803acb3bcf859607d9d3ed445b6f3a063ecedd6b3ea02a7a8fdd80
        imagePullPolicy: IfNotPresent
        name: openvswitch
        resources:
          requests:
            cpu: 15m
            memory: 400Mi
        securityContext:
          privileged: true
        terminationMessagePath: /dev/termination-log
        terminationMessagePolicy: FallbackToLogsOnError
        volumeMounts:
        - mountPath: /lib/modules
          name: host-modules
          readOnly: true
        - mountPath: /run
          name: host-run
        - mountPath: /sys
          name: host-sys
          readOnly: true
        - mountPath: /etc/openvswitch
          name: host-config-openvswitch
        - mountPath: /host
          name: host-slash
          readOnly: true
      dnsPolicy: ClusterFirst
      hostNetwork: true
      nodeSelector:
        kubernetes.io/os: linux
      priorityClassName: system-node-critical
      restartPolicy: Always
      schedulerName: default-scheduler
      securityContext: {}
      #serviceAccount: sdn
      #serviceAccountName: sdn
      terminationGracePeriodSeconds: 30
      tolerations:
      - operator: Exists
      volumes:
      - hostPath:
          path: /lib/modules
          type: ""
        name: host-modules
      - hostPath:
          path: /run
          type: ""
        name: host-run
      - hostPath:
          path: /sys
          type: ""
        name: host-sys
      - hostPath:
          path: /var/lib/openvswitch
          type: DirectoryOrCreate
        name: host-config-openvswitch
      - hostPath:
          path: /
          type: ""
        name: host-slash
  updateStrategy:
    rollingUpdate:
      maxUnavailable: 1
    type: RollingUpdate
`)

func assetsAppsOvsDsYamlBytes() ([]byte, error) {
	return _assetsAppsOvsDsYaml, nil
}

func assetsAppsOvsDsYaml() (*asset, error) {
	bytes, err := assetsAppsOvsDsYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/apps/ovs-ds.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsAppsSdnControllerDsYaml = []byte(`apiVersion: apps/v1
kind: DaemonSet
metadata:
  labels:
    app: sdn-controller
  name: sdn-controller
  namespace: openshift-sdn
spec:
  selector:
    matchLabels:
      app: sdn-controller
  template:
    metadata:
      creationTimestamp: null
      labels:
        app: sdn-controller
    spec:
      containers:
      - command:
        - /bin/bash
        - -c
        - |
          if [[ -f /env/_master ]]; then
            set -o allexport
            source /env/_master
            set +o allexport
          fi

          exec openshift-sdn-controller --v=${OPENSHIFT_SDN_LOG_LEVEL:-2}
        env:
        - name: KUBERNETES_SERVICE_PORT
          value: "6443"
        - name: KUBERNETES_SERVICE_HOST
          value: api-int.crc.testing
        image: quay.io/openshift/okd-content@sha256:71dbab00e9803acb3bcf859607d9d3ed445b6f3a063ecedd6b3ea02a7a8fdd80
        imagePullPolicy: IfNotPresent
        name: sdn-controller
        resources:
          requests:
            cpu: 10m
            memory: 50Mi
        terminationMessagePath: /dev/termination-log
        terminationMessagePolicy: FallbackToLogsOnError
        volumeMounts:
        - mountPath: /env
          name: env-overrides
      dnsPolicy: ClusterFirst
      hostNetwork: true
      nodeSelector:
        node-role.kubernetes.io/master: ""
      priorityClassName: system-cluster-critical
      restartPolicy: Always
      schedulerName: default-scheduler
      securityContext:
        runAsNonRoot: true
        runAsUser: 65534
      #serviceAccount: sdn-controller
      #serviceAccountName: sdn-controller
      terminationGracePeriodSeconds: 30
      tolerations:
      - effect: NoSchedule
        key: node-role.kubernetes.io/master
        operator: Exists
      - effect: NoSchedule
        key: node.kubernetes.io/not-ready
        operator: Exists
      volumes:
      - configMap:
          defaultMode: 420
          name: env-overrides
          optional: true
        name: env-overrides
  updateStrategy:
    rollingUpdate:
      maxUnavailable: 1
    type: RollingUpdate
`)

func assetsAppsSdnControllerDsYamlBytes() ([]byte, error) {
	return _assetsAppsSdnControllerDsYaml, nil
}

func assetsAppsSdnControllerDsYaml() (*asset, error) {
	bytes, err := assetsAppsSdnControllerDsYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/apps/sdn-controller-ds.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsAppsSdnDsYaml = []byte(`apiVersion: apps/v1
kind: DaemonSet
metadata:
  labels:
    app: sdn
  name: sdn
  namespace: openshift-sdn
spec:
  selector:
    matchLabels:
      app: sdn
  template:
    metadata:
      labels:
        app: sdn
        component: network
        openshift.io/component: network
        type: infra
    spec:
      containers:
      - command:
        - /bin/bash
        - -c
        - |
          #!/bin/bash
          set -euo pipefail

          # if another process is listening on the cni-server socket, wait until it exits
          trap 'kill $(jobs -p); rm -f /etc/cni/net.d/80-openshift-network.conf ; exit 0' TERM
          retries=0
          while true; do
            if echo 'test' | socat - UNIX-CONNECT:/var/run/openshift-sdn/cniserver/socket &>/dev/null; then
              echo "warning: Another process is currently listening on the CNI socket, waiting 15s ..." 2>&1
              sleep 15 & wait
              (( retries += 1 ))
            else
              break
            fi
            if [[ "${retries}" -gt 40 ]]; then
              echo "error: Another process is currently listening on the CNI socket, exiting" 2>&1
              exit 1
            fi
          done

          # local environment overrides
          if [[ -f /etc/sysconfig/openshift-sdn ]]; then
            set -o allexport
            source /etc/sysconfig/openshift-sdn
            set +o allexport
          fi
          #BUG: cdc accidentally mounted /etc/sysconfig/openshift-sdn as DirectoryOrCreate; clean it up so we can ultimately mount /etc/sysconfig/openshift-sdn as FileOrCreate
          # Once this is released, then we can mount it properly
          if [[ -d /etc/sysconfig/openshift-sdn ]]; then
            rmdir /etc/sysconfig/openshift-sdn || true
          fi

          # configmap-based overrides
          if [[ -f /env/${K8S_NODE_NAME} ]]; then
            set -o allexport
            source /env/${K8S_NODE_NAME}
            set +o allexport
          fi

          # Take over network functions on the node
          rm -f /etc/cni/net.d/80-openshift-network.conf
          cp -f /opt/cni/bin/openshift-sdn /host/opt/cni/bin/

          # Launch the network process
          exec /usr/bin/openshift-sdn-node \
            --node-name ${K8S_NODE_NAME} --node-ip ${K8S_NODE_IP} \
            --proxy-config /config/kube-proxy-config.yaml \
            --v ${OPENSHIFT_SDN_LOG_LEVEL:-2}
        env:
        - name: KUBERNETES_SERVICE_PORT
          value: "6443"
        - name: KUBERNETES_SERVICE_HOST
          value: 127.0.0.1
        - name: OPENSHIFT_DNS_DOMAIN
          value: ushift.testing
        - name: K8S_NODE_NAME
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: spec.nodeName
        - name: K8S_NODE_IP
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: status.hostIP
        image: quay.io/openshift/okd-content@sha256:71dbab00e9803acb3bcf859607d9d3ed445b6f3a063ecedd6b3ea02a7a8fdd80
        imagePullPolicy: IfNotPresent
        lifecycle:
          preStop:
            exec:
              command:
              - rm
              - -f
              - /etc/cni/net.d/80-openshift-network.conf
              - /host/opt/cni/bin/openshift-sdn
        name: sdn
        ports:
        - containerPort: 10256
          hostPort: 10256
          name: healthz
          protocol: TCP
        readinessProbe:
          exec:
            command:
            - test
            - -f
            - /etc/cni/net.d/80-openshift-network.conf
          failureThreshold: 3
          initialDelaySeconds: 5
          periodSeconds: 5
          successThreshold: 1
          timeoutSeconds: 1
        resources:
          requests:
            cpu: 100m
            memory: 200Mi
        securityContext:
          privileged: true
        terminationMessagePath: /dev/termination-log
        terminationMessagePolicy: FallbackToLogsOnError
        volumeMounts:
        - mountPath: /config
          name: config
          readOnly: true
        - mountPath: /env
          name: env-overrides
        - mountPath: /var/run
          name: host-var-run
        - mountPath: /var/run/dbus/
          name: host-var-run-dbus
          readOnly: true
        - mountPath: /var/run/openvswitch/
          name: host-var-run-ovs
          readOnly: true
        - mountPath: /var/run/kubernetes/
          name: host-var-run-kubernetes
          readOnly: true
        - mountPath: /run/netns
          mountPropagation: HostToContainer
          name: host-run-netns
          readOnly: true
        - mountPath: /host/var/run/netns
          mountPropagation: HostToContainer
          name: host-var-run-netns
          readOnly: true
        - mountPath: /var/run/openshift-sdn
          name: host-var-run-openshift-sdn
        - mountPath: /host
          mountPropagation: HostToContainer
          name: host-slash
          readOnly: true
        - mountPath: /host/opt/cni/bin
          name: host-cni-bin
        - mountPath: /etc/cni/net.d
          name: host-cni-conf
        - mountPath: /var/lib/cni/networks/openshift-sdn
          name: host-var-lib-cni-networks-openshift-sdn
        - mountPath: /lib/modules
          name: host-modules
          readOnly: true
        - mountPath: /etc/sysconfig
          name: etc-sysconfig
          readOnly: true
      - command:
        - /bin/bash
        - -c
        - |
          #!/bin/bash
          set -euo pipefail
          TLS_PK=/etc/pki/tls/metrics-certs/tls.key
          TLS_CERT=/etc/pki/tls/metrics-certs/tls.crt

          # As the secret mount is optional we must wait for the files to be present.
          # The service is created in monitor.yaml and this is created in sdn.yaml.
          # If it isn't created there is probably an issue so we want to crashloop.
          TS=$(date +%s)
          WARN_TS=$(( ${TS} + $(( 20 * 60)) ))
          HAS_LOGGED_INFO=0

          log_missing_certs(){
              CUR_TS=$(date +%s)
              if [[ "${CUR_TS}" -gt "WARN_TS"  ]]; then
                echo $(date -Iseconds) WARN: sdn-metrics-certs not mounted after 20 minutes.
              elif [[ "${HAS_LOGGED_INFO}" -eq 0 ]] ; then
                echo $(date -Iseconds) INFO: sdn-metrics-certs not mounted. Waiting 20 minutes.
                HAS_LOGGED_INFO=1
              fi
          }

          while [[ ! -f "${TLS_PK}" ||  ! -f "${TLS_CERT}" ]] ; do
            log_missing_certs
            sleep 5
          done

          echo $(date -Iseconds) INFO: sdn-metrics-certs mounted, starting kube-rbac-proxy
          exec /usr/bin/kube-rbac-proxy \
            --logtostderr \
            --secure-listen-address=:9101 \
            --tls-cipher-suites=TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,TLS_RSA_WITH_AES_128_CBC_SHA256,TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256,TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256 \
            --upstream=http://127.0.0.1:29101/ \
            --tls-private-key-file=${TLS_PK} \
            --tls-cert-file=${TLS_CERT}
        image: quay.io/openshift/okd-content@sha256:1aa5bb03d0485ec2db2c7871a1eeaef83e9eabf7e9f1bc2c841cf1a759817c99
        imagePullPolicy: IfNotPresent
        name: kube-rbac-proxy
        ports:
        - containerPort: 9101
          hostPort: 9101
          name: https
          protocol: TCP
        resources:
          requests:
            cpu: 10m
            memory: 20Mi
        terminationMessagePath: /dev/termination-log
        terminationMessagePolicy: FallbackToLogsOnError
        volumeMounts:
        - mountPath: /etc/pki/tls/metrics-certs
          name: sdn-metrics-certs
          readOnly: true
      dnsPolicy: ClusterFirst
      hostNetwork: true
      hostPID: true
      nodeSelector:
        kubernetes.io/os: linux
      priorityClassName: system-node-critical
      restartPolicy: Always
      schedulerName: default-scheduler
      securityContext: {}
      #serviceAccount: sdn
      #serviceAccountName: sdn
      terminationGracePeriodSeconds: 30
      tolerations:
      - operator: Exists
      volumes:
      - configMap:
          defaultMode: 420
          name: sdn-config
        name: config
      - configMap:
          defaultMode: 420
          name: env-overrides
          optional: true
        name: env-overrides
      - hostPath:
          path: /etc/sysconfig
          type: ""
        name: etc-sysconfig
      - hostPath:
          path: /lib/modules
          type: ""
        name: host-modules
      - hostPath:
          path: /var/run
          type: ""
        name: host-var-run
      - hostPath:
          path: /run/netns
          type: ""
        name: host-run-netns
      - hostPath:
          path: /var/run/netns
          type: ""
        name: host-var-run-netns
      - hostPath:
          path: /var/run/dbus
          type: ""
        name: host-var-run-dbus
      - hostPath:
          path: /var/run/openvswitch
          type: ""
        name: host-var-run-ovs
      - hostPath:
          path: /var/run/kubernetes
          type: ""
        name: host-var-run-kubernetes
      - hostPath:
          path: /var/run/openshift-sdn
          type: ""
        name: host-var-run-openshift-sdn
      - hostPath:
          path: /
          type: ""
        name: host-slash
      - hostPath:
          path: /var/lib/cni/bin
          type: ""
        name: host-cni-bin
      - hostPath:
          path: /var/run/multus/cni/net.d
          type: ""
        name: host-cni-conf
      - hostPath:
          path: /var/lib/cni/networks/openshift-sdn
          type: ""
        name: host-var-lib-cni-networks-openshift-sdn
      - name: sdn-metrics-certs
        secret:
          defaultMode: 420
          optional: true
          secretName: sdn-metrics-certs
  updateStrategy:
    rollingUpdate:
      maxUnavailable: 1
    type: RollingUpdate
`)

func assetsAppsSdnDsYamlBytes() ([]byte, error) {
	return _assetsAppsSdnDsYaml, nil
}

func assetsAppsSdnDsYaml() (*asset, error) {
	bytes, err := assetsAppsSdnDsYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/apps/sdn-ds.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("Asset %s can't read by error: %v", name, err)
		}
		return a.bytes, nil
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// MustAsset is like Asset but panics when Asset would return an error.
// It simplifies safe initialization of global variables.
func MustAsset(name string) []byte {
	a, err := Asset(name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}

// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func AssetInfo(name string) (os.FileInfo, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("AssetInfo %s can't read by error: %v", name, err)
		}
		return a.info, nil
	}
	return nil, fmt.Errorf("AssetInfo %s not found", name)
}

// AssetNames returns the names of the assets.
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string]func() (*asset, error){
	"assets/apps/0000_00_cluster-version-operator_03_deployment.yaml": assetsApps0000_00_clusterVersionOperator_03_deploymentYaml,
	"assets/apps/0000_50_service-ca-operator_05_deploy.yaml":          assetsApps0000_50_serviceCaOperator_05_deployYaml,
	"assets/apps/0000_60_service-ca_05_deploy.yaml":                   assetsApps0000_60_serviceCa_05_deployYaml,
	"assets/apps/0000_70_dns-operator_02-deployment.yaml":             assetsApps0000_70_dnsOperator_02DeploymentYaml,
	"assets/apps/0000_70_dns_01-daemonset.yaml":                       assetsApps0000_70_dns_01DaemonsetYaml,
	"assets/apps/0000_80_openshift-router-deployment.yaml":            assetsApps0000_80_openshiftRouterDeploymentYaml,
	"assets/apps/000_80_hostpath-provisioner-daemonset.yaml":          assetsApps000_80_hostpathProvisionerDaemonsetYaml,
	"assets/apps/ovs-ds.yaml":                                         assetsAppsOvsDsYaml,
	"assets/apps/sdn-controller-ds.yaml":                              assetsAppsSdnControllerDsYaml,
	"assets/apps/sdn-ds.yaml":                                         assetsAppsSdnDsYaml,
}

// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//     data/
//       foo.txt
//       img/
//         a.png
//         b.png
// then AssetDir("data") would return []string{"foo.txt", "img"}
// AssetDir("data/img") would return []string{"a.png", "b.png"}
// AssetDir("foo.txt") and AssetDir("notexist") would return an error
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		cannonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(cannonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, fmt.Errorf("Asset %s not found", name)
			}
		}
	}
	if node.Func != nil {
		return nil, fmt.Errorf("Asset %s not found", name)
	}
	rv := make([]string, 0, len(node.Children))
	for childName := range node.Children {
		rv = append(rv, childName)
	}
	return rv, nil
}

type bintree struct {
	Func     func() (*asset, error)
	Children map[string]*bintree
}

var _bintree = &bintree{nil, map[string]*bintree{
	"assets": {nil, map[string]*bintree{
		"apps": {nil, map[string]*bintree{
			"0000_00_cluster-version-operator_03_deployment.yaml": {assetsApps0000_00_clusterVersionOperator_03_deploymentYaml, map[string]*bintree{}},
			"0000_50_service-ca-operator_05_deploy.yaml":          {assetsApps0000_50_serviceCaOperator_05_deployYaml, map[string]*bintree{}},
			"0000_60_service-ca_05_deploy.yaml":                   {assetsApps0000_60_serviceCa_05_deployYaml, map[string]*bintree{}},
			"0000_70_dns-operator_02-deployment.yaml":             {assetsApps0000_70_dnsOperator_02DeploymentYaml, map[string]*bintree{}},
			"0000_70_dns_01-daemonset.yaml":                       {assetsApps0000_70_dns_01DaemonsetYaml, map[string]*bintree{}},
			"0000_80_openshift-router-deployment.yaml":            {assetsApps0000_80_openshiftRouterDeploymentYaml, map[string]*bintree{}},
			"000_80_hostpath-provisioner-daemonset.yaml":          {assetsApps000_80_hostpathProvisionerDaemonsetYaml, map[string]*bintree{}},
			"ovs-ds.yaml":            {assetsAppsOvsDsYaml, map[string]*bintree{}},
			"sdn-controller-ds.yaml": {assetsAppsSdnControllerDsYaml, map[string]*bintree{}},
			"sdn-ds.yaml":            {assetsAppsSdnDsYaml, map[string]*bintree{}},
		}},
	}},
}}

// RestoreAsset restores an asset under the given directory
func RestoreAsset(dir, name string) error {
	data, err := Asset(name)
	if err != nil {
		return err
	}
	info, err := AssetInfo(name)
	if err != nil {
		return err
	}
	err = os.MkdirAll(_filePath(dir, filepath.Dir(name)), os.FileMode(0755))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(_filePath(dir, name), data, info.Mode())
	if err != nil {
		return err
	}
	err = os.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
	if err != nil {
		return err
	}
	return nil
}

// RestoreAssets restores an asset under the given directory recursively
func RestoreAssets(dir, name string) error {
	children, err := AssetDir(name)
	// File
	if err != nil {
		return RestoreAsset(dir, name)
	}
	// Dir
	for _, child := range children {
		err = RestoreAssets(dir, filepath.Join(name, child))
		if err != nil {
			return err
		}
	}
	return nil
}

func _filePath(dir, name string) string {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	return filepath.Join(append([]string{dir}, strings.Split(cannonicalName, "/")...)...)
}
