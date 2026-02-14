package itacservices

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"terraform-provider-intelcloud/pkg/itacservices/common"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	getAllMachineImagesURL        = "{{.Host}}/v1/machineimages"
	getAllInstanceTypesURL        = "{{.Host}}/v1/instancetypes"
	getPublicInstanceTypesImisURL = "{{.Host}}/v1/cloudaccounts/{{.Cloudaccount}}/iks/clusters/{{.ClusterUUID}}/metadata/imis"
)

type MachineImageResponse struct {
	Items []struct {
		Metadata struct {
			Name string `json:"name"`
		} `json:"metadata"`
		Spec struct {
			Description        string   `json:"description"`
			InstanceCategories []string `json:"instanceCategories"`
			InstanceTypes      []string `json:"instanceTypes"`
		} `json:"spec"`
		Hidden bool `json:"hidden"`
	} `json:"items"`
}

type ImisResponse struct {
	InstanceTypes []struct {
		Name      string `json:"instancetypename"`
		WorkerImi []struct {
			ImiName      string `json:"imiName"`
			Info         string `json:"info"`
			IsDefaultImi bool   `json:"isDefaultImi"`
		} `json:"workerImi"`
	} `json:"instanceTypes"`
}

type InstanceTypeResponse struct {
	Items []struct {
		Metadata struct {
			Name string `json:"name"`
		} `json:"metadata"`
		Spec struct {
			Description      string `json:"description"`
			InstanceCategory string `json:"instanceCategory"`
		} `json:"spec"`
	} `json:"items"`
}

func (client *IDCServicesClient) GetMachineImages(ctx context.Context) (*MachineImageResponse, error) {
	params := struct {
		Host string
	}{
		Host: *client.Host,
	}

	// Parse the template string with the provided data
	parsedURL, err := common.ParseString(getAllMachineImagesURL, params)
	if err != nil {
		return nil, fmt.Errorf("error parsing the url")
	}
	retcode, retval, err := common.MakeGetAPICall(ctx, parsedURL, *client.Apitoken, nil)
	if err != nil {
		tflog.Debug(ctx, "machine images api error", map[string]any{"retcode": retcode, "err": err, "token": *client.Apitoken})
		return nil, fmt.Errorf("error reading machine images")
	}
	tflog.Debug(ctx, "machine images api", map[string]any{"retcode": retcode, "retval": string(retval), "token": *client.Apitoken})
	if retcode != http.StatusOK {
		return nil, common.MapHttpError(retcode, retval)
	}

	images := MachineImageResponse{}
	if err := json.Unmarshal(retval, &images); err != nil {
		return nil, fmt.Errorf("error parsing machine image response")
	}
	return &images, nil
}

func (client *IDCServicesClient) GetInstanceTypes(ctx context.Context) (*InstanceTypeResponse, error) {
	params := struct {
		Host string
	}{
		Host: *client.Host,
	}

	// Parse the template string with the provided data
	parsedURL, err := common.ParseString(getAllInstanceTypesURL, params)
	if err != nil {
		return nil, fmt.Errorf("error parsing the url")
	}

	retcode, retval, err := common.MakeGetAPICall(ctx, parsedURL, *client.Apitoken, nil)
	if err != nil {
		return nil, fmt.Errorf("error reading machine images")
	}

	if retcode != http.StatusOK {
		return nil, common.MapHttpError(retcode, retval)
	}
	instType := InstanceTypeResponse{}
	if err := json.Unmarshal(retval, &instType); err != nil {
		return nil, fmt.Errorf("error parsing machine image response")
	}

	return &instType, nil
}

func (client *IDCServicesClient) GetImis(ctx context.Context, clusterUUID string) (*ImisResponse, error) {
	params := struct {
		Host         string
		Cloudaccount string
		ClusterUUID  string
	}{
		Host:         *client.Host,
		Cloudaccount: *client.Cloudaccount,
		ClusterUUID:  clusterUUID,
	}

	// Parse the template string with the provided data
	parsedURL, err := client.APIClient.ParseString(getPublicInstanceTypesImisURL, params)
	if err != nil {
		return nil, fmt.Errorf("error parsing the url for getting ")
	}
	retcode, retval, err := client.APIClient.MakeGetAPICall(ctx, parsedURL, *client.Apitoken, nil)
	if err != nil {
		tflog.Debug(ctx, "imis api error", map[string]any{"retcode": retcode, "err": err, "token": *client.Apitoken})
		return nil, fmt.Errorf("error reading imis")
	}
	tflog.Debug(ctx, "imis api", map[string]any{"retcode": retcode, "retval": string(retval), "token": *client.Apitoken})
	if retcode != http.StatusOK {
		return nil, common.MapHttpError(retcode, retval)
	}

	imis := ImisResponse{}
	if err := json.Unmarshal(retval, &imis); err != nil {
		return nil, fmt.Errorf("error parsing machine image response")
	}
	return &imis, nil
}
