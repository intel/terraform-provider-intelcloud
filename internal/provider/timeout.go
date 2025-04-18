package provider

import (
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	//Instance defaults
	InstanceResourceName    = "instance"
	InstanceResourceTimeout = "15m"

	//IKS Nodegroup defaults
	IKSNodegroupResourceName    = "iksnodegroup"
	IKSNodegroupResourceTimeout = "30m"

	//IKS Cluster defaults
	IKSClusterResourceName    = "ikscluster"
	IKSClusterResourceTimeout = "30m"

	//Filesystem defaults
	FilesystemResourceName    = "filesystem"
	FileSystemResourceTimeout = "5m"

	//Objectstore default
	ObjectStorageResourceName    = "objectstorage"
	ObjectstorageResourceTimeout = "5m"
)

var DefaultTimeouts = map[string]string{
	InstanceResourceName:      InstanceResourceTimeout,
	IKSNodegroupResourceName:  IKSNodegroupResourceTimeout,
	IKSClusterResourceName:    IKSClusterResourceTimeout,
	FilesystemResourceName:    FileSystemResourceTimeout,
	ObjectStorageResourceName: ObjectstorageResourceTimeout,
}

type timeoutsModel struct {
	ResourceTimeout types.String `tfsdk:"resource_timeout"`
}

func (t *timeoutsModel) GetTimeouts(resource string) (time.Duration, error) {
	var timeout time.Duration
	var err error
	if t != nil && !t.ResourceTimeout.IsNull() {
		timeout, err = time.ParseDuration(t.ResourceTimeout.ValueString())
		if err != nil {
			return 0, fmt.Errorf("invalid timeout value for resource %s: %v", resource, err)
		}
	} else {
		timeout, _ = time.ParseDuration(DefaultTimeouts[resource])
	}
	fmt.Printf("Timeout for %s resource is %s\n", resource, timeout)
	return timeout, err
}
