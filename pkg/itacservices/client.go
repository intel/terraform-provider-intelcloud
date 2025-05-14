package itacservices

import (
	"bytes"
	"context"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"terraform-provider-intelcloud/pkg/itacservices/common"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type IDCServicesClient struct {
	Host         *string
	Cloudaccount *string
	Apitoken     *string
	Region       *string
	Clientid     *string
	Clientsecret *string
	ExpireAt     time.Time
	APIClient    common.APIClient
}

var (
	getTokenURL = "{{.Host}}/oauth2/token"
)

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

func NewClient(ctx context.Context, host, tokenSvc, cloudaccount, clientid, clientsecret, region *string) (*IDCServicesClient, error) {
	os.Setenv("NO_PROXY", "")
	os.Setenv("no_proxy", "")

	params := struct {
		Host string
	}{
		Host: *tokenSvc,
	}

	// Parse the template string with the provided data
	parsedURL, err := common.ParseString(getTokenURL, params)
	if err != nil {
		return nil, fmt.Errorf("error parsing the url")
	}

	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", *clientid)

	req, err := http.NewRequest("POST", parsedURL, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("error creating ITAC Token request")
	}

	authStr := fmt.Sprintf("%s:%s", *clientid, *clientsecret)
	authEncoded := fmt.Sprintf("Basic %s", b64.StdEncoding.EncodeToString([]byte(authStr)))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	req.Header.Set("Authorization", authEncoded)
	client := &http.Client{Timeout: 60 * time.Second}

	tflog.Info(ctx, "making api client request", map[string]interface{}{"request": req.Header, "url": parsedURL})

	resp, err := client.Do(req)
	if err != nil {
		tflog.Info(ctx, "error making api client request", map[string]interface{}{"error": err})
		return nil, fmt.Errorf("error creating ITAC Token request")
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	retcode := resp.StatusCode
	tokenResp := TokenResponse{}
	if retcode != http.StatusOK {
		tflog.Info(ctx, "error making api client request", map[string]interface{}{"retcode": retcode, "body": string(body)})
		return nil, fmt.Errorf("error creating ITAC Token request")
	}

	if err = json.Unmarshal(body, &tokenResp); err != nil {
		tflog.Info(ctx, "error making api client request", map[string]interface{}{"error": err})
		return nil, fmt.Errorf("error creating ITAC Token request")
	}

	tflog.Info(ctx, "Token Response", map[string]interface{}{"token": tokenResp.AccessToken, "expires_in": tokenResp.ExpiresIn})
	return &IDCServicesClient{
		Host:         host,
		Cloudaccount: cloudaccount,
		Clientid:     clientid,
		Clientsecret: clientsecret,
		Region:       region,
		Apitoken:     &tokenResp.AccessToken,
		ExpireAt:     time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
	}, nil
}
