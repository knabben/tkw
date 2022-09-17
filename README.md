# <img src="https://upload.wikimedia.org/wikipedia/commons/8/87/Windows_logo_-_2021.svg" data-canonical-src="https://upload.wikimedia.org/wikipedia/commons/8/87/Windows_logo_-_2021.svg" width="3%"/>   <img src="https://avatars.githubusercontent.com/u/54452117?s=200&v=4" data-canonical-src="https://avatars.githubusercontent.com/u/54452117?s=200&v=4" width="3%"/> Tanzu Kubernetes Grid Windows Toolbox


## vSphere Windows Image Build

Build a Windows node with working defaults.

```shell
$ tkw vsphere win build --config=${PWD}/examples/mgmt.yaml
```

## vSphere Template List


<img src="https://user-images.githubusercontent.com/1223213/190836839-c6791eff-f109-4a30-821d-64f68c18c0b8.png" data-canonical-src="https://user-images.githubusercontent.com/1223213/190836839-c6791eff-f109-4a30-821d-64f68c18c0b8.png" width="50%" />

Make sure your configuration file exists on `examples/mgmt.yaml`, this file is parsed
and all configuration is mapped internally. To list the existent vSphere templates and 
vApp properties defined for the template run:

```shell
$ tkw vsphere template list --config=${PWD}/examples/mgmt.yaml
```

## vSphere Template Upload

Upload an existent OVA to a picked up datacenter and mark it as a VM template.

```shell
$ tkw vsphere template upload --config=${PWD}/examples/mgmt.yaml --ova-url="http://"
```

