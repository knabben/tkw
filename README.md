# <img src="https://upload.wikimedia.org/wikipedia/commons/8/87/Windows_logo_-_2021.svg" data-canonical-src="https://upload.wikimedia.org/wikipedia/commons/8/87/Windows_logo_-_2021.svg" width="3%"/>   <img src="https://avatars.githubusercontent.com/u/54452117?s=200&v=4" data-canonical-src="https://avatars.githubusercontent.com/u/54452117?s=200&v=4" width="3%"/> Tanzu Kubernetes Grid Windows Toolkit

This toolkit is a bunch of helpers functions to extract data or automatize OVA
Windows template on vSphere environments. It's better used in conjunction with
[TKGm](https://github.com/vmware-tanzu/tanzu-framework) <= *1.6*. The following 
pre-requisites are required to use this tool:

1. Kubeconfig from the management cluster.
2. Management cluster YAML configuration.
3. Access to a Vsphere 7 hypervisor.
4. Docker running in the local machine.
5. Windows and VMtools ISO files.

ATM the tool support creating a new Windows OVA using 
[image-builder](https://github.com/kubernetes-sigs/image-builder)
extracting existent template information, can be used to debug template information and properties
after the build or any other existent one.

## vSphere Windows Image Builder

<img src="https://user-images.githubusercontent.com/1223213/190880625-527893bb-c42f-4ca9-b85b-6b9d2d68f133.png" data-canonical-src="https://user-images.githubusercontent.com/1223213/190880625-527893bb-c42f-4ca9-b85b-6b9d2d68f133.png" width="80%" />


Build a Windows OVA with working defaults.

```shell
$ tkw vsphere template build \
  --config=${PWD}/examples/mgmt.yaml \
  --kubeconfig=${PWD}/.kube/config \
  --isopath=${PWD}/examples/isos/windows.iso
  --vmisopath=${PWD}/examples/isos/vmtools.iso
```

NOTE: Manual steps are provided [here](https://docs.vmware.com/en/VMware-Tanzu-Kubernetes-Grid/1.6/vmware-tanzu-kubernetes-grid-16/GUID-build-images-windows.html).

## vSphere Template List

<img src="https://user-images.githubusercontent.com/1223213/190836839-c6791eff-f109-4a30-821d-64f68c18c0b8.png" data-canonical-src="https://user-images.githubusercontent.com/1223213/190836839-c6791eff-f109-4a30-821d-64f68c18c0b8.png" width="50%" />

Make sure your configuration file exists on `examples/mgmt.yaml`, this file is parsed
and all configuration is mapped internally. To list the existent vSphere templates and 
vApp properties defined for the template run:

```shell
$ tkw vsphere template list --config=${PWD}/examples/mgmt.yaml
```
