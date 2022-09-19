package vsphere

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/session"
	"github.com/vmware/govmomi/vapi/rest"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/soap"
	"github.com/vmware/govmomi/vim25/types"
	"net/url"
	"strings"
	"tkw/pkg/vsphere/models"
)

// VSphere resource tags for tkg resource
const (
	VMGuestInfoUserDataKey = "guestinfo.userdata"
)

// Constant representing the number of version types tracked in a semver
const numOfSemverVersionNumbers = 3

// DefaultClient dafaults vc client
type DefaultClient struct {
	vmomiClient *govmomi.Client
	restClient  *rest.Client
}

// NewClient returns a new VC Client
func NewClient(vcURL *url.URL, thumbprint string, insecure bool) (Client, error) {
	vmomiClient, err := newGovmomiClient(vcURL, thumbprint, insecure)
	if err != nil {
		return nil, err
	}
	restClient := rest.NewClient(vmomiClient.Client)
	return &DefaultClient{
		vmomiClient: vmomiClient,
		restClient:  restClient,
	}, nil
}

func newGovmomiClient(vcURL *url.URL, thumbprint string, insecure bool) (*govmomi.Client, error) {
	ctx := context.Background()
	var vmomiClient *govmomi.Client
	var err error

	soapClient := soap.NewClient(vcURL, insecure)
	if !insecure && thumbprint != "" {
		soapClient.SetThumbprint(vcURL.Host, thumbprint)
	}
	vimClient, err := vim25.NewClient(ctx, soapClient)
	if err != nil {
		return nil, err
	}
	vmomiClient = &govmomi.Client{
		Client:         vimClient,
		SessionManager: session.NewManager(vimClient),
	}

	// Only login if the URL contains user information.
	if vcURL.User != nil {
		err = vmomiClient.Login(ctx, vcURL.User)
		if err != nil {
			return nil, err
		}
	}
	return vmomiClient, err
}

// Login authenticates with vCenter using user/password
func (c *DefaultClient) Login(ctx context.Context, user, password string) (string, error) {
	var err error
	var token string

	client := c.vmomiClient
	if client == nil {
		return "", fmt.Errorf("uninitialized vmomi client")
	}

	userInfo := url.UserPassword(user, password)
	if err = client.Login(ctx, userInfo); err != nil {
		return "", errors.Wrap(err, "cannot login to vc")
	}

	restClient := c.restClient
	if restClient == nil {
		return "", fmt.Errorf("uninitialized vapi rest client")
	}
	if err = restClient.Login(ctx, userInfo); err != nil {
		return "", errors.Wrap(err, "cannot login to vc")
	}

	token, err = c.AcquireTicket()
	return token, err
}

// AcquireTicket acquires a new session ticket for the user associated with
// the authenticated client.
func (c *DefaultClient) AcquireTicket() (string, error) {
	var err error
	var token string
	ctx := context.Background()

	client := c.vmomiClient
	if client == nil {
		return "", fmt.Errorf("uninitialized vmomi client")
	}

	if token, err = client.SessionManager.AcquireCloneTicket(ctx); err != nil {
		return "", errors.Wrap(err, "could not acquire ticket session")
	}

	return token, nil
}

// CheckUserSessionActive checks if a user session is Active
func (c *DefaultClient) CheckUserSessionActive() (bool, error) {
	ctx := context.Background()

	client := c.vmomiClient
	if client == nil {
		return false, fmt.Errorf("uninitialized vmomi client")
	}
	return client.SessionManager.SessionIsActive(ctx)
}

// GetDatacenters returns a list of all datacenters in the vSphere inventory.
func (c *DefaultClient) GetDatacenters(ctx context.Context) ([]*models.VSphereDatacenter, error) {
	if c.vmomiClient == nil {
		return nil, fmt.Errorf("uninitialized vmomi client")
	}
	viewTypes := []string{TypeDatacenter}
	v, err := c.createContainerView(ctx, "", viewTypes)
	if err != nil {
		return nil, errors.Wrap(err, "error creating datacenter view")
	}

	var dcs []mo.Datacenter
	err = v.Retrieve(ctx, viewTypes, []string{"name"}, &dcs)
	if err != nil {
		return nil, errors.Wrap(err, "error retrieving datacenters")
	}

	datacenters := make([]*models.VSphereDatacenter, 0, len(dcs))

	for i := range dcs {
		path, _, err := c.GetPath(ctx, dcs[i].Reference().Value)
		if err != nil {
			continue
		}

		dcModel := models.VSphereDatacenter{
			Moid: dcs[i].Reference().Value,
			Name: path,
		}
		datacenters = append(datacenters, &dcModel)
	}
	return datacenters, nil
}

// GetPath takes in the MOID of a vsphere resource and returns a fully qualified path
func (c *DefaultClient) GetPath(ctx context.Context, moid string) (string, []*models.VSphereManagementObject, error) {
	client := c.vmomiClient
	var objects []*models.VSphereManagementObject
	if moid == "" {
		return "", objects, errors.New("a non-empty moid should be passed to GetPath")
	}
	if client == nil {
		return "", []*models.VSphereManagementObject{}, fmt.Errorf("uninitialized vmomi client")
	}
	path := []string{}
	defaultFolder := ""
	for {
		ref, commonProps, resourceType, err := c.populateGoVCVars(moid)
		if err != nil {
			break
		}

		managedEntity := &mo.ManagedEntity{}
		name, err := commonProps.ObjectName(ctx)
		if err != nil {
			return "", objects, err
		}
		path = append([]string{name}, path...)
		err = commonProps.Properties(ctx, ref, []string{"parent"}, managedEntity)
		if err != nil {
			return "", objects, err
		}
		if managedEntity.Parent == nil {
			break
		}

		if isFolder(moid) && isDatacenter(managedEntity.Parent.Reference().Value) {
			defaultFolder = moid
		} else if !isDatacenter(moid) {
			obj := &models.VSphereManagementObject{
				Name:         name,
				Moid:         ref.Value,
				ParentMoid:   managedEntity.Parent.Reference().Value,
				ResourceType: resourceType,
			}

			objects = append(objects, obj)
		}
		moid = managedEntity.Parent.Reference().Value
	}

	objects = c.unsetDefaultFolder(objects, defaultFolder)

	if len(path) <= 1 {
		return "", objects, errors.New("not a valid path")
	}

	path = path[1:]
	res := "/" + strings.Join(path, "/")

	return res, objects, nil
}

func (c *DefaultClient) populateGoVCVars(moid string) (ref types.ManagedObjectReference, commonProps object.Common, resourceType string, err error) {
	switch {
	case isResourcePool(moid):
		ref = types.ManagedObjectReference{Type: TypeResourcePool, Value: moid}
		commonProps = object.NewResourcePool(c.vmomiClient.Client, ref).Common
		resourceType = models.VSphereManagementObjectResourceTypeRespool
	case isClusterComputeResource(moid):
		ref = types.ManagedObjectReference{Type: TypeCluster, Value: moid}
		commonProps = object.NewClusterComputeResource(c.vmomiClient.Client, ref).Common
		resourceType = models.VSphereManagementObjectResourceTypeCluster
	case isHostComputeResource(moid):
		ref = types.ManagedObjectReference{Type: TypeComputeResource, Value: moid}
		commonProps = object.NewComputeResource(c.vmomiClient.Client, ref).Common
		resourceType = models.VSphereManagementObjectResourceTypeHost
	case isDatastore(moid):
		ref = types.ManagedObjectReference{Type: TypeDatastore, Value: moid}
		commonProps = object.NewDatastore(c.vmomiClient.Client, ref).Common
		resourceType = models.VSphereManagementObjectResourceTypeDatastore
	case isFolder(moid):
		ref = types.ManagedObjectReference{Type: TypeFolder, Value: moid}
		commonProps = object.NewFolder(c.vmomiClient.Client, ref).Common
		resourceType = models.VSphereManagementObjectResourceTypeFolder
	case isVirtualMachine(moid):
		ref = types.ManagedObjectReference{Type: TypeVirtualMachine, Value: moid}
		commonProps = object.NewVirtualMachine(c.vmomiClient.Client, ref).Common
		resourceType = models.VSphereManagementObjectResourceTypeVM
	case isDatacenter(moid):
		ref = types.ManagedObjectReference{Type: TypeDatacenter, Value: moid}
		commonProps = object.NewDatacenter(c.vmomiClient.Client, ref).Common
		resourceType = models.VSphereManagementObjectResourceTypeDatacenter
	case isNetwork(moid):
		ref = types.ManagedObjectReference{Type: TypeNetwork, Value: moid}
		commonProps = object.NewNetwork(c.vmomiClient.Client, ref).Common
		resourceType = models.VSphereManagementObjectResourceTypeNetwork
	case isDvPortGroup(moid):
		ref = types.ManagedObjectReference{Type: TypeDvpg, Value: moid}
		commonProps = object.NewDistributedVirtualPortgroup(c.vmomiClient.Client, ref).Common
		resourceType = models.VSphereManagementObjectResourceTypeNetwork
	case isDvs(moid):
		ref = types.ManagedObjectReference{Type: TypeDvs, Value: moid}
		commonProps = object.NewDistributedVirtualSwitch(c.vmomiClient.Client, ref).Common
		resourceType = models.VSphereManagementObjectResourceTypeNetwork
	default:
		err = errors.New("moid value not recognized")
	}
	return ref, commonProps, resourceType, err
}

func (c *DefaultClient) unsetDefaultFolder(objects []*models.VSphereManagementObject, defaultFolder string) []*models.VSphereManagementObject {
	for i, obj := range objects {
		if obj.ParentMoid == defaultFolder {
			objects[i].ParentMoid = ""
		}
	}
	return objects
}

func isDuplicate(names map[string]bool, name string) bool {
	_, exists := names[name]
	return exists
}

// getVirtualMachines returns list of virtual machines in the given datacenter
func (c *DefaultClient) getVirtualMachines(ctx context.Context, datacenterMOID string) ([]mo.VirtualMachine, error) {
	if c.vmomiClient == nil {
		return nil, fmt.Errorf("uninitialized vmomi client")
	}

	var vms []mo.VirtualMachine

	viewTypes := []string{TypeVirtualMachine}

	dcRef := TypeDatacenter + ":" + datacenterMOID

	view, err := c.createContainerView(context.Background(), dcRef, viewTypes)
	if err != nil {
		return vms, errors.Wrap(err, "error creating container view")
	}

	err = view.Retrieve(ctx, viewTypes, []string{"name", "config"}, &vms)
	if err != nil {
		return vms, errors.Wrap(err, "failed to get virtual machines")
	}

	return vms, nil
}

// GetVirtualMachines gets vms under given datacenter
func (c *DefaultClient) GetVirtualMachines(ctx context.Context, datacenterMOID string) ([]*models.VSphereVirtualMachine, error) {
	results := []*models.VSphereVirtualMachine{}

	vms, err := c.getVirtualMachines(ctx, datacenterMOID)
	if err != nil {
		return results, err
	}

	for i := range vms {
		path, _, err := c.GetPath(ctx, vms[i].Self.Value)
		if err != nil {
			continue
		}
		obj := &models.VSphereVirtualMachine{Name: path, Moid: vms[i].Reference().Value}
		results = append(results, obj)
	}
	return results, nil
}

// GetImportedVirtualMachinesImages gets imported virtual machine images used for tkg
func (c *DefaultClient) GetImportedVirtualMachinesImages(ctx context.Context, datacenterMOID string) ([]mo.VirtualMachine, error) {
	if c.vmomiClient == nil {
		return nil, fmt.Errorf("uninitialized vmomi client")
	}

	var vms []mo.VirtualMachine

	viewTypes := []string{TypeVirtualMachine}

	dcRef := TypeDatacenter + ":" + datacenterMOID

	view, err := c.createContainerView(context.Background(), dcRef, viewTypes)
	if err != nil {
		return vms, errors.Wrap(err, "error creating container view")
	}

	filter := property.Filter{}
	filter["runtime.powerState"] = types.VirtualMachinePowerStatePoweredOff

	var content []types.ObjectContent
	err = view.Retrieve(ctx, viewTypes, filter.Keys(), &content)
	if err != nil {
		return vms, err
	}

	objs := filter.MatchObjectContent(content)
	if len(objs) == 0 {
		return vms, nil
	}
	pc := property.DefaultCollector(c.vmomiClient.Client)

	err = pc.Retrieve(ctx, objs, []string{"name", "config", "runtime.powerState"}, &vms)
	if err != nil {
		return vms, err
	}

	results := []mo.VirtualMachine{}

	for i := range vms {
		if vms[i].Config == nil {
			continue
		}

		if vms[i].Config.Template {
			results = append(results, vms[i])
			continue
		}
		isImported := true
		for _, exConfig := range vms[i].Config.ExtraConfig {
			// user-imported node image should not have the the key
			if exConfig.GetOptionValue().Key == VMGuestInfoUserDataKey {
				isImported = false
				break
			}
		}
		if isImported {
			results = append(results, vms[i])
		}
	}

	return results, nil
}

func (c *DefaultClient) GetVMMetadata(vm *mo.VirtualMachine) (properties map[string]string) {
	if vm.Config == nil {
		return
	}
	if vm.Config.VAppConfig == nil {
		return
	}

	vmConfigInfo := vm.Config.VAppConfig.GetVmConfigInfo()
	if vmConfigInfo == nil {
		return
	}

	properties = map[string]string{}
	for i := range vmConfigInfo.Property {
		p := &vmConfigInfo.Property[i]
		properties[p.Id] = p.DefaultValue
	}
	return
}

func isFolder(moID string) bool {
	return strings.HasPrefix(moID, "group-") || strings.HasPrefix(moID, "folder-")
}

func isResourcePool(moID string) bool {
	return strings.HasPrefix(moID, "resgroup-")
}

func isClusterComputeResource(moID string) bool {
	return strings.HasPrefix(moID, "domain-c") || strings.HasPrefix(moID, "clustercomputeresource-")
}

func isHostComputeResource(moID string) bool {
	return strings.HasPrefix(moID, "domain-s")
}

func isDatacenter(moID string) bool {
	return strings.HasPrefix(moID, "datacenter-")
}

func isDatastore(moID string) bool {
	return strings.HasPrefix(moID, "datastore-")
}

func isVirtualMachine(moID string) bool {
	return strings.HasPrefix(moID, "vm-")
}

func isNetwork(moID string) bool {
	return strings.HasPrefix(moID, "network-")
}

func isDvPortGroup(moID string) bool {
	return strings.HasPrefix(moID, "dvportgroup-")
}

func isDvs(moID string) bool {
	return strings.HasPrefix(moID, "dvs-")
}
