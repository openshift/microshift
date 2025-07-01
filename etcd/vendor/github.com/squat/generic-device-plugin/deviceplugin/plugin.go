// Copyright 2020 the generic-device-plugin authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package deviceplugin

import (
	"context"
	"encoding/base64"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/oklog/run"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

const (
	socketPrefix        = "gdp"
	socketCheckInterval = 1 * time.Second
	restartInterval     = 5 * time.Second
)

// Plugin is a Kubernetes device plugin that can be run.
type Plugin interface {
	v1beta1.DevicePluginServer
	Run(context.Context) error
}

// plugin is a Kubernetes device plugin.
// It handles the registration and lifecycle
// of the device plugin server.
type plugin struct {
	v1beta1.DevicePluginServer
	resource   string
	pluginDir  string
	socket     string
	grpcServer *grpc.Server
	logger     log.Logger

	// metrics
	restartsTotal prometheus.Counter
}

// NewPlugin creates a new instance of a device plugin.
func NewPlugin(resource, pluginDir string, dps v1beta1.DevicePluginServer, logger log.Logger, reg prometheus.Registerer) Plugin {
	if logger == nil {
		logger = log.NewNopLogger()
	}
	p := &plugin{
		DevicePluginServer: dps,
		resource:           resource,
		pluginDir:          pluginDir,
		socket:             filepath.Join(pluginDir, fmt.Sprintf("%s-%s-%d.sock", socketPrefix, base64.StdEncoding.EncodeToString([]byte(resource)), time.Now().Unix())),
		logger:             logger,
		restartsTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "device_plugin_restarts_total",
			Help: "The number of times that the device plugin has restarted.",
		}),
	}

	if reg != nil {
		reg.MustRegister(p.restartsTotal)
	}

	return p
}

// Run runs the device plugin until the given context is cancelled.
func (p *plugin) Run(ctx context.Context) error {
Outer:
	for {
		select {
		case <-ctx.Done():
			break Outer
		default:
			if err := p.runOnce(ctx); err != nil {
				level.Warn(p.logger).Log("msg", "encountered error while running plugin; trying again in 5 seconds", "err", err)
				select {
				case <-ctx.Done():
					break Outer
				case <-time.After(restartInterval):
					p.restartsTotal.Inc()
				}
			}
		}
	}
	return p.cleanUp()
}

// serve starts the gRPC server and waits for it to be running
// and accepting connections before returning. It returns a function
// to wait for its completion as well as another to interrupt it.
// This makes it convenient to run in a run.Group.
func (p *plugin) serve(ctx context.Context) (func() error, func(error), error) {
	// Run the gRPC server.
	level.Info(p.logger).Log("msg", "listening on Unix socket", "socket", p.socket)
	l, err := net.Listen("unix", p.socket)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to listen on Unix socket %q: %v", p.socket, err)
	}

	ch := make(chan error)
	go func() {
		level.Info(p.logger).Log("msg", "starting gRPC server")
		ch <- p.grpcServer.Serve(l)
		close(ch)
	}()
	t := time.NewTimer(1 * time.Second)
	defer t.Stop()
Outer:
	for ctx.Err() == nil {
		for range p.grpcServer.GetServiceInfo() {
			break Outer
		}
		level.Info(p.logger).Log("msg", "waiting for gRPC server to be ready")
		select {
		case <-ctx.Done():
			return nil, nil, ctx.Err()
		case <-t.C:
			t.Reset(1 * time.Second)
		}
	}
	return func() error {
			return <-ch
		},
		func(_ error) {
			p.grpcServer.Stop()
			// Drain the channel to clean up.
			<-ch
			if err := l.Close(); err != nil {

				level.Warn(p.logger).Log("msg", "encountered error while closing the listener", "err", err)
			}
		}, nil
}

// runOnce runs the plugin one time until an error is encountered,
// until the socket is removed, or until the context is cancelled.
func (p *plugin) runOnce(ctx context.Context) error {
	p.grpcServer = grpc.NewServer()
	v1beta1.RegisterDevicePluginServer(p.grpcServer, p.DevicePluginServer)

	var g run.Group
	{
		// Run the gRPC server.
		execute, interrupt, err := p.serve(ctx)
		if err != nil {
			return fmt.Errorf("failed to start gRPC server: %v", err)
		}
		g.Add(execute, interrupt)
	}

	{
		// Register the plugin with the kubelet.
		ctx, cancel := context.WithCancel(ctx)
		g.Add(func() error {
			defer cancel()
			level.Info(p.logger).Log("msg", "waiting for the gRPC server to be ready")
			c, err := grpc.DialContext(ctx, p.socket, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock(),
				grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
					return (&net.Dialer{}).DialContext(ctx, "unix", addr)
				}),
			)
			if err != nil {
				return fmt.Errorf("failed to create connection to local gRPC server: %v", err)
			}
			if err := c.Close(); err != nil {
				return fmt.Errorf("failed to close connection to local gRPC server: %v", err)
			}
			level.Info(p.logger).Log("msg", "the gRPC server is ready")
			if err := p.registerWithKubelet(); err != nil {
				return fmt.Errorf("failed to register with kubelet: %v", err)
			}
			<-ctx.Done()
			return nil
		}, func(error) {
			cancel()
		})
	}

	{
		// Watch the socket.
		t := time.NewTicker(socketCheckInterval)
		ctx, cancel := context.WithCancel(ctx)
		defer t.Stop()
		g.Add(func() error {
			for {
				select {
				case <-t.C:
					if _, err := os.Lstat(p.socket); err != nil {
						return fmt.Errorf("failed to stat plugin socket %q: %v", p.socket, err)
					}
				case <-ctx.Done():
					return nil
				}

			}
		}, func(error) {
			cancel()
		})

	}

	return g.Run()
}

func (p *plugin) registerWithKubelet() error {
	level.Info(p.logger).Log("msg", "registering plugin with kubelet")
	conn, err := grpc.Dial(filepath.Join(p.pluginDir, filepath.Base(v1beta1.KubeletSocket)), grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
			d := &net.Dialer{}
			return d.DialContext(ctx, "unix", addr)
		}))
	if err != nil {
		return fmt.Errorf("failed to connect to kubelet: %v", err)
	}
	defer conn.Close()

	client := v1beta1.NewRegistrationClient(conn)
	request := &v1beta1.RegisterRequest{
		Version:      v1beta1.Version,
		Endpoint:     filepath.Base(p.socket),
		ResourceName: p.resource,
	}
	if _, err = client.Register(context.Background(), request); err != nil {
		return fmt.Errorf("failed to register plugin with kubelet service: %v", err)
	}
	return nil
}

func (p *plugin) cleanUp() error {
	if err := os.Remove(p.socket); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove socket: %v", err)
	}
	return nil
}
