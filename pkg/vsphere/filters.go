package vsphere

import (
	"context"
)

// FilterDatacenter find the datacenter object in the mapper and returns the MOID string
func FilterDatacenter(ctx context.Context, client Client, dcName string) (string, error) {
	dcs, err := client.GetDatacenters(ctx)
	if err != nil {
		return "", err
	}
	for _, dc := range dcs {
		if dc.Name == dcName {
			return dc.Moid, nil
		}
	}
	return "", nil
}
