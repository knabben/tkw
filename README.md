# Tanzu Kubernetes Grid Windows

## vSphere image build

Build a Windows node with working defaults.

```shell
$ tkw vsphere win build --config=${PWD}/examples/mgmt.yaml
```

## vSphere template list

<img src="https://user-images.githubusercontent.com/1223213/190836464-fc876b49-d1dc-4fe2-986b-01a234f4ce23.png" data-canonical-src="https://user-images.githubusercontent.com/1223213/190836464-fc876b49-d1dc-4fe2-986b-01a234f4ce23.png" width="50%" />

Make sure your configuration file exists on `examples/mgmt.yaml`, this file is parsed
and all configuration is mapped internally. To list the existent vSphere templates and 
vApp properties defined for the template run:

```shell
$ tkw vsphere template list --config=${PWD}/examples/mgmt.yaml
```

## vSphere template upload

Upload an existent OVA to a picked up datacenter and mark it as a VM template.

```shell
$ tkw vsphere template upload --config=${PWD}/examples/mgmt.yaml --ova-url="http://"
```

