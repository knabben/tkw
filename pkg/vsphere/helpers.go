package vsphere

import (
	"context"
	"fmt"
	"github.com/knabben/tkw/pkg/vsphere/models"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25/types"
	"net/url"
	"strings"
)

// ConnectFilterDC connects on vSphere and login using credentials
func ConnectFilterDC(ctx context.Context, vc, user, pass string) (Client, *models.VSphereDatacenter, error) {
	var (
		client Client
		err    error
	)
	client, err = ConnectVCLogin(vc, user, pass)
	if err != nil {
		return nil, nil, err
	}

	// Search for the existence of the datacenter.
	var dc *models.VSphereDatacenter
	dc, err = FilterDatacenter(ctx, client, "/dc0")
	if err != nil {
		return nil, nil, err
	}

	return client, dc, nil
}

// ConnectVCLogin returns the logged client.
func ConnectVCLogin(server, username, password string) (Client, error) {
	var ctx = context.Background()
	if !strings.HasPrefix(server, "http") {
		server = "https://" + server
	}

	vc, err := url.Parse(server)
	if err != nil {
		return nil, err
	}
	vc.Path = "/sdk"
	vcClient, err := NewClient(vc, "", true)
	if err != nil {
		return nil, err
	}
	_, err = vcClient.Login(ctx, username, password)
	if err != nil {
		return nil, err
	}

	return vcClient, nil
}

func (c *DefaultClient) createContainerView(ctx context.Context, parentID string, viewTypes []string) (*view.ContainerView, error) {
	m := view.NewManager(c.vmomiClient.Client)
	container := &c.vmomiClient.Client.ServiceContent.RootFolder
	if parentID != "" {
		container = &types.ManagedObjectReference{}
		if !container.FromString(parentID) {
			return nil, fmt.Errorf("incorrect managed object reference format for %s", parentID)
		}
	}

	return m.CreateContainerView(ctx, *container, viewTypes, true)
}
