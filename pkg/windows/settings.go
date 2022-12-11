package windows

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"tkw/pkg/config"
	"tkw/pkg/vsphere"
)

type WindowsSettings struct {
	OSImagePath string
	VMToolsPath string
	NodeIP      string

	WindowsConfiguration *WindowsConfiguration
}

func NewWindowsSettings(osp, vmp, nodeip string) *WindowsSettings {
	return &WindowsSettings{
		OSImagePath:          osp,
		VMToolsPath:          vmp,
		NodeIP:               nodeip,
		WindowsConfiguration: &WindowsConfiguration{},
	}
}

// SaveTempJSON dumps the marshal JSON on a temp file and returns the path
func (w *WindowsSettings) SaveTempJSON(content []byte) (string, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	// create and open a temporary file
	f, err := os.CreateTemp(pwd, ".tmpfile-") // in Go version older than 1.17 you can use ioutil.TempFile
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	if _, err := f.Write(content); err != nil {
		log.Fatal(err)
	}
	return f.Name(), nil
}

// GenerateJSONConfig renders the full Window.WindowsConfiguration. settings in JSON
func (w *WindowsSettings) GenerateJSONConfig(mapper *config.Mapper) ([]byte, error) {
	baseUrl := w.BaseBurritoURL()
	kubernetesVersion := "v1.23.8" // todo(knabben) - fix this
	datastore := mapper.Get(vsphere.VsphereDataStore)

	w.WindowsConfiguration.Cluster = "cluster0"                   // todo(knabben) - detect this info
	w.WindowsConfiguration.UnattendTimezone = "GMT Standard Time" // todo(knabben - allow user change
	w.WindowsConfiguration.WindowsUpdatesCategories = "CriticalUpdates SecurityUpdates UpdateRollups"
	w.WindowsConfiguration.KubernetesSemver = kubernetesVersion
	w.WindowsConfiguration.Runtime = "containerd"
	w.WindowsConfiguration.ConvertToTemplate = "true"
	w.WindowsConfiguration.Folder = mapper.Get(vsphere.VsphereFolder)
	w.WindowsConfiguration.Password = mapper.Get(vsphere.VspherePassword)
	w.WindowsConfiguration.Username = mapper.Get(vsphere.VsphereUsername)
	w.WindowsConfiguration.Datastore = mapper.Get(vsphere.VsphereDataStore)
	w.WindowsConfiguration.Datacenter = mapper.Get(vsphere.VsphereDataCenter)
	w.WindowsConfiguration.Network = mapper.Get(vsphere.VsphereNetwork)
	w.WindowsConfiguration.ResourcePool = mapper.Get(vsphere.VsphereResourcePool)
	w.WindowsConfiguration.VcenterServer = mapper.Get(vsphere.VsphereServer)
	w.WindowsConfiguration.OsIsoPath = generateIsoPath(datastore, filepath.Base(w.OSImagePath))
	w.WindowsConfiguration.VmtoolsIsoPath = generateIsoPath(datastore, filepath.Base(w.VMToolsPath))
	w.WindowsConfiguration.InsecureConnection = "true"
	w.WindowsConfiguration.LinkedClone = "false"
	w.WindowsConfiguration.DisableHypervisor = "false"
	w.WindowsConfiguration.CreateSnapshot = "false"
	w.WindowsConfiguration.NetbiosHostNameCompatibility = "false"
	w.WindowsConfiguration.PauseImage = "mcr.microsoft.com/oss/kubernetes/pause:3.6"
	w.WindowsConfiguration.Prepull = "false"
	w.WindowsConfiguration.AdditionalExecutables = "true"

	const (
		containerdFile = "cri-containerd-v1.6.6+vmware.2.windows-amd64.tar"
		containerdHash = "a5348e2e7cc63194c2bb4575dd3c414a26c829380e72a81c3dc2d12454f67fcd"
		antreaFile     = "antrea-windows-advanced.zip"
	)

	w.WindowsConfiguration.KubernetesBaseURL = fmt.Sprintf("%s/files/kubernetes/", baseUrl)
	w.WindowsConfiguration.ContainerdURL = fmt.Sprintf("%s/files/containerd/%s", baseUrl, containerdFile)
	w.WindowsConfiguration.ContainerdSha256Windows = containerdHash
	w.WindowsConfiguration.LoadAdditionalComponents = "true"
	w.WindowsConfiguration.AdditionalExecutablesDestinationPath = "c:/k/antrea/"
	w.WindowsConfiguration.AdditionalExecutablesList = fmt.Sprintf("%s/files/antrea-windows/%s", baseUrl, antreaFile)
	return json.Marshal(w.WindowsConfiguration)
}

// generateISOPath returns the full path for file access
// ie [datastore1] iso/VMw.WindowsConfiguration.re-tools-windows-11.3.5-18557794.iso)
func generateIsoPath(datastore, path string) string {
	ds := strings.Split(datastore, "/")
	return fmt.Sprintf("[%s] ./%s", ds[len(ds)-1], path)
}

// BaseBurritoURL returns the service endpoint for assets download
func (w *WindowsSettings) BaseBurritoURL() string {
	return fmt.Sprintf("http://%s:%d", w.NodeIP, NodePort)
}
