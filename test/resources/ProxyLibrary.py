from http.server import HTTPServer, BaseHTTPRequestHandler
import socket
import select
import threading
from robot.libraries.BuiltIn import BuiltIn


class ProxyHandler(BaseHTTPRequestHandler):
    def log_message(self, format, *args):
        pass

    def do_CONNECT(self):
        """Handle tunneling via CONNECT"""
        target_host, target_port = self.path.split(":")
        target_port = int(target_port)

        try:
            with socket.create_connection((target_host, target_port)) as upstream:
                self.send_response(200, "Connection Established")
                self.end_headers()
                self._tunnel_data(self.connection, upstream)
        except Exception as e:
            self.send_error(502, f"Proxy Error: {e}")

    def do_GET(self):
        self._proxy_request()

    def do_POST(self):
        self._proxy_request()

    def do_PUT(self):
        self._proxy_request()

    def do_DELETE(self):
        self._proxy_request()

    def do_PATCH(self):
        self._proxy_request()

    def do_OPTIONS(self):
        self._proxy_request()

    def do_HEAD(self):
        self._proxy_request()

    def do_TRACE(self):
        self._proxy_request()

    def _proxy_request(self):
        target_url = self.path
        target_host, target_port = target_url.split(":")
        target_port = int(target_port)

        with socket.create_connection((target_host, target_port)) as upstream:
            upstream.sendall(f"{self.command} {target_url} HTTP/1.1\r\n".encode())
            for key, value in self.headers.items():
                upstream.sendall(f"{key}: {value}\r\n".encode())
            upstream.sendall(b"\r\n")
            self._tunnel_data(upstream, self.connection)

    def _tunnel_data(self, src, dst):
        sockets = [src, dst]
        while True:
            r, _, _ = select.select(sockets, [], [])
            for s in r:
                data = s.recv(4096)
                if not data:
                    return
                (dst if s is src else src).sendall(data)

class ProxyServer:
    def __init__(self, host="0.0.0.0", port=8080):
        self.server = HTTPServer((host, port), ProxyHandler)
        self.thread = None

    def start(self):
        """Start the proxy server in a separate thread."""
        self.thread = threading.Thread(target=self.server.serve_forever, daemon=True)
        self.thread.start()
        print(f"Proxy server running on port {self.server.server_address[1]}")

    def stop(self):
        """Stop the proxy server."""
        self.server.shutdown()
        self.server.server_close()
        print("Proxy server stopped")

class ProxyLibrary:
    def __init__(self):
        self.proxy = None

    def start_proxy_server(self, host="0.0.0.0", port=8080):
        """Start the proxy server."""
        if self.proxy is None:
            self.proxy = ProxyServer(host, port)
            self.proxy.start()
        else:
            BuiltIn().fail("Proxy server was already running.")

    def stop_proxy_server(self):
        """Stop the proxy server."""
        if self.proxy is None:
            BuiltIn().fail("Proxy server was not running.")
        else:
            self.proxy.stop()
