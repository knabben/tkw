package windows

import (
	"encoding/json"
	"fmt"
	"strings"
	"tkw/pkg/config"
	"tkw/pkg/vsphere"
)

// WindowsConfiguration holds image-builder configuration parameters
type WindowsConfiguration struct {
	UnattendTimezone                     string `json:"unattend_timezone"`
	WindowsUpdatesCategories             string `json:"windows_updates_categories"`
	WindowsUpdatesKbs                    string `json:"windows_updates_kbs"`
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

// generateISOPath returns the full path for file access
// ie [datastore1] iso/VMware-tools-windows-11.3.5-18557794.iso)
func generateIsoPath(datastore, path string) string {
	ds := strings.Split(datastore, "/")
	return fmt.Sprintf("[%s] ./%s", ds[len(ds)-1], path)
}

// PopulateWindowsConfiguration renders the full Windows settings in JSON
func (w *WindowsConfiguration) PopulateWindowsConfiguration(mapper *config.Mapper, isoPath, vmToolPath string) ([]byte, error) {
	kubernetesVersion := "v1.23.8" // todo(knabben) - fix this
	datastore := mapper.Get(vsphere.VsphereDataStore)

	w.UnattendTimezone = "GMT Standard Time"
	w.WindowsUpdatesCategories = "CriticalUpdates SecurityUpdates UpdateRollups"
	w.KubernetesSemver = kubernetesVersion
	w.Runtime = "containerd"
	w.ConvertToTemplate = "true"
	w.Folder = mapper.Get(vsphere.VsphereFolder)
	w.Password = mapper.Get(vsphere.VspherePassword)
	w.Username = mapper.Get(vsphere.VsphereUsername)
	w.Datastore = mapper.Get(vsphere.VsphereDataStore)
	w.Datacenter = mapper.Get(vsphere.VsphereDataCenter)
	w.VmtoolsIsoPath = generateIsoPath(datastore, isoPath)
	w.OsIsoPath = generateIsoPath(datastore, vmToolPath)
	w.InsecureConnection = mapper.Get(vsphere.VsphereInsecure)
	w.DisableHypervisor = "false"
	w.Network = mapper.Get(vsphere.VsphereNetwork)
	w.LinkedClone = "false"
	w.ResourcePool = mapper.Get(vsphere.VsphereResourcePool)
	w.VcenterServer = mapper.Get(vsphere.VsphereServer)
	w.CreateSnapshot = "false"
	w.NetbiosHostNameCompatibility = "false"
	w.KubernetesBaseURL = ""       // base URL from kubeconfig burrito service
	w.ContainerdURL = ""           // burrito
	w.ContainerdSha256Windows = "" // sha256 -  "a5348e2e7cc63194c2bb4575dd3c414a26c829380e72a81c3dc2d12454f67fcd"
	w.PauseImage = "mcr.microsoft.com/oss/kubernetes/pause:3.6"
	w.Prepull = "false"
	w.AdditionalExecutables = "true"
	w.AdditionalExecutablesDestinationPath = "c:/k/antrea/"
	w.AdditionalExecutablesList = "http://<controlplane>:30008/files/antrea-windows/antrea-windows-advanced.zip" // burrito
	w.LoadAdditionalComponents = "true"

	// todo(knabben) - save the output in a file.
	return json.Marshal(w)
}
