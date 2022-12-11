package vsphere

import (
	"context"
	"tkw/pkg/vsphere/models"
)

// FilterDatacenter find the datacenter object in the mapper and returns the DC model
func FilterDatacenter(ctx context.Context, client Client, dcName string) (*models.VSphereDatacenter, error) {
	dcs, err := client.GetDatacenters(ctx)
	if err != nil {
		return nil, err
	}
	for _, dc := range dcs {
		if dc.Name == dcName {
			return dc, nil
		}
	}
	return nil, nil
}
