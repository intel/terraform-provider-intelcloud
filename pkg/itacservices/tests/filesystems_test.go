package itacservices_test

import (
	"context"
	"encoding/json"
	"log"
	"terraform-provider-intelcloud/pkg/itacservices"
	"terraform-provider-intelcloud/pkg/mocks"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetFilesystems_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish) // Safer than defer

	mockAPI := mocks.NewMockAPIClient(ctrl)

	client := &itacservices.IDCServicesClient{
		Host:         strPtr("https://example.com"),
		Cloudaccount: strPtr("cloudacct-1"),
		Apitoken:     strPtr("token"),
		APIClient:    mockAPI,
	}

	ctx := context.Background()

	mockAPI.EXPECT().
		ParseString(gomock.Any(), gomock.Any()).
		Return("https://example.com/v1/cloudaccounts/cloudacct-1/filesystems?metadata.filterType=ComputeGeneral", nil).AnyTimes()

	mockAPI.EXPECT().
		MakeGetAPICall(gomock.AssignableToTypeOf(context.Background()), gomock.Any(), gomock.Any(), gomock.Nil()).
		Return(200, []byte(`{
			"items": [{
				"metadata": {
					"resourceId": "fs-1",
					"cloudAccountId": "cloudacct-1",
					"name": "fs-name",
					"description": "test desc",
					"creationTimestamp": "now"
				},
				"spec": {
					"request": {"storage": "100Gi"},
					"storageClass": "standard",
					"accessModes": "ReadWriteOnce",
					"filesystemType": "ext4",
					"Encrypted": false,
					"availabilityZone": "zone-a"
				},
				"status": {
					"phase": "FSReady",
					"mount": {
						"clusterAddr": "addr",
						"clusterVersion": "v1",
						"namespace": "ns",
						"username": "user",
						"password": "secret",
						"filesystemName": "fs-name"
					}
				}
			}]
		}`), nil).AnyTimes()

	filesystems, err := client.GetFilesystems(ctx)
	log.Printf("filesystems: %v", filesystems)

	assert.NoError(t, err)
	assert.Len(t, filesystems.FilesystemList, 1)
	assert.Equal(t, "FSReady", filesystems.FilesystemList[0].Status.Phase)
	assert.Equal(t, "fs-name", filesystems.FilesystemList[0].Metadata.Name)
}

func TestCreateFilesystem_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockAPI := mocks.NewMockAPIClient(ctrl)

	client := &itacservices.IDCServicesClient{
		Host:         strPtr("https://example.com"),
		Cloudaccount: strPtr("cloudacct-1"),
		Apitoken:     strPtr("token"),
		APIClient:    mockAPI,
	}

	ctx := context.Background()

	// Mock CreateFilesystem POST URL
	mockAPI.EXPECT().
		ParseString(gomock.Any(), gomock.Any()).
		Return("https://example.com/v1/cloudaccounts/cloudacct-1/filesystems", nil)

		// Mock CreateFilesystem POST call
	mockAPI.EXPECT().
		MakePOSTAPICall(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(200, []byte(`{
        "metadata": {
            "resourceId": "fs-1",
            "cloudAccountId": "cloudacct-1",
            "name": "fs-name",
            "description": "desc",
            "creationTimestamp": "now"
        },
        "spec": {
            "request": {"storage": "100Gi"},
            "storageClass": "standard",
            "accessModes": "ReadWriteOnce",
            "filesystemType": "ext4",
            "Encrypted": false,
            "availabilityZone": "zone-a"
        },
        "status": {
            "phase": "FSReady",
            "mount": {
                "clusterAddr": "addr",
                "clusterVersion": "v1",
                "namespace": "ns",
                "username": "user",
                "password": "",
                "filesystemName": "fs-name"
            }
        }
    }`), nil).
		AnyTimes()

	// Mock GetFilesystemByResourceId URL for retry logic
	mockAPI.EXPECT().
		ParseString(gomock.Any(), gomock.Any()).
		Return("https://example.com/v1/cloudaccounts/cloudacct-1/filesystems/id/fs-1", nil).
		AnyTimes() // Since retry may cause multiple attempts

	// Mock GET call for polling
	mockAPI.EXPECT().
		MakeGetAPICall(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Nil()).
		Return(200, []byte(`{
			"metadata": {
				"resourceId": "fs-1",
				"cloudAccountId": "cloudacct-1",
				"name": "fs-name",
				"description": "desc",
				"creationTimestamp": "now"
			},
			"spec": {
				"request": {"storage": "100Gi"},
				"storageClass": "standard",
				"accessModes": "ReadWriteOnce",
				"filesystemType": "ext4",
				"Encrypted": false,
				"availabilityZone": "zone-a"
			},
			"status": {
				"phase": "FSReady",
				"size": "100Gi",
				"mount": {
					"clusterAddr": "addr",
					"clusterVersion": "v1",
					"namespace": "ns",
					"username": "user",
					"password": "",
					"filesystemName": "fs-name"
				}
			}
		}`), nil).
		AnyTimes()

	// Optionally mock GenerateFilesystemLoginCredentials
	//client.GenerateFilesystemLoginCredentials = func(ctx context.Context, resourceId string) (*string, error) {
	//	p := "secret"
	//	return &p, nil
	//}

	req := &itacservices.FilesystemCreateRequest{
		Metadata: struct {
			Name string `json:"name"`
		}{
			Name: "fs-name",
		},
		Spec: struct {
			Request struct {
				Size string `json:"storage"`
			} `json:"request"`
			StorageClass     string `json:"storageClass"`
			AccessMode       string `json:"accessModes"`
			FilesystemType   string `json:"filesystemType"`
			InstanceType     string `json:"instanceType"`
			Encrypted        bool   `json:"Encrypted"`
			AvailabilityZone string `json:"availabilityZone"`
		}{
			Request: struct {
				Size string `json:"storage"`
			}{
				Size: "100Gi",
			},
			StorageClass:     "standard",
			AccessMode:       "ReadWriteOnce",
			FilesystemType:   "ext4",
			InstanceType:     "general",
			Encrypted:        false,
			AvailabilityZone: "zone-a",
		},
	}

	resp, err := client.CreateFilesystem(ctx, req)
	require.NotNil(t, resp)
	assert.NoError(t, err)
	assert.Equal(t, "fs-1", resp.Metadata.ResourceId)
	assert.Equal(t, "100Gi", resp.Spec.Request.Size)
}

func TestUpdateFilesystem_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAPI := mocks.NewMockAPIClient(ctrl)

	// Prepare test client
	client := &itacservices.IDCServicesClient{
		Host:         strPtr("https://example.com"),
		Cloudaccount: strPtr("cloudacct-1"),
		Apitoken:     strPtr("token"),
		APIClient:    mockAPI,
	}

	ctx := context.Background()

	// Sample request
	updateReq := &itacservices.FilesystemUpdateRequest{
		Metadata: struct {
			Name string `json:"name"`
		}{
			Name: "fs-name",
		},
		Payload: itacservices.FileSystemUpdatePayload{
			Spec: struct {
				Request struct {
					Size string `json:"storage"`
				} `json:"request"`
			}{
				Request: struct {
					Size string `json:"storage"`
				}{
					Size: "200Gi",
				},
			},
		},
	}

	// ---- MOCKS ----
	// Mock ParseString
	mockAPI.EXPECT().
		ParseString(gomock.Any(), gomock.Any()).
		Return("https://example.com/v1/cloudaccounts/cloudacct-1/filesystems/name/fs-name", nil)

	// Mock MakePutAPICall
	expectedPayload, err := json.Marshal(updateReq.Payload)
	require.NoError(t, err)
	mockAPI.EXPECT().
		MakePutAPICall(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Eq(expectedPayload)).
		Return(200, []byte(`{}`), nil).AnyTimes()

	err = client.UpdateFilesystem(ctx, updateReq)
	assert.NoError(t, err)
}

func strPtr(s string) *string {
	return &s
}
func TestDeleteFilesystemByResourceId_Success(t *testing.T) {
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
	resourceId := "fs-123"

	// Set up mocks
	mockAPI.EXPECT().
		ParseString(gomock.Any(), gomock.Any()).
		Return("https://example.com/v1/cloudaccounts/cloudacct-1/filesystems/name/fs-name", nil)

	mockAPI.EXPECT().
		MakeDeleteAPICall(gomock.Any(), gomock.Any(), "token", gomock.Nil()).
		Return(200, []byte(`{}`), nil)

	// Execute
	err := client.DeleteFilesystemByResourceId(ctx, resourceId)
	assert.NoError(t, err)
}
