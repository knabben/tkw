package vsphere

import (
	"context"
	"fmt"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/soap"
)

func (c *DefaultClient) FindDatastore(ctx context.Context, dcPath, path string) (*object.Datastore, error) {
	finder, err := c.newFinder(ctx, dcPath)
	if err != nil {
		return nil, err
	}
	obj, err := finder.Datastore(ctx, path)
	if err != nil {
		return nil, err
	}
	return obj, nil
}

func (c *DefaultClient) Upload(ctx context.Context, src, dst string, obj *object.Datastore) error {
	p := soap.DefaultUpload

	u, ticket, err := obj.ServiceTicket(ctx, dst, p.Method)
	if err != nil {
		return err
	}
	p.Ticket = ticket

	return obj.Client().UploadFile(ctx, src, u, &p)
}

func (c *DefaultClient) newFinder(ctx context.Context, dcPath string) (*find.Finder, error) {
	if c.vmomiClient == nil {
		return nil, fmt.Errorf("uninitialized vmomi client")
	}
	finder := find.NewFinder(c.vmomiClient.Client)

	if dcPath != "" {
		dc, err := finder.Datacenter(ctx, dcPath)
		if err != nil {
			return nil, err
		}
		_ = finder.SetDatacenter(dc)
	}

	return finder, nil
}
