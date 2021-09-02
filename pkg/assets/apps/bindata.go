// Code generated for package assets by go-bindata DO NOT EDIT. (@generated)
// sources:
// assets/apps/0000_00_flannel-daemonset.yaml
// assets/apps/0000_60_service-ca_05_deploy.yaml
// assets/apps/0000_70_dns_01-daemonset.yaml
// assets/apps/0000_80_openshift-router-deployment.yaml
// assets/apps/000_80_hostpath-provisioner-daemonset.yaml
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

var _assetsApps0000_00_flannelDaemonsetYaml = []byte(`apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: kube-flannel-ds
  namespace: kube-system
  labels:
    tier: node
    app: flannel
spec:
  selector:
    matchLabels:
      app: flannel
  template:
    metadata:
      labels:
        tier: node
        app: flannel
    spec:
      affinity:
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
            - matchExpressions:
              - key: kubernetes.io/os
                operator: In
                values:
                - linux
      hostNetwork: true
      priorityClassName: system-node-critical
      tolerations:
      - operator: Exists
        effect: NoSchedule
      serviceAccountName: flannel
      initContainers:
      - name: install-cni
        image: {{ .ReleaseImage.kube_flannel }}
        command:
        - cp
        args:
        - -f
        - /etc/kube-flannel/cni-conf.json
        - /etc/cni/net.d/10-flannel.conflist
        volumeMounts:
        - name: cni
          mountPath: /etc/cni/net.d
        - name: flannel-cfg
          mountPath: /etc/kube-flannel/
      containers:
      - name: kube-flannel
        image: {{ .ReleaseImage.kube_flannel }}
        command:
        - /opt/bin/flanneld
        args:
        - --ip-masq
        - --kube-subnet-mgr
        resources:
          requests:
            cpu: "100m"
            memory: "50Mi"
          limits:
            cpu: "100m"
            memory: "50Mi"
        securityContext:
          privileged: false
          capabilities:
            add: ["NET_ADMIN", "NET_RAW"]
        env:
        - name: POD_NAME
          valueFrom:
            fieldRef:
              fieldPath: metadata.name
        - name: POD_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        volumeMounts:
        - name: run
          mountPath: /run/flannel
        - name: flannel-cfg
          mountPath: /etc/kube-flannel/
      volumes:
      - name: run
        hostPath:
          path: /run/flannel
      - name: cni
        hostPath:
          path: /etc/cni/net.d
      - name: flannel-cfg
        configMap:
          name: kube-flannel-cfg`)

func assetsApps0000_00_flannelDaemonsetYamlBytes() ([]byte, error) {
	return _assetsApps0000_00_flannelDaemonsetYaml, nil
}

func assetsApps0000_00_flannelDaemonsetYaml() (*asset, error) {
	bytes, err := assetsApps0000_00_flannelDaemonsetYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/apps/0000_00_flannel-daemonset.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
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
    service-ca: "true"
spec:
  replicas: 1
  strategy:
    type: Recreate
  selector:
    matchLabels:
      app: service-ca
      service-ca: "true"
  template:
    metadata:
      name: service-ca
      labels:
        app: service-ca
        service-ca: "true"
    spec:
      serviceAccountName: service-ca
      containers:
      - name: service-ca-controller
        image: {{ .ReleaseImage.service_ca_operator }}
        imagePullPolicy: IfNotPresent
        command: ["service-ca-operator", "controller"]
        ports:
        - containerPort: 8443
        securityContext:
          runAsNonRoot: true
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

var _assetsApps0000_70_dns_01DaemonsetYaml = []byte(`kind: DaemonSet
apiVersion: apps/v1
metadata:
  labels:
    dns.operator.openshift.io/owning-dns: default
  name: dns-default
  namespace: openshift-dns
spec:
  template:
    spec:
      serviceAccountName: dns
      priorityClassName: system-node-critical
      containers:
      - name: dns
        image: {{ .ReleaseImage.coredns }}
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
          httpGet:
            path: /ready
            port: 8181
            scheme: HTTP
          initialDelaySeconds: 10
          periodSeconds: 3
          successThreshold: 1
          failureThreshold: 3
          timeoutSeconds: 3
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
            scheme: HTTP
          initialDelaySeconds: 60
          timeoutSeconds: 5
          successThreshold: 1
          failureThreshold: 5
        resources:
          requests:
            cpu: 50m
            memory: 70Mi
      - name: kube-rbac-proxy
        image: {{ .ReleaseImage.kube_rbac_proxy }}
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
        image: {{ .ReleaseImage.cli }}
        imagePullPolicy: IfNotPresent
        terminationMessagePolicy: FallbackToLogsOnError
        securityContext:
          privileged: true
        volumeMounts:
        - name: hosts-file
          mountPath: /etc/hosts
        env:
        - name: SERVICES
          # Comma or space separated list of services
          # NOTE: For now, ensure these are relative names; for each relative name,
          # an alias with the CLUSTER_DOMAIN suffix will also be added.
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
  template:
    metadata:
      annotations:
        "unsupported.do-not-use.openshift.io/override-liveness-grace-period-seconds": "10"
    spec:
      serviceAccountName: router
      # nodeSelector is set at runtime.
      priorityClassName: system-cluster-critical
      containers:
        - name: router
          image: {{ .ReleaseImage.haproxy_router }}
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
          livenessProbe:
            initialDelaySeconds: 10
            httpGet:
              path: /healthz
              port: 1936
          readinessProbe:
            initialDelaySeconds: 10
            httpGet:
              path: /healthz/ready
              port: 1936
          resources:
            requests:
              cpu: 100m
              memory: 256Mi
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
          defaultMode: 420
          secretName: router-certs-default
      - name: service-ca-bundle
        configMap:
          items:
          - key: service-ca.crt
            path: service-ca.crt
          name: service-ca-bundle
          optional: false
        defaultMode: 420
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
          image: {{ .ReleaseImage.kubevirt_hostpath_provisioner }}
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
            path: /var/hpvolumes
`)

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
	"assets/apps/0000_00_flannel-daemonset.yaml":             assetsApps0000_00_flannelDaemonsetYaml,
	"assets/apps/0000_60_service-ca_05_deploy.yaml":          assetsApps0000_60_serviceCa_05_deployYaml,
	"assets/apps/0000_70_dns_01-daemonset.yaml":              assetsApps0000_70_dns_01DaemonsetYaml,
	"assets/apps/0000_80_openshift-router-deployment.yaml":   assetsApps0000_80_openshiftRouterDeploymentYaml,
	"assets/apps/000_80_hostpath-provisioner-daemonset.yaml": assetsApps000_80_hostpathProvisionerDaemonsetYaml,
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
			"0000_00_flannel-daemonset.yaml":             {assetsApps0000_00_flannelDaemonsetYaml, map[string]*bintree{}},
			"0000_60_service-ca_05_deploy.yaml":          {assetsApps0000_60_serviceCa_05_deployYaml, map[string]*bintree{}},
			"0000_70_dns_01-daemonset.yaml":              {assetsApps0000_70_dns_01DaemonsetYaml, map[string]*bintree{}},
			"0000_80_openshift-router-deployment.yaml":   {assetsApps0000_80_openshiftRouterDeploymentYaml, map[string]*bintree{}},
			"000_80_hostpath-provisioner-daemonset.yaml": {assetsApps000_80_hostpathProvisionerDaemonsetYaml, map[string]*bintree{}},
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
