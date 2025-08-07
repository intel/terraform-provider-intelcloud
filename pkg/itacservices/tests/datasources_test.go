package itacservices_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"terraform-provider-intelcloud/pkg/itacservices"
	"terraform-provider-intelcloud/pkg/mocks"
)

func TestGetImis(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAPI := mocks.NewMockAPIClient(ctrl)

	idcClient := &itacservices.IDCServicesClient{
		Host:         strPtr("https://example.com"),
		Cloudaccount: strPtr("my-account"),
		Apitoken:     strPtr("secret-token"),
		APIClient:    mockAPI,
	}

	ctx := context.Background()
	expectedURL := "https://example.com/public/cloudresources/instanceTypesImis?cloudaccount=my-account"

	// Sample ImisResponse object and corresponding JSON
	expectedResponse := itacservices.ImisResponse{
		InstanceTypes: []struct {
			Name      string `json:"instancetypename"`
			WorkerImi []struct {
				ImiName      string `json:"imiName"`
				Info         string `json:"info"`
				IsDefaultImi bool   `json:"isDefaultImi"`
			} `json:"workerImi"`
		}{
			{
				Name: "vm-spr-sml",
				WorkerImi: []struct {
					ImiName      string `json:"imiName"`
					Info         string `json:"info"`
					IsDefaultImi bool   `json:"isDefaultImi"`
				}{
					{
						ImiName:      "iks-vm-u22-wk-1-30-3-v250330",
						Info:         "",
						IsDefaultImi: true,
					},
					{
						ImiName:      "iks-vm-u22-wk-1-30-3-v250404",
						Info:         "",
						IsDefaultImi: false,
					},
				},
			},
			{
				Name: "bm-icx-gaudi2",
				WorkerImi: []struct {
					ImiName      string `json:"imiName"`
					Info         string `json:"info"`
					IsDefaultImi bool   `json:"isDefaultImi"`
				}{
					{
						ImiName:      "iks-gd-u22-cd-wk-1-30-3-habana-v1.20.1-hwe-v250403",
						Info:         "",
						IsDefaultImi: true,
					},
				},
			},
		},
	}

	// Marshal sample response to JSON
	mockJSON, err := json.Marshal(expectedResponse)
	require.NoError(t, err)

	// Expect ParseString to be called and return the expected URL
	mockAPI.EXPECT().
		ParseString(gomock.Any(), gomock.Any()).
		Return(expectedURL, nil)

	// Expect MakeGetAPICall to be called and return the sample JSON
	mockAPI.EXPECT().
		MakeGetAPICall(ctx, expectedURL, "secret-token", gomock.Nil()).
		Return(http.StatusOK, mockJSON, nil)

	// Call the function
	resp, err := idcClient.GetImis(ctx, "clusteruuid-1")

	// Verify
	require.NoError(t, err)
	require.Equal(t, expectedResponse, *resp)
	require.Equal(t, len(resp.InstanceTypes), 2)
	require.Equal(t, resp.InstanceTypes[0].Name, "vm-spr-sml")
	require.Equal(t, resp.InstanceTypes[0].WorkerImi[0].ImiName, "iks-vm-u22-wk-1-30-3-v250330")
	require.Equal(t, resp.InstanceTypes[0].WorkerImi[0].IsDefaultImi, true)
	require.Equal(t, resp.InstanceTypes[0].WorkerImi[1].ImiName, "iks-vm-u22-wk-1-30-3-v250404")
	require.Equal(t, resp.InstanceTypes[1].Name, "bm-icx-gaudi2")
}
