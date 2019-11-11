package driver

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/google/lvmd/commands"
)

const (
	linearType  = "linear"
	stripedType = "striped"
	mirrorType  = "mirror"
)

type LVMStorage struct {
}

func (s *LVMStorage) createVG(name string, devicesPattern string) (string, error) {
	vgs, err := commands.ListVG(context.Background())
	if err != nil {
		log.Printf("unable to list existing volumegroups:%v", err)
	}
	vgexists := false
	for _, vg := range vgs {
		log.Printf("compare vg:%s with:%s\n", vg.Name, name)
		if vg.Name == name {
			vgexists = true
			break
		}
	}
	if vgexists {
		log.Printf("volumegroup: %s already exists\n", name)
		return name, nil
	}
	physicalVolumes, err := s.devices(devicesPattern)
	if err != nil {
		return "", fmt.Errorf("unable to lookup devices from devicesPattern %s, err:%v", devicesPattern, err)
	}
	tags := []string{"vg.metal-pod.io/csi-lvm"}

	args := []string{"-v", name}
	args = append(args, physicalVolumes...)
	for _, tag := range tags {
		args = append(args, "--add-tag", tag)
	}
	log.Printf("create vg with command: vgcreate %v", args)
	cmd := exec.Command("vgcreate", args...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// createLV creates a new volume
func (s *LVMStorage) createLVS(ctx context.Context, vg string, name string, size uint64, lvmType string) (string, error) {
	lvs, err := commands.ListLV(context.Background(), vg+"/"+name)
	if err != nil {
		log.Printf("unable to list existing logicalvolumes:%v", err)
	}
	lvExists := false
	for _, lv := range lvs {
		log.Printf("compare lv:%s with:%s\n", lv.Name, name)
		if strings.Contains(lv.Name, name) {
			lvExists = true
			break
		}
	}

	if lvExists {
		log.Printf("logicalvolume: %s already exists\n", name)
		return name, nil
	}

	if size == 0 {
		return "", fmt.Errorf("size must be greater than 0")
	}

	args := []string{"-v", "-n", name, "-W", "y", "-L", fmt.Sprintf("%db", size)}

	pvs, err := s.pvCount(vg)
	if err != nil {
		return "", fmt.Errorf("unable to determine pv count of vg: %v", err)
	}
	switch lvmType {
	case stripedType:
		if pvs < 2 {
			return "", fmt.Errorf("cannot use type %s when pv count is smaller than 2", lvmType)
		}
		args = append(args, "--type", "striped", "--stripes", fmt.Sprintf("%d", pvs))
	case mirrorType:
		if pvs < 2 {
			return "", fmt.Errorf("cannot use type %s when pv count is smaller than 2", lvmType)
		}
		args = append(args, "--type", "raid1", "--mirrors", "1", "--nosync")
	case linearType:
	default:
		return "", fmt.Errorf("unsupported lvmtype: %s", lvmType)
	}

	tags := []string{"lv.metal-pod.io/csi-lvm"}
	for _, tag := range tags {
		args = append(args, "--add-tag", tag)
	}
	args = append(args, vg)
	log.Printf("lvreate %s", args)
	cmd := exec.Command("lvcreate", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("lvcreate output:%s error:%v ", out, err)
	}
	log.Printf("lv created:%s", out)
	return name, nil
}

func (s *LVMStorage) pvCount(vgname string) (int, error) {
	cmd := exec.Command("vgs", vgname, "--noheadings", "-o", "pv_count")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return 0, err
	}
	outStr := strings.TrimSpace(string(out))
	count, err := strconv.Atoi(outStr)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (s *LVMStorage) devices(devicesPattern string) (devices []string, err error) {
	log.Printf("search devices :%s ", devicesPattern)
	matches, err := filepath.Glob(devicesPattern)
	if err != nil {
		return nil, err
	}
	log.Printf("found: %s", matches)
	devices = append(devices, matches...)

	return devices, nil
}
