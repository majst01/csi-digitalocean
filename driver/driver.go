/*
Copyright 2018 DigitalOcean

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package driver

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sync"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

const (
	// DefaultDriverName defines the name that is used in Kubernetes and the CSI
	// system for the canonical, official name of this plugin
	DefaultDriverName = "lvm.csi.metal-pod.io"
)

var (
	gitTreeState = "not a git tree"
	commit       string
	version      string
)

// Driver implements the following CSI interfaces:
//
//   csi.IdentityServer
//   csi.ControllerServer
//   csi.NodeServer
//
type Driver struct {
	name string
	// publishInfoVolumeName is used to pass the volume name from
	// `ControllerPublishVolume` to `NodeStageVolume or `NodePublishVolume`
	publishInfoVolumeName string

	endpoint     string
	nodeID       string
	isController bool

	vgName         string
	devicesPattern string

	srv     *grpc.Server
	log     *logrus.Entry
	mounter Mounter
	storage *LVMStorage

	// ready defines whether the driver is ready to function. This value will
	// be used by the `Identity` service via the `Probe()` method.
	readyMu sync.Mutex // protects ready
	ready   bool
}

// NewDriver returns a CSI plugin that contains the necessary gRPC
// interfaces to interact with Kubernetes over unix domain sockets for
// managaing DigitalOcean Block Storage
func NewDriver(ep, nodeID, driverName, vgName, devicesPattern string, isController bool) (*Driver, error) {
	if driverName == "" {
		driverName = DefaultDriverName
	}

	if version == "" {
		version = "dev"
	}

	log := logrus.New().WithFields(logrus.Fields{
		"node_id": nodeID,
		"version": version,
	})

	return &Driver{
		name:                  driverName,
		publishInfoVolumeName: driverName + "/volume-name",

		endpoint: ep,
		nodeID:   nodeID,

		vgName:         vgName,
		devicesPattern: devicesPattern,
		mounter:        newMounter(log),
		storage:        &LVMStorage{},
		log:            log,
		// for now we're assuming only the controller has a non-empty token. In
		// the future we should pass an explicit flag to the driver.
		isController: isController,
	}, nil
}

// Run starts the CSI plugin by communication over the given endpoint
func (d *Driver) Run() error {
	u, err := url.Parse(d.endpoint)
	if err != nil {
		return fmt.Errorf("unable to parse address: %q", err)
	}

	addr := path.Join(u.Host, filepath.FromSlash(u.Path))
	if u.Host == "" {
		addr = filepath.FromSlash(u.Path)
	}

	// CSI plugins talk only over UNIX sockets currently
	if u.Scheme != "unix" {
		return fmt.Errorf("currently only unix domain sockets are supported, have: %s", u.Scheme)
	}
	// remove the socket if it's already there. This can happen if we
	// deploy a new version and the socket was created from the old running
	// plugin.
	d.log.WithField("socket", addr).Info("removing socket")
	if err := os.Remove(addr); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove unix domain socket file %s, error: %s", addr, err)
	}

	listener, err := net.Listen(u.Scheme, addr)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	// log response errors for better observability
	errHandler := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		resp, err := handler(ctx, req)
		if err != nil {
			d.log.WithError(err).WithField("method", info.FullMethod).Error("method failed")
		}
		return resp, err
	}

	d.srv = grpc.NewServer(grpc.UnaryInterceptor(errHandler))
	csi.RegisterIdentityServer(d.srv, d)
	csi.RegisterControllerServer(d.srv, d)
	csi.RegisterNodeServer(d.srv, d)

	d.ready = true // we're now ready to go!
	d.log.WithField("addr", addr).Info("server started")
	return d.srv.Serve(listener)
}

// Stop stops the plugin
func (d *Driver) Stop() {
	d.readyMu.Lock()
	d.ready = false
	d.readyMu.Unlock()

	d.log.Info("server stopped")
	d.srv.Stop()
}

// When building any packages that import version, pass the build/install cmd
// ldflags like so:
//   go build -ldflags "-X github.com/digitalocean/csi-digitalocean/driver.version=0.0.1"

// GetVersion returns the current release version, as inserted at build time.
func GetVersion() string {
	return version
}

// GetCommit returns the current commit hash value, as inserted at build time.
func GetCommit() string {
	return commit
}

// GetTreeState returns the current state of git tree, either "clean" or
// "dirty".
func GetTreeState() string {
	return gitTreeState
}
