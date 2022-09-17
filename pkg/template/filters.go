package template

import (
	"context"
	"tkw/pkg/vsphere"
)

// FilterDatacenter find the datacenter object in the mapper and returns the MOID string
func (m Mapper) FilterDatacenter(ctx context.Context, client vsphere.Client, dcName string) (string, error) {
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
