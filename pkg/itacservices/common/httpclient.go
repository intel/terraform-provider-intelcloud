package common

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

type APIClient interface {
	MakeGetAPICall(ctx context.Context, url, token string, headers map[string]string) (int, []byte, error)
	MakePOSTAPICall(ctx context.Context, url, token string, payload []byte) (int, []byte, error)
	MakePutAPICall(ctx context.Context, url, token string, payload []byte) (int, []byte, error)
	MakeDeleteAPICall(ctx context.Context, url, token string, headers map[string]string) (int, []byte, error)
	GenerateFilesystemLoginCredentials(ctx context.Context, resourceId string) (*string, error)
	ParseString(tmpl string, data any) (string, error)
}

// MakeGetAPICall :
func MakeGetAPICall(ctx context.Context, connURL, auth string, payload []byte) (int, []byte, error) {

	req, err := http.NewRequest("GET", connURL, bytes.NewBuffer(payload))
	if err != nil {
		return http.StatusInternalServerError, nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if auth != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", auth))
	}
	printRequest(req)
	retries := 3
	body := []byte{}
	retcode := http.StatusOK
	for try := 1; try <= retries; try++ {
		client := &http.Client{Timeout: 60 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			if try == retries {
				return http.StatusInternalServerError, nil,
					errors.New("error conencting to  api service")
			}
			time.Sleep(5 * time.Second)
			continue
		}
		defer resp.Body.Close()
		body, _ = io.ReadAll(resp.Body)
		retcode = resp.StatusCode
		break
	}
	return retcode, body, nil
}

// MakePOSTAPICall :
func MakePOSTAPICall(ctx context.Context, connURL, auth string, payload []byte) (int, []byte, error) {

	req, err := http.NewRequest("POST", connURL, bytes.NewBuffer(payload))
	if err != nil {
		return http.StatusInternalServerError, nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	auth = "eyJhbGciOiJSUzI1NiIsInR5cCIgOiAiSldUIiwia2lkIiA6ICJvWUhGY3lhcXU3bFpnNzlfQkNpcEtNWkQ5YlRQOTZ4Z1lLQkt3b1kybUhnIn0.eyJleHAiOjE3NDg0NjM3NzksImlhdCI6MTc0ODQ2MDE3OSwianRpIjoiMTQzMGE1MDMtMDljOC00NGFkLWFjMjAtNDFkNDRmNzk0OTcxIiwiaXNzIjoiaHR0cHM6Ly9jbGllbnQtdG9rZW4uc3RhZ2luZy5hcGkuaWRjc2VydmljZS5uZXQvYXV0aC9yZWFsbXMvbWFzdGVyIiwiYXVkIjoiYWNjb3VudCIsInN1YiI6ImRkM2FmZDkzLThkYjgtNDA2Ni04MTAyLWNhZDM4MDc5MWVkNCIsInR5cCI6IkJlYXJlciIsImF6cCI6IllacmNtTjFhYUNZU2N1cnJVakx2VlhaNWhKIiwiYWNyIjoiMSIsInJlYWxtX2FjY2VzcyI6eyJyb2xlcyI6WyJkZWZhdWx0LXJvbGVzLW1hc3RlciIsIm9mZmxpbmVfYWNjZXNzIiwidW1hX2F1dGhvcml6YXRpb24iXX0sInJlc291cmNlX2FjY2VzcyI6eyJhY2NvdW50Ijp7InJvbGVzIjpbIm1hbmFnZS1hY2NvdW50IiwibWFuYWdlLWFjY291bnQtbGlua3MiLCJ2aWV3LXByb2ZpbGUiXX19LCJzY29wZSI6ImVtYWlsIHByb2ZpbGUiLCJjbGllbnRIb3N0IjoiMTAuOTkuNC4xMzUiLCJlbWFpbF92ZXJpZmllZCI6ZmFsc2UsInByZWZlcnJlZF91c2VybmFtZSI6InNlcnZpY2UtYWNjb3VudC15enJjbW4xYWFjeXNjdXJydWpsdnZ4ejVoaiIsImNsaWVudEFkZHJlc3MiOiIxMC45OS40LjEzNSIsImNsaWVudF9pZCI6IllacmNtTjFhYUNZU2N1cnJVakx2VlhaNWhKIn0.AOZG2TNth26Hh0HYV0NWyGBsvnl-pK7TvI0HxXzbUtQhTikqwztf7kT3QhSSpMLqZWUlPsUW0f5Z7xZrGw64pO9lnX1DWqOq3qLP3690P-9_hiZzOx9o1tZ_x--OIOxfVvlgXB43AUTijycQaRdMxtUeqGhzJsRjbyLRk4_RHGHc76-EuXx6MxnjetuFxuTVtIUBK-XPOJLJ5ZSFeIDgaJa1AJ_0gChbihS3ErzGQiAgrK45rtVUoD74y-iIh6fDH_qdxMZsysoGEhFxz8vTukZin3eLgLWYSY_TjD64v2U3Qz75MYqrR0ga9vpWfekkrg4AClloQLcMokLPeThsGw"
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", auth))
	//if auth != "" {
	//	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", auth))
	//}
	printRequest(req)
	retries := 3
	body := []byte{}
	retcode := http.StatusOK
	for try := 1; try <= retries; try++ {
		client := &http.Client{Timeout: 60 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			if try == retries {
				return http.StatusInternalServerError, nil,
					errors.New("error conencting to  api service")
			}
			time.Sleep(5 * time.Second)
			continue
		}
		defer resp.Body.Close()
		body, _ = io.ReadAll(resp.Body)
		retcode = resp.StatusCode
		break
	}
	return retcode, body, nil
}

// MakeDeleteAPICall :
func MakeDeleteAPICall(ctx context.Context, connURL string, auth string, payload []byte) (int, []byte, error) {
	req, err := http.NewRequest("DELETE", connURL, bytes.NewBuffer(payload))
	if err != nil {
		return http.StatusInternalServerError, nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if auth != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", auth))
	}
	printRequest(req)
	retries := 3
	body := []byte{}
	retcode := http.StatusOK
	for try := 1; try <= retries; try++ {
		client := &http.Client{Timeout: 60 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			if try == retries {
				return http.StatusInternalServerError, nil,
					errors.New("error conencting to  api service")
			}
			time.Sleep(5 * time.Second)
			continue
		}
		defer resp.Body.Close()
		body, _ = io.ReadAll(resp.Body)
		retcode = resp.StatusCode
		break
	}
	return retcode, body, nil
}

// MakePutAPICall :
func MakePutAPICall(ctx context.Context, connURL, auth string, payload []byte) (int, []byte, error) {
	req, err := http.NewRequest("PUT", connURL, bytes.NewBuffer(payload))
	if err != nil {
		return http.StatusInternalServerError, nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if auth != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", auth))
	}
	printRequest(req)
	retries := 3
	body := []byte{}
	retcode := http.StatusOK
	for try := 1; try <= retries; try++ {
		client := &http.Client{Timeout: 60 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			if try == retries {
				return http.StatusInternalServerError, nil,
					errors.New("error conencting to  api service")
			}
			time.Sleep(5 * time.Second)
			continue
		}
		defer resp.Body.Close()
		body, _ = io.ReadAll(resp.Body)
		retcode = resp.StatusCode
		break
	}
	return retcode, body, nil
}

func printRequest(req *http.Request) {
	fmt.Printf("RK=>Method: %s\nURL: %s\nHeaders: %v\n", req.Method, req.URL.String(), req.Header)

	if req.Body != nil {
		bodyBytes, _ := io.ReadAll(req.Body)
		fmt.Printf("Body: %s\n", string(bodyBytes))

		// Recreate the body since io.ReadAll drains it
		req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}
}

type apiClientImpl struct{}

// NewAPIClient returns a concrete implementation of the APIClient interface.
func NewAPIClient() APIClient {
	return &apiClientImpl{}
}

func (c *apiClientImpl) MakeGetAPICall(ctx context.Context, url, token string, headers map[string]string) (int, []byte, error) {
	return MakeGetAPICall(ctx, url, token, nil)
}

func (c *apiClientImpl) MakePOSTAPICall(ctx context.Context, url, token string, payload []byte) (int, []byte, error) {
	return MakePOSTAPICall(ctx, url, token, payload)
}

func (c *apiClientImpl) MakePutAPICall(ctx context.Context, url, token string, payload []byte) (int, []byte, error) {
	return MakePutAPICall(ctx, url, token, payload)
}

func (c *apiClientImpl) MakeDeleteAPICall(ctx context.Context, url, token string, headers map[string]string) (int, []byte, error) {
	return MakeDeleteAPICall(ctx, url, token, nil)
}

func (c *apiClientImpl) GenerateFilesystemLoginCredentials(ctx context.Context, resourceId string) (*string, error) {
	// Placeholder: implement your actual logic here
	return nil, fmt.Errorf("GenerateFilesystemLoginCredentials not implemented")
}

func (c *apiClientImpl) ParseString(tmpl string, data any) (string, error) {
	// Placeholder: implement your actual logic here
	return ParseString(tmpl, data)
}
