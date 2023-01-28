package windows

import (
	"encoding/json"
	"fmt"
	"github.com/knabben/tkw/api/v1alpha1"
	"github.com/knabben/tkw/pkg/config"
	"github.com/knabben/tkw/pkg/vsphere"
	"path/filepath"
	"strings"
)

// WindowsConfiguration holds image-builder configuration parameters
type WindowsConfiguration struct {
	UnattendTimezone                     string `json:"unattend_timezone"`
	WindowsUpdatesCategories             string `json:"windows_updates_categories"`
	WindowUpdatesKbs                     string `json:"windows_updates_kbs"`
	KubernetesSemver                     string `json:"kubernetes_semver"`
	Cluster                              string `json:"cluster"`
	Template                             string `json:"template"`
	Password                             string `json:"password"`
	Folder                               string `json:"folder"`
	Runtime                              string `json:"runtime"`
	Username                             string `json:"username"`
	Datastore                            string `json:"datastore"`
	Datacenter                           string `json:"datacenter"`
	ConvertToTemplate                    string `json:"convert_to_template"`
	VmtoolsIsoPath                       string `json:"vmtools_iso_path"`
	InsecureConnection                   string `json:"insecure_connection"`
	DisableHypervisor                    string `json:"disable_hypervisor"`
	Network                              string `json:"network"`
	LinkedClone                          string `json:"linked_clone"`
	OsIsoPath                            string `json:"os_iso_path"`
	ResourcePool                         string `json:"resource_pool"`
	VcenterServer                        string `json:"vcenter_server"`
	CreateSnapshot                       string `json:"create_snapshot"`
	NetbiosHostNameCompatibility         string `json:"netbios_host_name_compatibility"`
	KubernetesBaseURL                    string `json:"kubernetes_base_url"`
	ContainerdURL                        string `json:"containerd_url"`
	ContainerdSha256Windows              string `json:"containerd_sha256_windows"`
	PauseImage                           string `json:"pause_image"`
	Prepull                              string `json:"prepull"`
	AdditionalPrepullImages              string `json:"additional_prepull_images"`
	AdditionalDownloadFiles              string `json:"additional_download_files"`
	AdditionalExecutables                string `json:"additional_executables"`
	AdditionalExecutablesDestinationPath string `json:"additional_executables_destination_path"`
	AdditionalExecutablesList            string `json:"additional_executables_list"`
	LoadAdditionalComponents             string `json:"load_additional_components"`
}

type WindowsSettings struct {
	OSImagePath          string
	VMToolsPath          string
	ServiceName          string
	ServiceNamespace     string
	ServicePort          int32
	WindowsConfiguration *WindowsConfiguration
}

func NewWindowsSettings(osp, vmp, svcName, svcNS string, svcPort int32, img *v1alpha1.OSImage) *WindowsSettings {
	return &WindowsSettings{
		OSImagePath:      osp,
		VMToolsPath:      vmp,
		ServiceName:      svcName,
		ServicePort:      svcPort,
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

	// todo(knabben): pass it via parameters on spec
	kubernetesVersion := "v1.23.8"
	w.WindowsConfiguration.WindowsUpdatesCategories = "CriticalUpdates SecurityUpdates UpdateRollups"
	w.WindowsConfiguration.UnattendTimezone = "GMT Standard Time"
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

// BaseBurritoURL returns the service endpoint for assets download
func (w *WindowsSettings) BaseBurritoURL() string {
	return fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", w.ServiceName, w.ServiceNamespace, w.ServicePort)
}

// generateISOPath returns the full path for file access
// ie [datastore1] iso/VMw.WindowsConfiguration.re-tools-windows-11.3.5-18557794.iso)
func generateIsoPath(datastore, path string) string {
	ds := strings.Split(datastore, "/")
	return fmt.Sprintf("[%s] ./%s", ds[len(ds)-1], path)
}
