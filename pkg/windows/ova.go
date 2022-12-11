package windows

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
	AdditionalDowloadFiles               string `json:"additional_download_files"`
	AdditionalExecutables                string `json:"additional_executables"`
	AdditionalExecutablesDestinationPath string `json:"additional_executables_destination_path"`
	AdditionalExecutablesList            string `json:"additional_executables_list"`
	LoadAdditionalComponents             string `json:"load_additional_components"`
}
