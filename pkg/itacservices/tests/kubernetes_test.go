package itacservices_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"terraform-provider-intelcloud/pkg/itacservices"
	"terraform-provider-intelcloud/pkg/itacservices/common"
	"terraform-provider-intelcloud/pkg/mocks"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/sethvargo/go-retry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	createK8sClusterURL = "{{.Host}}/v1/cloudaccounts/{{.Cloudaccount}}/iks/clusters"
)

type testIDCServicesClient struct {
	*itacservices.IDCServicesClient
	OverrideGetByUUID func(ctx context.Context, uuid string) (*itacservices.IKSCluster, *string, error)
}

func (c *testIDCServicesClient) GetIKSClusterByClusterUUID(ctx context.Context, uuid string) (*itacservices.IKSCluster, *string, error) {
	if c.OverrideGetByUUID != nil {
		return c.OverrideGetByUUID(ctx, uuid)
	}
	// Fallback to the real method if no override is set
	return c.IDCServicesClient.GetIKSClusterByClusterUUID(ctx, uuid)
}

// Implementing it here instead of direct mocks to avoid circular dependency
func (c *testIDCServicesClient) CreateIKSCluster(ctx context.Context, in *itacservices.IKSCreateRequest, async bool) (*itacservices.IKSCluster, *string, error) {
	params := struct {
		Host         string
		Cloudaccount string
	}{
		Host:         *c.Host,
		Cloudaccount: *c.Cloudaccount,
	}

	// Parse the template string with the provided data
	parsedURL, err := c.APIClient.ParseString(createK8sClusterURL, params)
	if err != nil {
		return nil, nil, fmt.Errorf("error parsing the url")
	}

	inArgs, err := json.MarshalIndent(in, "", "    ")
	if err != nil {
		return nil, nil, fmt.Errorf("error parsing input arguments")
	}

	tflog.Debug(ctx, "iks create api request", map[string]any{"url": parsedURL, "inArgs": string(inArgs)})
	retcode, retval, err := c.APIClient.MakePOSTAPICall(ctx, parsedURL, *c.Apitoken, inArgs)

	if err != nil {
		return nil, nil, fmt.Errorf("error reading iks create response")
	}
	tflog.Debug(ctx, "iks create api response", map[string]any{"retcode": retcode, "retval": string(retval)})

	if retcode != http.StatusOK {
		return nil, nil, common.MapHttpError(retcode, retval)
	}

	cluster := &itacservices.IKSCluster{}
	if err := json.Unmarshal(retval, cluster); err != nil {
		return nil, nil, fmt.Errorf("error parsing instance response")
	}

	if async {
		cluster, _, err = c.GetIKSClusterByClusterUUID(ctx, cluster.ResourceId)
		if err != nil {
			return cluster, nil, fmt.Errorf("error reading iks cluster state")
		}
	} else {
		backoffTimer := retry.NewConstant(5 * time.Second)
		backoffTimer = retry.WithMaxDuration(3000*time.Second, backoffTimer)

		if err := retry.Do(ctx, backoffTimer, func(_ context.Context) error {
			cluster, _, err = c.GetIKSClusterByClusterUUID(ctx, cluster.ResourceId)
			if err != nil {
				return fmt.Errorf("error reading instance state")
			}
			if cluster.ClusterState == "Active" {
				return nil
			} else if cluster.ClusterState == "Failed" {
				return fmt.Errorf("instance state failed")
			} else {
				return retry.RetryableError(fmt.Errorf("iks cluster state not ready, retry again"))
			}
		}); err != nil {
			return nil, nil, fmt.Errorf("iks cluster state not ready after maximum retries")
		}
	}

	return cluster, c.Cloudaccount, nil
}

func TestGetKubernetesClusters_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAPI := mocks.NewMockAPIClient(ctrl)

	client := &itacservices.IDCServicesClient{
		Host:         strPtr("https://example.com"),
		Cloudaccount: strPtr("cloudacct-1"),
		Apitoken:     strPtr("token"),
		APIClient:    mockAPI,
	}

	ctx := context.Background()

	// Expected URL (after template parsing)
	expectedURL := "https://example.com/v1/cloudaccounts/cloudacct-1/k8sclusters"

	// Mock ParseString call
	mockAPI.EXPECT().
		ParseString(gomock.Any(), gomock.Any()).
		Return(expectedURL, nil)

	// Mock MakeGetAPICall call
	mockAPI.EXPECT().
		MakeGetAPICall(ctx, expectedURL, "token", gomock.Nil()).
		Return(http.StatusOK, []byte(`{
		    "clusters": [
        {
            "uuid": "k8s-1",
            "name": "cluster-1",
            "description": "test cluster",
            "createddate": "2024-01-01T00:00:00Z",
            "clusterstate": "running",
            "k8sversion": "1.25",
            "upgradeavailable": true,
            "upgradek8sversionavailable": ["1.26"],
            "network": {},
            "nodegroups": [],
            "storageenabled": true,
            "storages": [],
            "vips": []
        }
    ]
}
			`), nil)

	// ---- Execute ----
	clusters, cloudaccount, err := client.GetKubernetesClusters(ctx)

	// ---- Verify ----
	require.NoError(t, err)
	require.NotNil(t, clusters)
	require.NotNil(t, cloudaccount)
	assert.Equal(t, "cloudacct-1", *cloudaccount)

	require.Len(t, clusters.Clusters, 1)
	cluster := clusters.Clusters[0]
	assert.Equal(t, "k8s-1", cluster.ResourceId)
	assert.Equal(t, "cluster-1", cluster.Name)
	assert.Equal(t, "test cluster", cluster.Description)
	assert.Equal(t, true, cluster.UpgradeAvailable)
}

func TestCreateIKSCluster_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAPI := mocks.NewMockAPIClient(ctrl)

	// Simulate state transition
	callCount := 0
	client := &testIDCServicesClient{
		IDCServicesClient: &itacservices.IDCServicesClient{
			Host:         strPtr("https://example.com"),
			Cloudaccount: strPtr("cloudacct-1"),
			Apitoken:     strPtr("token"),
			APIClient:    mockAPI,
		},
		OverrideGetByUUID: func(ctx context.Context, uuid string) (*itacservices.IKSCluster, *string, error) {
			callCount++
			state := "Provisioning"
			if callCount >= 2 {
				state = "Active"
			}
			return &itacservices.IKSCluster{
				ResourceId:   "iks-cluster-1",
				ClusterState: state,
			}, strPtr("cloudacct-1"), nil
		},
	}

	ctx := context.Background()

	createReq := &itacservices.IKSCreateRequest{
		// Fill in with appropriate mock data as per your struct definition
		Name:       "my-cluster",
		K8sVersion: "1.30",
	}

	// Expected URL
	expectedURL := "https://example.com/v1/cloudaccounts/cloudacct-1/iks/clusters/iks-cluster-1"

	// Expected cluster creation response
	response := `{
		"resourceId": "iks-cluster-1",
		"name": "my-cluster",
		"clusterState": "Active"
	}`

	// Mock ParseString
	mockAPI.EXPECT().
		ParseString(gomock.Any(), gomock.Any()).
		Return(expectedURL, nil).AnyTimes()

	mockAPI.EXPECT().
		MakePOSTAPICall(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(http.StatusOK, []byte(response), nil).
		AnyTimes()

	cluster, cloudAccount, err := client.CreateIKSCluster(ctx, createReq, false)

	assert.NoError(t, err)
	assert.NotNil(t, cluster)
	assert.Equal(t, "iks-cluster-1", cluster.ResourceId)
	assert.Equal(t, "cloudacct-1", *cloudAccount)
	assert.Equal(t, "Active", cluster.ClusterState)
}
