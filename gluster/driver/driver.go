package driver

import (
	"fmt"

	"github.com/sapk/docker-volume-gluster/rest"
	"github.com/sapk/docker-volume-helpers/basic"
	"github.com/sapk/docker-volume-helpers/driver"

	"github.com/docker/go-plugins-helpers/volume"
	"github.com/sirupsen/logrus"
)

var (
	//MountTimeout timeout before killing a mount try in seconds
	MountTimeout = 30
	//CfgVersion current config version compat
	CfgVersion = 2
	//CfgFolder config folder
	CfgFolder = "/etc/docker-volumes/gluster/"

	gfsBase = "/mnt/"
)

//GlusterDriver docker volume plugin driver extension of basic plugin
type GlusterDriver = basic.Driver

//Init start all needed deps and serve response to API call
func Init(root string, mountUniqName bool) *GlusterDriver {
	logrus.Debugf("Init gluster driver at %s, UniqName: %v", root, mountUniqName)
	config := basic.DriverConfig{
		Version: CfgVersion,
		Root:    root,
		Folder:  CfgFolder,
		CustomOptions: map[string]interface{}{
			"mountUniqName": mountUniqName,
		},
	}
	eventHandler := basic.DriverEventHandler{
		OnMountVolume: mountVolume,
		GetMountName:  GetMountName,
	}
	return basic.Init(&config, &eventHandler)
}

func mountVolume(d *basic.Driver, v driver.Volume, m driver.Mount, r *volume.MountRequest) (*volume.MountResponse, error) {
	cmd := fmt.Sprintf("glusterfs %s %s", parseVolURI(v.GetOptions()["voluri"]), m.GetPath())
	//cmd := fmt.Sprintf("/usr/bin/mount -t glusterfs %s %s", v.VolumeURI, m.Path)
	//TODO fuseOpts   /usr/bin/mount -t glusterfs v.VolumeURI -o fuseOpts v.Mountpoint

	volName, glusterServers := getVolAndServerNames(v.GetOptions()["voluri"])
	if err := createGlusterVol(v.GetOptions()["rest"], volName, glusterServers); err != nil {
		return nil, err
	}

	if err := d.RunCmd(cmd); err != nil {
		return nil, err
	}
	return &volume.MountResponse{Mountpoint: m.GetPath()}, nil
}

func createGlusterVol(restURL string, volName string, glusterServers []string) error {
	logrus.Debugf("gluster REST API %s", restURL)
	client := rest.NewClient(restURL, gfsBase)

	logrus.Debugf("Checking if gluster volume %s exists", volName)

	exist, err := client.VolumeExist(volName)
	if err != nil {
		logrus.Warn("Unable to check if gluster volume exists, %v", err)
		return err
	}
	if !exist {
		logrus.Debugf("Creating gluster volume %s ...", volName)
		if err := client.CreateVolume(volName, glusterServers); err != nil {
			logrus.Errorf("Unable to create gluster volume, %s", err.Error())
			return err
		}
		logrus.Debugf("Gluster volume %s successfully created", volName)

	}

	return nil
}
