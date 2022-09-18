package vsphere

// code from: hashicorp/terraform-provider-vsphere

import (
	"context"
	"fmt"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/types"
	"path"
	"time"
)

const DefaultAPITimeout = time.Minute * 5

// Browser returns the HostDatastoreBrowser for a certain datastore. This is a
// convenience method that exists to abstract the context.
func Browser(ds *object.Datastore) (*object.HostDatastoreBrowser, error) {
	ctx, cancel := context.WithTimeout(context.Background(), DefaultAPITimeout)
	defer cancel()
	return ds.Browser(ctx)
}

// FileExists takes a path in the datastore and checks to see if it exists.
//
// The path should be a bare path, not a datastore path. Globs are not allowed.
func FileExists(ds *object.Datastore, name string) (bool, error) {
	files, err := SearchDatastore(ds, name)
	if err != nil {
		return false, err
	}
	if len(files) > 1 {
		return false, fmt.Errorf("multiple results returned for %q in datastore %q, use a more specific search", name, ds)
	}
	if len(files) < 1 {
		return false, nil
	}
	return path.Base(name) == files[0].Path, nil
}

func searchDatastore(ds *object.Datastore, name string) (*types.HostDatastoreBrowserSearchResults, error) {
	browser, err := Browser(ds)
	if err != nil {
		return nil, err
	}
	var p, m string

	switch {
	case path.Dir(name) == ".":
		fallthrough
	case path.Base(name) == "":
		p = name
		m = "*"
	default:
		p = path.Dir(name)
		m = path.Base(name)
	}
	dp := &object.DatastorePath{
		Datastore: ds.Name(),
		Path:      p,
	}
	spec := &types.HostDatastoreBrowserSearchSpec{
		MatchPattern: []string{m},
		Details: &types.FileQueryFlags{
			FileType:     true,
			FileSize:     true,
			FileOwner:    types.NewBool(true),
			Modification: true,
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), DefaultAPITimeout)
	defer cancel()
	task, err := browser.SearchDatastore(ctx, dp.String(), spec)
	if err != nil {
		return nil, err
	}
	tctx, tcancel := context.WithTimeout(context.Background(), DefaultAPITimeout)
	defer tcancel()
	info, err := task.WaitForResult(tctx, nil)
	if err != nil {
		return nil, err
	}
	r := info.Result.(types.HostDatastoreBrowserSearchResults)
	return &r, nil
}

// SearchDatastore searches a datastore using the supplied HostDatastoreBrowser
// and a supplied path. The current implementation only returns the basic
// information, so all FileQueryFlags set, but not any flags for specific types
// of files.
func SearchDatastore(ds *object.Datastore, name string) ([]*types.FileInfo, error) {
	result, err := searchDatastore(ds, name)
	if err != nil {
		return nil, err
	}
	var files []*types.FileInfo
	for _, bfi := range result.File {
		files = append(files, bfi.GetFileInfo())
	}
	return files, nil
}
