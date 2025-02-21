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
	getAllSSHKeysURLByAccount = "{{.Host}}/v1/cloudaccounts/{{.Cloudaccount}}/sshpublickeys"
	createSSHKeyURL           = "{{.Host}}/v1/cloudaccounts/{{.Cloudaccount}}/sshpublickeys"
	getSSHKeyByResourceId     = "{{.Host}}/v1/cloudaccounts/{{.Cloudaccount}}/sshpublickeys/id/{{.ResourceId}}"
	deleteSSHKeyByResourceId  = "{{.Host}}/v1/cloudaccounts/{{.Cloudaccount}}/sshpublickeys/id/{{.ResourceId}}"
)

type SSHKeys struct {
	SSHKey []SSHKey `json:"items"`
}

type SSHKey struct {
	Metadata struct {
		ResourceId   string `json:"resourceId"`
		Cloudaccount string `json:"cloudAccountId"`
		Name         string `json:"name"`
		Description  string `json:"description"`
	} `json:"metadata"`
	Spec struct {
		SSHPublicKey string `json:"sshPublicKey"`
		OwnerEmail   string `json:"ownerEmail"`
	} `json:"spec"`
}

type SSHKeyCreateRequest struct {
	Metadata struct {
		Name string `json:"name"`
	} `json:"metadata"`
	Spec struct {
		SSHPublicKey string `json:"sshPublicKey"`
		OwnerEmail   string `json:"ownerEmail"`
	} `json:"spec"`
}

func (client *IDCServicesClient) GetSSHKeys(ctx context.Context) (*SSHKeys, error) {
	params := struct {
		Host         string
		Cloudaccount string
	}{
		Host:         *client.Host,
		Cloudaccount: *client.Cloudaccount,
	}

	// Parse the template string with the provided data
	parsedURL, err := common.ParseString(getAllSSHKeysURLByAccount, params)
	if err != nil {
		return nil, fmt.Errorf("error parsing the url")
	}

	retcode, retval, err := common.MakeGetAPICall(ctx, parsedURL, *client.Apitoken, nil)
	tflog.Debug(ctx, "sshkeys read api", map[string]any{"retcode": retcode, "retval": string(retval)})
	if err != nil {
		return nil, fmt.Errorf("error reading sshkeys")
	}
	if retcode != http.StatusOK {
		return nil, common.MapHttpError(retcode, retval)
	}

	sshkeys := SSHKeys{}
	if err := json.Unmarshal(retval, &sshkeys); err != nil {
		return nil, fmt.Errorf("error parsing sshkey response")
	}
	return &sshkeys, nil
}

func (client *IDCServicesClient) CreateSSHkey(ctx context.Context, in *SSHKeyCreateRequest) (*SSHKey, error) {
	params := struct {
		Host         string
		Cloudaccount string
	}{
		Host:         *client.Host,
		Cloudaccount: *client.Cloudaccount,
	}

	// Parse the template string with the provided data
	parsedURL, err := common.ParseString(createSSHKeyURL, params)
	if err != nil {
		return nil, fmt.Errorf("error parsing the url")
	}

	inArgs, err := json.MarshalIndent(in, "", "    ")
	if err != nil {
		return nil, fmt.Errorf("error parsing input arguments")
	}

	tflog.Debug(ctx, "sshkey create api request", map[string]any{"url": parsedURL, "inArgs": string(inArgs)})
	retcode, retval, err := common.MakePOSTAPICall(ctx, parsedURL, *client.Apitoken, inArgs)
	tflog.Debug(ctx, "sshkey create api response", map[string]any{"retcode": retcode, "retval": string(retval)})
	if err != nil {
		return nil, fmt.Errorf("error reading sshkey create response")
	}

	if retcode != http.StatusOK {
		return nil, common.MapHttpError(retcode, retval)
	}

	sshkey := SSHKey{}
	if err := json.Unmarshal(retval, &sshkey); err != nil {
		return nil, fmt.Errorf("error parsing sshkey response")
	}
	return &sshkey, nil
}

func (client *IDCServicesClient) GetSSHKeyByResourceId(ctx context.Context, resourceId string) (*SSHKey, error) {
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
	parsedURL, err := common.ParseString(getSSHKeyByResourceId, params)
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

	tflog.Debug(ctx, "sshkey read api", map[string]any{"retcode": retcode})
	sshkey := SSHKey{}
	if err := json.Unmarshal(retval, &sshkey); err != nil {
		return nil, fmt.Errorf("error parsing sshkey response")
	}
	return &sshkey, nil
}

func (client *IDCServicesClient) DeleteSSHKeyByResourceId(ctx context.Context, resourceId string) error {
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
	parsedURL, err := common.ParseString(deleteSSHKeyByResourceId, params)
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

	tflog.Debug(ctx, "sshkey delete api", map[string]any{"retcode": retcode})

	return nil
}
