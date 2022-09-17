package vsphere

import (
	"context"
	"github.com/vmware/govmomi/vim25/mo"
	"tkw/pkg/vsphere/models"
)

// vCenter Managed Object Type Names
const (
	TypeCluster         = "ClusterComputeResource"
	TypeComputeResource = "ComputeResource"
	TypeResourcePool    = "ResourcePool"
	TypeDatacenter      = "Datacenter"
	TypeFolder          = "Folder"
	TypeDatastore       = "Datastore"
	TypeNetwork         = "Network"
	TypeDvpg            = "DistributedVirtualPortgroup"
	TypeDvs             = "VmwareDistributedVirtualSwitch"
	TypeVirtualMachine  = "VirtualMachine"
)

// Client represents a vCenter client
type Client interface {
	Login(ctx context.Context, user, password string) (string, error)
	AcquireTicket() (string, error)
	CheckUserSessionActive() (bool, error)
	GetDatacenters(ctx context.Context) ([]*models.VSphereDatacenter, error)
	GetVirtualMachines(ctx context.Context, datacenterMOID string) ([]*models.VSphereVirtualMachine, error)
	GetVMMetadata(vm *mo.VirtualMachine) (properties map[string]string)
	GetImportedVirtualMachinesImages(ctx context.Context, datacenterMOID string) ([]mo.VirtualMachine, error)
}
