// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package vc ...
package vsphere

import (
	"context"
	"fmt"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25/types"
	"net/url"
	"strings"
	"time"
	"tkw/pkg/config"
	"tkw/pkg/template"
)

const (
	dialTCPTimeout = 5 * time.Second
)

func ConnectAndFilterDC(ctx context.Context, mapper *config.Mapper) (Client, string, error) {
	var client Client

	vsphereServer := mapper.Get(VsphereServer)
	vsphereTlsThumbprint := mapper.Get(VsphereTlsThumbprint)
	vsphereUsername := mapper.Get(VsphereUsername)
	vspherePassword := mapper.Get(VspherePassword)

	// Connecting vSphere server with configuration.
	message := fmt.Sprintf("Connecting to the vSphere server... %s", vsphereServer)
	tmpStyle := template.BaseStyle.Copy()
	fmt.Println(tmpStyle.Padding(3, 2, 3, 2).Render(message))

	client, err := ConnectVCAndLogin(vsphereServer, vsphereTlsThumbprint, vsphereUsername, vspherePassword)
	if err != nil {
		return nil, "", err
	}

	// Search for existent Datacenter.
	dcMOID, err := FilterDatacenter(ctx, client, mapper.Get(VsphereDataCenter))
	if err != nil {
		return nil, "", err
	}
	return client, dcMOID, nil
}

// ConnectVCAndLogin returns the logged client.
func ConnectVCAndLogin(server, tlsThumbprint, username, password string) (Client, error) {
	var ctx = context.Background()
	if !strings.HasPrefix(server, "http") {
		server = "https://" + server
	}

	vc, err := url.Parse(server)
	if err != nil {
		return nil, err
	}
	vc.Path = "/sdk"
	vcClient, err := NewClient(vc, tlsThumbprint, false)
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
