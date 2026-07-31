# Multi-line MicroShift YAML drop-in configs used by the router test suites.
# Imported as both Variables (module-level constants become ${VAR_NAME}) and
# Library (functions become RF keywords) by each .robot file.

# ---------------------------------------------------------------------------
# router-basic.robot
# ---------------------------------------------------------------------------

ROUTER_REMOVED = """
ingress:
  status: Removed
"""

OWNERSHIP_ALLOW = """
ingress:
  status: Managed
  routeAdmissionPolicy:
    namespaceOwnership: InterNamespaceAllowed
"""

OWNERSHIP_STRICT = """
ingress:
  status: Managed
  routeAdmissionPolicy:
    namespaceOwnership: Strict
"""

ROUTER_EXPOSE_FULL = """
ingress:
  status: Managed
  ports:
    http: 8000
    https: 8001
  listenAddress:
  - br-ex
"""

ROUTER_TUNING_CONFIG = """
ingress:
  defaultHTTPVersion: 2
  forwardedHeaderPolicy: Never
  httpEmptyRequestsPolicy: Ignore
  logEmptyRequests: Ignore
  httpCompression:
    mimeTypes:
    - "text/html"
    - "application/*"
  tuningOptions:
    headerBufferBytes: 5556
    headerBufferMaxRewriteBytes: 8000
    healthCheckInterval: 4s
    clientTimeout: 20s
    clientFinTimeout: 1.5s
    serverTimeout: 40s
    serverFinTimeout: 2s
    tunnelTimeout: 1h30m0s
    tlsInspectDelay: 6s
    threadCount: 3
    maxConnections: 60000
"""

ROUTER_SECURITY_CONFIG = """
ingress:
  certificateSecret: router-certs-custom
  routeAdmissionPolicy:
    wildcardPolicy: WildcardsAllowed
  clientTLS:
    allowedSubjectPatterns: ["route-custom.apps.example.com"]
    clientCertificatePolicy: Required
    clientCA:
      name: router-ca-certs-custom
  tlsSecurityProfile:
    type: Custom
    custom:
      Ciphers:
      - ECDHE-RSA-AES256-GCM-SHA384
      - DHE-RSA-AES256-GCM-SHA384
      - TLS_CHACHA20_POLY1305_SHA256
      MinTLSVersion: VersionTLS13
"""

ROUTER_ACCESS_LOGGING_CONFIG = """
ingress:
  accessLogging:
    status: Enabled
    destination:
      type: Container
      container:
        maxLength: 2000
    httpCaptureCookies:
    - matchType: Exact
      maxLength: 20
      name: cookie
    httpCaptureHeaders:
      request:
      - maxLength: 11
        name: header1
      response:
      - maxLength: 12
        name: header2
    httpLogFormat: some-format
  httpErrorCodePages:
    name: router-error-pages
"""

ROUTER_ACCESS_LOGGING_CONFIG_SYSLOG = """
ingress:
  accessLogging:
    status: Enabled
    destination:
      type: Syslog
      syslog:
        address: 1.2.3.4
        port: 9000
        facility: local7
        maxLength: 1000
"""

# ---------------------------------------------------------------------------
# router-config-infra.robot
# ---------------------------------------------------------------------------

LOGGING_INVALID_MAXLENGTH_NEG1 = """
ingress:
  accessLogging:
    httpCaptureCookies:
    - matchType: Exact
      maxLength: -1
      name: foo
    status: Enabled
"""

LOGGING_INVALID_MAXLENGTH_ZERO = """
ingress:
  accessLogging:
    httpCaptureCookies:
    - matchType: Exact
      maxLength: 0
      name: foo
    status: Enabled
"""

LOGGING_INVALID_COOKIE_NAME = """
ingress:
  accessLogging:
    httpCaptureCookies:
    - matchType: Exact
      maxLength: 100
      name: "foo 33#?-"
    status: Enabled
"""

LOGGING_INVALID_HEADER_MAXLENGTH = """
ingress:
  accessLogging:
    httpCaptureHeaders:
      request:
      - maxLength: -1
        name: Host
      response:
      - maxLength: 10
        name: "Server"
    status: Enabled
"""

LOGGING_INVALID_STATUS = """
ingress:
  accessLogging:
    httpCaptureHeaders:
      request:
      - maxLength: 10
        name: Host
    status: Enable
"""

LOGGING_COOKIES_NO_STATUS = """
ingress:
  accessLogging:
    httpCaptureCookies:
    - matchType: Prefix
      maxLength: 100
      namePrefix: foo
"""

CONFIG_ROUTER_REMOVED = """
ingress:
  status: Removed
"""

CONFIG_ROUTER_MANAGED = """
ingress:
  status: Managed
"""


def config_custom_listen(iface, http_port, https_port):
    return f"""
ingress:
  listenAddress:
  - {iface}
  ports:
    http: {http_port}
    https: {https_port}
"""


def config_syslog(address, facility="local1"):
    return f"""
ingress:
  accessLogging:
    destination:
      syslog:
        address: {address}
        port: 514
        facility: {facility}
      type: Syslog
    status: Enabled
"""


# ---------------------------------------------------------------------------
# router-config-logging.robot
# ---------------------------------------------------------------------------

CONFIG_COOKIE_EXACT_100 = """
ingress:
  accessLogging:
    httpCaptureCookies:
    - matchType: Exact
      maxLength: 100
      name: foo
    status: Enabled
"""

CONFIG_COOKIE_EXACT_10 = """
ingress:
  accessLogging:
    httpCaptureCookies:
    - matchType: Exact
      maxLength: 10
      name: foo
    status: Enabled
"""

CONFIG_COOKIE_PREFIX = """
ingress:
  accessLogging:
    httpCaptureCookies:
    - matchType: Prefix
      maxLength: 100
      namePrefix: foo
    status: Enabled
"""

CONFIG_CAPTURE_HEADERS_120 = """
ingress:
  accessLogging:
    httpCaptureHeaders:
      request:
      - maxLength: 120
        name: Host
      response:
      - maxLength: 120
        name: "Server"
    status: Enabled
"""

CONFIG_CAPTURE_HEADERS_MAXLEN = """
ingress:
  accessLogging:
    httpCaptureHeaders:
      request:
      - maxLength: 16
        name: Host
      response:
      - maxLength: 5
        name: "Server"
    status: Enabled
"""

CONFIG_ERROR_PAGES = """
ingress:
  httpErrorCodePages:
    name: custom-82004-error-code-pages
"""

# ---------------------------------------------------------------------------
# router-config-policies.robot
# ---------------------------------------------------------------------------

CONFIG_TUNING_CUSTOM = """
ingress:
  forwardedHeaderPolicy: "Replace"
  httpCompression:
    mimeTypes:
    - "image"
  logEmptyRequests: "Ignore"
  tuningOptions:
    clientFinTimeout: "2s"
    clientTimeout: "60s"
    headerBufferBytes: 65536
    headerBufferMaxRewriteBytes: 16384
    healthCheckInterval: "10s"
    maxConnections: 100000
    serverFinTimeout: "2s"
    serverTimeout: "60s"
    threadCount: 8
    tlsInspectDelay: "10s"
    tunnelTimeout: "2h\"
"""

CONFIG_MTLS18_SUBJECT_FILTER = """
ingress:
  clientTLS:
    allowedSubjectPatterns: ["/CN=example-test.com"]
    clientCA:
      name: "ocp80518"
    clientCertificatePolicy: "Required\"
"""

CONFIG_WILDCARD_ALLOWED = """
ingress:
  routeAdmissionPolicy:
    wildcardPolicy: "WildcardsAllowed\"
"""

CONFIG_WILDCARD_DISALLOWED = """
ingress:
  routeAdmissionPolicy:
    wildcardPolicy: "WildcardsDisallowed\"
"""

CONFIG_LOG_FORMAT_1 = """
ingress:
  accessLogging:
    httpLogFormat: "%{+Q}r"
    status: Enabled
"""

CONFIG_LOG_FORMAT_2 = """
ingress:
  accessLogging:
    httpLogFormat: "%ci:%cp %si:%sp %HU %ST"
    status: Enabled
"""

# ---------------------------------------------------------------------------
# router-config-tls.robot
# ---------------------------------------------------------------------------

CONFIG_OLD_TLS = """
ingress:
  tlsSecurityProfile:
    old: {}
    type: Old
"""

CONFIG_INTERMEDIATE_TLS = """
ingress:
  tlsSecurityProfile:
    intermediate: {}
    type: Intermediate
"""

CONFIG_MODERN_TLS = """
ingress:
  tlsSecurityProfile:
    modern: {}
    type: Modern
"""

CONFIG_CUSTOM_TLS = """
ingress:
  tlsSecurityProfile:
    custom:
      ciphers:
      - DHE-RSA-AES256-GCM-SHA384
      - ECDHE-ECDSA-AES256-GCM-SHA384
      minTLSVersion: VersionTLS12
    type: Custom
"""

CONFIG_MTLS17_REQUIRED = """
ingress:
  clientTLS:
    clientCA:
      name: "ocp80517"
    clientCertificatePolicy: "Required\"
"""

CONFIG_MTLS17_OPTIONAL = """
ingress:
  clientTLS:
    clientCA:
      name: "ocp80517"
    clientCertificatePolicy: "Optional\"
"""

CONFIG_CUSTOM_CERT = """
ingress:
  certificateSecret: "router-test-cert\"
"""


def san_extension(base_domain):
    return f"""\
[ v3_req ]
subjectAltName = @alt_names
[ alt_names ]
DNS.1 = *.{base_domain}
"""
