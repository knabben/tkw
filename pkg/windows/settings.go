package windows

import (
	"encoding/json"
	"fmt"
	"github.com/knabben/tkw/api/v1alpha1"
	"github.com/knabben/tkw/pkg/config"
	"github.com/knabben/tkw/pkg/vsphere"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type WindowsSettings struct {
	OSImagePath          string
	VMToolsPath          string
	ServiceName          string
	ServiceNamespace string
	ServicePort          int32
	WindowsConfiguration *WindowsConfiguration
}

func NewWindowsSettings(osp, vmp, svcName, svcNS string, svcPort int32, img *v1alpha1.OSImage) *WindowsSettings {
	return &WindowsSettings{
		OSImagePath: osp,
		VMToolsPath: vmp,
		ServiceName: svcName,
		ServicePort: svcPort,
		ServiceNamespace: svcNS,
		WindowsConfiguration: &WindowsConfiguration{
			Folder:       img.Spec.VSphereFolder,
			Datastore:    img.Spec.VSphereDataStore,
			Network:      img.Spec.VSphereNetwork,
			ResourcePool: img.Spec.VSphereResourcePool,
			Cluster:      img.Spec.VSphereCluster,
		},
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

	w.WindowsConfiguration.OsIsoPath = generateIsoPath(w.WindowsConfiguration.Datastore, filepath.Base(w.OSImagePath))
	w.WindowsConfiguration.VmtoolsIsoPath = generateIsoPath(w.WindowsConfiguration.Datastore, filepath.Base(w.VMToolsPath))

	w.WindowsConfiguration.Password = mapper.Get(vsphere.VspherePassword)
	w.WindowsConfiguration.Username = mapper.Get(vsphere.VsphereUsername)
	w.WindowsConfiguration.VcenterServer = mapper.Get(vsphere.VsphereServer)
	w.WindowsConfiguration.Datacenter = mapper.Get(vsphere.VsphereDataCenter)

	w.WindowsConfiguration.Runtime = "containerd"
	w.WindowsConfiguration.ConvertToTemplate = "true"

	// todo(knabben): pass it to paremeters
	kubernetesVersion := "v1.23.8" // todo(knabben) - fix this
	w.WindowsConfiguration.WindowsUpdatesCategories = "CriticalUpdates SecurityUpdates UpdateRollups"
	w.WindowsConfiguration.UnattendTimezone = "GMT Standard Time" // todo(knabben) pass to parameter
	w.WindowsConfiguration.KubernetesSemver = kubernetesVersion

	const (
		containerdFile = "cri-containerd-v1.6.6+vmware.2.windows-amd64.tar"
		containerdHash = "a5348e2e7cc63194c2bb4575dd3c414a26c829380e72a81c3dc2d12454f67fcd"
		antreaFile     = "antrea-windows-advanced.zip"
	)

	w.WindowsConfiguration.ContainerdURL = fmt.Sprintf("%s/files/containerd/%s", baseUrl, containerdFile)
	w.WindowsConfiguration.ContainerdSha256Windows = containerdHash

	w.WindowsConfiguration.InsecureConnection = "true"
	w.WindowsConfiguration.LinkedClone = "false"
	w.WindowsConfiguration.DisableHypervisor = "false"
	w.WindowsConfiguration.CreateSnapshot = "false"
	w.WindowsConfiguration.NetbiosHostNameCompatibility = "false"
	w.WindowsConfiguration.PauseImage = "mcr.microsoft.com/oss/kubernetes/pause:3.6"
	w.WindowsConfiguration.Prepull = "false"
	w.WindowsConfiguration.AdditionalExecutables = "true"

	w.WindowsConfiguration.KubernetesBaseURL = fmt.Sprintf("%s/files/kubernetes/", baseUrl)
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
	return fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", w.ServiceName, w.ServiceNamespace, w.ServicePort)
}
