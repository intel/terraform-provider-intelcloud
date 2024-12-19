package itacservices

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"terraform-provider-intelcloud/pkg/itacservices/common"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	retry "github.com/sethvargo/go-retry"
)

const (
	createObjectStorageBucketURL          = "{{.Host}}/v1/cloudaccounts/{{.Cloudaccount}}/objects/buckets"
	getObjectStorageBucketByResourceId    = "{{.Host}}/v1/cloudaccounts/{{.Cloudaccount}}/objects/buckets/id/{{.ResourceId}}"
	deleteObjectStorageBucketByResourceId = "{{.Host}}/v1/cloudaccounts/{{.Cloudaccount}}/objects/buckets/id/{{.ResourceId}}"
	createObjectStorageUserURL            = "{{.Host}}/v1/cloudaccounts/{{.Cloudaccount}}/objects/users"
	deleteObjectStorageUserURL            = "{{.Host}}/v1/cloudaccounts/{{.Cloudaccount}}/objects/users/id/{{.ResourceId}}"
	getObjectStorageUserURL               = "{{.Host}}/v1/cloudaccounts/{{.Cloudaccount}}/objects/users/id/{{.ResourceId}}"
)

type ObjectBucketCreateRequest struct {
	Metadata struct {
		Name string `json:"name"`
	} `json:"metadata"`
	Spec struct {
		Versioned    bool   `json:"versioned"`
		InstanceType string `json:"instanceType"`
	} `json:"spec"`
}

type ObjectBucket struct {
	Metadata struct {
		Name         string `json:"name"`
		ResourceId   string `json:"resourceId"`
		Cloudaccount string `json:"cloudAccountId"`
	} `json:"metadata"`
	Spec struct {
		Versioned    bool   `json:"versioned"`
		InstanceType string `json:"instanceType"`
		Request      struct {
			Size string `json:"size"`
		} `json:"request"`
	} `json:"spec"`
	Status struct {
		Phase   string `json:"phase"`
		Cluster struct {
			AccessEndpoint string `json:"accessEndpoint"`
			ClusterId      string `json:"clusterId"`
		} `json:"cluster"`
		SecurityGroups struct {
			NetworkFilterAllow []struct {
				Gateway      string `json:"gateway"`
				PrefixLength int    `json:"prefixLength"`
				Subnet       string `json:"subnet"`
			} `json:"networkFilterAllow"`
		} `json:"securityGroup"`
	} `json:"status"`
}

type ObjectUserCreateRequest struct {
	Metadata struct {
		Name string `json:"name"`
	} `json:"metadata"`
	Spec []BucketPolicy `json:"spec"`
}

type BucketPolicy struct {
	BucketId    string   `json:"bucketId"`
	Actions     []string `json:"actions"`
	Permissions []string `json:"permission"`
	Prefix      string   `json:"prefix"`
}

type ObjectUser struct {
	Metadata struct {
		Name         string `json:"name"`
		UserId       string `json:"userId"`
		Cloudaccount string `json:"cloudAccountId"`
	} `json:"metadata"`
	Spec   []BucketPolicy `json:"spec"`
	Status struct {
		Phase     string `json:"phase"`
		Principal struct {
			Credentials struct {
				AccessKey string `json:"accessKey"`
				SecretKey string `json:"secretKey"`
			} `json:"credentials"`
		} `json:"principal"`
	}
}

func (client *IDCServicesClient) CreateObjectStorageBucket(ctx context.Context, in *ObjectBucketCreateRequest) (*ObjectBucket, error) {
	params := struct {
		Host         string
		Cloudaccount string
	}{
		Host:         *client.Host,
		Cloudaccount: *client.Cloudaccount,
	}

	// Parse the template string with the provided data
	parsedURL, err := common.ParseString(createObjectStorageBucketURL, params)
	if err != nil {
		return nil, fmt.Errorf("error parsing the url")
	}

	inArgs, err := json.MarshalIndent(in, "", "    ")
	if err != nil {
		return nil, fmt.Errorf("error parsing input arguments")
	}

	tflog.Debug(ctx, "bucket create api", map[string]any{"url": parsedURL, "inArgs": string(inArgs)})
	retcode, retval, err := common.MakePOSTAPICall(ctx, parsedURL, *client.Apitoken, inArgs)
	tflog.Debug(ctx, "bucket create api", map[string]any{"retcode": retcode, "retval": string(retval)})
	if err != nil {
		return nil, fmt.Errorf("error reading bucket create response")
	}
	if retcode != http.StatusOK {
		return nil, common.MapHttpError(retcode)
	}

	bucket := &ObjectBucket{}
	if err := json.Unmarshal(retval, bucket); err != nil {
		return nil, fmt.Errorf("error parsing bucket response")
	}

	backoffTimer := retry.NewConstant(5 * time.Second)
	backoffTimer = retry.WithMaxDuration(300*time.Second, backoffTimer)

	if err := retry.Do(ctx, backoffTimer, func(_ context.Context) error {
		bucket, err = client.GetObjectBucketByResourceId(ctx, bucket.Metadata.ResourceId)
		if err != nil {
			return fmt.Errorf("error reading bucket state")
		}
		if bucket.Status.Phase == "BucketReady" {
			return nil
		} else if bucket.Status.Phase == "BucketFailed" {
			return fmt.Errorf("bucket state failed")
		} else {
			return retry.RetryableError(fmt.Errorf("bucket state not ready, retry again"))
		}
	}); err != nil {
		return nil, fmt.Errorf("bucket state not ready after maximum retries")
	}

	return bucket, nil
}

func (client *IDCServicesClient) GetObjectBucketByResourceId(ctx context.Context, resourceId string) (*ObjectBucket, error) {
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
	parsedURL, err := common.ParseString(getObjectStorageBucketByResourceId, params)
	if err != nil {
		return nil, fmt.Errorf("error parsing the url")
	}

	retcode, retval, err := common.MakeGetAPICall(ctx, parsedURL, *client.Apitoken, nil)
	if err != nil {
		return nil, fmt.Errorf("error reading bucket by resource id")
	}

	if retcode != http.StatusOK {
		return nil, common.MapHttpError(retcode)
	}

	tflog.Debug(ctx, "object read api", map[string]any{"retcode": retcode})
	bucket := ObjectBucket{}
	if err := json.Unmarshal(retval, &bucket); err != nil {
		return nil, fmt.Errorf("error parsing bucket response")
	}
	return &bucket, nil
}

func (client *IDCServicesClient) DeleteBucketByResourceId(ctx context.Context, resourceId string) error {
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
	parsedURL, err := common.ParseString(deleteObjectStorageBucketByResourceId, params)
	if err != nil {
		return fmt.Errorf("error parsing the url")
	}

	retcode, retval, err := common.MakeDeleteAPICall(ctx, parsedURL, *client.Apitoken, nil)
	if err != nil {
		return fmt.Errorf("error deleting object bucket by resource id")
	}

	tflog.Debug(ctx, "object bucket delete api", map[string]any{"retcode": retcode, "retval": string(retval)})

	if retcode != http.StatusOK {
		return common.MapHttpError(retcode)
	}

	tflog.Debug(ctx, "object bucket delete api", map[string]any{"retcode": retcode})

	return nil
}

func (client *IDCServicesClient) CreateObjectStorageUser(ctx context.Context, in *ObjectUserCreateRequest) (*ObjectUser, error) {
	params := struct {
		Host         string
		Cloudaccount string
	}{
		Host:         *client.Host,
		Cloudaccount: *client.Cloudaccount,
	}

	// Parse the template string with the provided data
	parsedURL, err := common.ParseString(createObjectStorageUserURL, params)
	if err != nil {
		return nil, fmt.Errorf("error parsing the url")
	}

	inArgs, err := json.MarshalIndent(in, "", "    ")
	if err != nil {
		return nil, fmt.Errorf("error parsing input arguments")
	}

	tflog.Debug(ctx, "bucket user create api", map[string]any{"url": parsedURL, "inArgs": string(inArgs)})
	retcode, retval, err := common.MakePOSTAPICall(ctx, parsedURL, *client.Apitoken, inArgs)
	tflog.Debug(ctx, "bucket user create api", map[string]any{"retcode": retcode, "retval": string(retval)})
	if err != nil {
		return nil, fmt.Errorf("error reading bucket user create response")
	}
	if retcode != http.StatusOK {
		return nil, common.MapHttpError(retcode)
	}

	objUser := &ObjectUser{}
	if err := json.Unmarshal(retval, objUser); err != nil {
		return nil, fmt.Errorf("error parsing bucket user response")
	}
	tflog.Debug(ctx, "bucket user create api", map[string]any{"retcode": retcode, "ret object": objUser})
	return objUser, nil
}

func (client *IDCServicesClient) DeleteObjectUserByResourceId(ctx context.Context, userId string) error {
	params := struct {
		Host         string
		Cloudaccount string
		ResourceId   string
	}{
		Host:         *client.Host,
		Cloudaccount: *client.Cloudaccount,
		ResourceId:   userId,
	}

	// Parse the template string with the provided data
	parsedURL, err := common.ParseString(deleteObjectStorageUserURL, params)
	if err != nil {
		return fmt.Errorf("error parsing the url")
	}

	retcode, retval, err := common.MakeDeleteAPICall(ctx, parsedURL, *client.Apitoken, nil)
	if err != nil {
		return fmt.Errorf("error deleting object bucket user by id")
	}

	tflog.Debug(ctx, "object bucket user delete api", map[string]any{"retcode": retcode, "retval": string(retval)})

	if retcode != http.StatusOK {
		return common.MapHttpError(retcode)
	}

	tflog.Debug(ctx, "object bucket user delete api", map[string]any{"retcode": retcode})

	return nil
}

func (client *IDCServicesClient) GetObjectUserByUserId(ctx context.Context, userId string) (*ObjectUser, error) {
	params := struct {
		Host         string
		Cloudaccount string
		ResourceId   string
	}{
		Host:         *client.Host,
		Cloudaccount: *client.Cloudaccount,
		ResourceId:   userId,
	}

	// Parse the template string with the provided data
	parsedURL, err := common.ParseString(getObjectStorageUserURL, params)
	if err != nil {
		return nil, fmt.Errorf("error parsing the url")
	}

	retcode, retval, err := common.MakeGetAPICall(ctx, parsedURL, *client.Apitoken, nil)
	if err != nil {
		return nil, fmt.Errorf("error reading bucket user by id")
	}

	if retcode != http.StatusOK {
		return nil, common.MapHttpError(retcode)
	}

	tflog.Debug(ctx, "object user read api", map[string]any{"retcode": retcode})
	user := ObjectUser{}
	if err := json.Unmarshal(retval, &user); err != nil {
		return nil, fmt.Errorf("error parsing bucket response")
	}
	return &user, nil
}
