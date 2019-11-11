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

package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/metal-pod/csi-lvm/driver"
)

func main() {
	var (
		endpoint     = flag.String("endpoint", "unix:///var/lib/kubelet/plugins/"+driver.DefaultDriverName+"/csi.sock", "CSI endpoint")
		nodeid       = flag.String("nodeid", "unknown-node", "the name of this node")
		vgName       = flag.String("vgname", "csi-lvm", "the name of the volume group")
		devicesGrok  = flag.String("devices-pattern", "", "the grok pattern to decide which devices to use")
		driverName   = flag.String("driver-name", driver.DefaultDriverName, "Name for the driver")
		isController = flag.Bool("controller", false, "act as csi controller")
		version      = flag.Bool("version", false, "Print the version and exit.")
	)
	flag.Parse()

	if *version {
		fmt.Printf("%s - %s (%s)\n", driver.GetVersion(), driver.GetCommit(), driver.GetTreeState())
		os.Exit(0)
	}

	drv, err := driver.NewDriver(*endpoint, *nodeid, *driverName, *vgName, *devicesGrok, *isController)
	if err != nil {
		log.Fatalln(err)
	}

	if err := drv.Run(); err != nil {
		log.Fatalln(err)
	}
}
