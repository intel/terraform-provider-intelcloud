package itacservices

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"terraform-provider-intelcloud/pkg/itacservices/common"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	retry "github.com/sethvargo/go-retry"
)

var (
	getAllInstancesByAccount   = "{{.Host}}/v1/cloudaccounts/{{.Cloudaccount}}/instances"
	createInstance             = "{{.Host}}/v1/cloudaccounts/{{.Cloudaccount}}/instances"
	getInstanceByResourceId    = "{{.Host}}/v1/cloudaccounts/{{.Cloudaccount}}/instances/id/{{.ResourceId}}"
	deleteInstanceByResourceId = "{{.Host}}/v1/cloudaccounts/{{.Cloudaccount}}/instances/id/{{.ResourceId}}"

	getAllVNetsByAccount = "{{.Host}}/v1/cloudaccounts/{{.Cloudaccount}}/vnets"
	createVNetByAccount  = "{{.Host}}/v1/cloudaccounts/{{.Cloudaccount}}/vnets"
)

type Instances struct {
	Instances []Instance `json:"items"`
}

type Instance struct {
	Metadata struct {
		ResourceId   string `json:"resourceId"`
		Cloudaccount string `json:"cloudAccountId"`
		Name         string `json:"name"`
		CreatedAt    string `json:"creationTimestamp"`
	} `json:"metadata"`
	Spec struct {
		AvailabilityZone string `json:"availabilityZone"`
		InstanceGroup    string `json:"instanceGroup,omitempty"`
		InstanceType     string `json:"instanceType"`
		Interfaces       []struct {
			Name string `json:"name"`
			VNet string `json:"vnet"`
		} `json:"interfaces"`
		MachineImage        string   `json:"machineImage"`
		SshPublicKeyNames   []string `json:"sshPublicKeyNames"`
		UserData            string   `json:"userData,omitempty"`
		QuickConnectEnabled string   `json:"quickConnectEnabled,omitempty"`
		QuickConnectUrl     string   `json:"quickConnectUrl,omitempty"`
	} `json:"spec"`
	Status struct {
		Interfaces []struct {
			Addresses    []string `json:"addresses"`
			DNSName      string   `json:"dnsName"`
			Gateway      string   `json:"gateway"`
			Name         string   `json:"name"`
			PrefixLength int64    `json:"prefixLength"`
			Subnet       string   `json:"subnet"`
			VNet         string   `json:"vNet"`
		} `json:"interfaces"`
		Message  string `json:"message"`
		Phase    string `json:"phase"`
		SSHProxy struct {
			Address string `json:"proxyAddress"`
			Port    int64  `json:"proxyPort"`
			User    string `json:"proxyUser"`
		} `json:"sshProxy"`
		UserName string `json:"userName"`
	}
}

type InstanceCreateRequest struct {
	Metadata struct {
		Name string `json:"name"`
	} `json:"metadata"`
	Spec struct {
		AvailabilityZone string `json:"availabilityZone"`
		InstanceGroup    string `json:"instanceGroup,omitempty"`
		InstanceType     string `json:"instanceType"`
		Interfaces       []struct {
			Name string `json:"name"`
			VNet string `json:"vNet"`
		} `json:"interfaces"`
		MachineImage        string   `json:"machineImage"`
		SshPublicKeyNames   []string `json:"sshPublicKeyNames"`
		UserData            string   `json:"userData,omitempty"`
		QuickConnectEnabled string   `json:"quickConnectEnabled,omitempty"`
	} `json:"spec"`
}

type VNets struct {
	Vnets []VNet `json:"items"`
}

type VNet struct {
	Metadata struct {
		ResourceId   string `json:"resourceId"`
		Cloudaccount string `json:"cloudAccountId"`
		Name         string `json:"name"`
	} `json:"metadata"`
	Spec struct {
		AvailabilityZone string `json:"availabilityZone"`
		Region           string `json:"region"`
		PrefixLength     int64  `json:"prefixLength"`
	}
}

type VNetCreateRequest struct {
	Metadata struct {
		Name string `json:"name"`
	} `json:"metadata"`
	Spec struct {
		AvailabilityZone string `json:"availabilityZone"`
		Region           string `json:"region"`
		PrefixLength     int64  `json:"prefixLength"`
	}
}

func (client *IDCServicesClient) GetInstances(ctx context.Context) (*Instances, error) {
	params := struct {
		Host         string
		Cloudaccount string
	}{
		Host:         *client.Host,
		Cloudaccount: *client.Cloudaccount,
	}

	// Parse the template string with the provided data
	parsedURL, err := common.ParseString(getAllInstancesByAccount, params)
	if err != nil {
		return nil, fmt.Errorf("error parsing the url")
	}

	retcode, retval, err := common.MakeGetAPICall(ctx, parsedURL, *client.Apitoken, nil)
	tflog.Debug(ctx, "instances read api", map[string]any{"retcode": retcode, "retval": string(retval)})
	if err != nil {
		return nil, fmt.Errorf("error reading instances")
	}

	if retcode != http.StatusOK {
		return nil, common.MapHttpError(retcode, retval)
	}

	instances := Instances{}
	if err := json.Unmarshal(retval, &instances); err != nil {
		return nil, fmt.Errorf("error parsing instances get response, %v", err)
	}
	return &instances, nil
}

func (client *IDCServicesClient) CreateInstance(ctx context.Context, in *InstanceCreateRequest, async bool) (*Instance, error) {
	params := struct {
		Host         string
		Cloudaccount string
	}{
		Host:         *client.Host,
		Cloudaccount: *client.Cloudaccount,
	}

	// Parse the template string with the provided data
	parsedURL, err := common.ParseString(createInstance, params)
	if err != nil {
		return nil, fmt.Errorf("error parsing the url")
	}

	inArgs, err := json.MarshalIndent(in, "", "    ")
	if err != nil {
		return nil, fmt.Errorf("error parsing input arguments")
	}

	tflog.Debug(ctx, "instance create api request", map[string]any{"url": parsedURL, "inArgs": string(inArgs)})
	retcode, retval, err := common.MakePOSTAPICall(ctx, parsedURL, *client.Apitoken, inArgs)

	if err != nil {
		return nil, fmt.Errorf("error reading instance create response")
	}
	tflog.Debug(ctx, "instance create api response", map[string]any{"retcode": retcode, "retval": string(retval)})

	if retcode != http.StatusOK {
		return nil, common.MapHttpError(retcode, retval)
	}

	instance := &Instance{}
	if err := json.Unmarshal(retval, instance); err != nil {
		return nil, fmt.Errorf("error parsing instance response")
	}

	if async {
		instance, err = client.GetInstanceByResourceId(ctx, instance.Metadata.ResourceId)
		if err != nil {
			return instance, fmt.Errorf("error reading instance state")
		}
	} else {
		backoffTimer := retry.NewConstant(5 * time.Second)
		backoffTimer = retry.WithMaxDuration(300*time.Second, backoffTimer)

		if err := retry.Do(ctx, backoffTimer, func(_ context.Context) error {
			instance, err = client.GetInstanceByResourceId(ctx, instance.Metadata.ResourceId)
			if err != nil {
				return fmt.Errorf("error reading instance state")
			}
			if instance.Status.Phase == "Ready" {
				return nil
			} else if instance.Status.Phase == "Failed" {
				return fmt.Errorf("instance state failed")
			} else {
				return retry.RetryableError(fmt.Errorf("instance state not ready, retry again"))
			}
		}); err != nil {
			return nil, fmt.Errorf("instance state not ready after maximum retries")
		}
	}
	return instance, nil
}

func (client *IDCServicesClient) GetInstanceByResourceId(ctx context.Context, resourceId string) (*Instance, error) {
	params := struct {
		Host         string
		Cloudaccount string
		ResourceId   string
	}{
		Host:         *client.Host,
		Cloudaccount: *client.Cloudaccount,
		ResourceId:   resourceId,
	}

	// Parse the template string with the provided data
	parsedURL, err := common.ParseString(getInstanceByResourceId, params)
	if err != nil {
		return nil, fmt.Errorf("error parsing the url")
	}

	retcode, retval, err := common.MakeGetAPICall(ctx, parsedURL, *client.Apitoken, nil)
	if err != nil {
		return nil, fmt.Errorf("error reading sshkey by resource id")
	}

	if retcode != http.StatusOK {
		return nil, common.MapHttpError(retcode, retval)
	}

	tflog.Debug(ctx, "get instance api", map[string]any{"retcode": retcode})
	instance := Instance{}
	if err := json.Unmarshal(retval, &instance); err != nil {
		return nil, fmt.Errorf("error parsing get instance response")
	}
	return &instance, nil
}

func (client *IDCServicesClient) DeleteInstanceByResourceId(ctx context.Context, resourceId string) error {
	params := struct {
		Host         string
		Cloudaccount string
		ResourceId   string
	}{
		Host:         *client.Host,
		Cloudaccount: *client.Cloudaccount,
		ResourceId:   resourceId,
	}

	// Parse the template string with the provided data
	parsedURL, err := common.ParseString(deleteInstanceByResourceId, params)
	if err != nil {
		return fmt.Errorf("error parsing the url")
	}

	retcode, retval, err := common.MakeDeleteAPICall(ctx, parsedURL, *client.Apitoken, nil)
	if err != nil {
		return fmt.Errorf("error deleting sshkey by resource id")
	}

	if retcode != http.StatusOK {
		return common.MapHttpError(retcode, retval)
	}

	tflog.Debug(ctx, "instance delete api", map[string]any{"retcode": retcode})

	return nil
}

func (client *IDCServicesClient) CreateVNetIfNotFound(ctx context.Context) (*VNet, error) {

	params := struct {
		Host         string
		Cloudaccount string
	}{
		Host:         *client.Host,
		Cloudaccount: *client.Cloudaccount,
	}

	// Parse the template string with the provided data
	parsedURL, err := common.ParseString(getAllVNetsByAccount, params)
	if err != nil {
		return nil, fmt.Errorf("error parsing the url")
	}
	tflog.Debug(ctx, "vnets get api request", map[string]any{"url": parsedURL})

	retcode, retval, err := common.MakeGetAPICall(ctx, parsedURL, *client.Apitoken, nil)

	if err != nil || retcode != http.StatusOK {
		tflog.Debug(ctx, "vnet get response", map[string]any{"retcode": retcode, "error": err})
		return nil, fmt.Errorf("error reading vnets get response")
	}

	vnets := VNets{}
	if err := json.Unmarshal(retval, &vnets); err != nil {
		return nil, fmt.Errorf("error parsing instance response")
	}
	tflog.Debug(ctx, "vnets get api response", map[string]any{"retcode": retcode, "retval": vnets})

	if len(vnets.Vnets) > 0 {
		return &(vnets.Vnets[0]), nil
	}

	tflog.Debug(ctx, "vnets not found, creating a new")

	inArgs := VNetCreateRequest{
		Metadata: struct {
			Name string "json:\"name\""
		}{
			Name: "us-staging-1a-default",
		},
		Spec: struct {
			AvailabilityZone string "json:\"availabilityZone\""
			Region           string "json:\"region\""
			PrefixLength     int64  "json:\"prefixLength\""
		}{
			AvailabilityZone: "us-staging-1a",
			Region:           "us-staging-1",
			PrefixLength:     24,
		},
	}

	payload, err := json.MarshalIndent(inArgs, "", "    ")
	if err != nil {
		return nil, fmt.Errorf("error parsing input arguments")
	}

	// Parse the template string with the provided data
	parsedURL, err = common.ParseString(createVNetByAccount, params)
	if err != nil {
		return nil, fmt.Errorf("error parsing the url")
	}

	retcode, retval, err = common.MakePOSTAPICall(ctx, parsedURL, *client.Apitoken, payload)

	if err != nil || retcode != http.StatusOK {
		return nil, fmt.Errorf("error reading vnet create response")
	}

	vnet := VNet{}
	if err := json.Unmarshal(retval, &vnet); err != nil {
		return nil, fmt.Errorf("error parsing vnet response")
	}
	tflog.Debug(ctx, "vnet create api response", map[string]any{"retcode": retcode, "retval": vnet})

	return &vnet, nil

}
