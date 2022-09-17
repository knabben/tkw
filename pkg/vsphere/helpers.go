// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package vc ...
package vsphere

import (
	"context"
	"fmt"
	"time"

	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25/types"
)

const (
	dialTCPTimeout = 5 * time.Second
)

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
