// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
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

// MakeGetAPICall :
func MakeGetAPICall(ctx context.Context, connURL, auth string, payload []byte) (int, []byte, error) {
	// logger := log.FromContext(ctx).WithName("common.MakeGetAPICall")
	// http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: false}
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
	// body = []byte(sampleFilesystemList)
	return retcode, body, nil
}

// MakePOSTAPICall :
func MakePOSTAPICall(ctx context.Context, connURL, auth string, payload []byte) (int, []byte, error) {

	// http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: false}
	req, err := http.NewRequest("POST", connURL, bytes.NewBuffer(payload))
	if err != nil {
		return http.StatusInternalServerError, nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
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
	// logger := log.FromContext(ctx).WithName("common.MakeGetAPICall")
	// http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: false}
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
	// body = []byte(sampleFilesystemList)
	return retcode, body, nil
}

func printRequest(req *http.Request) {
	fmt.Printf("Method: %s\nURL: %s\nHeaders: %v\n", req.Method, req.URL.String(), req.Header)
	//fmt.Printf("nHeaders: %v\n", req.Header)

	if req.Body != nil {
		bodyBytes, _ := io.ReadAll(req.Body)
		fmt.Printf("Body: %s\n", string(bodyBytes))

		// Recreate the body since io.ReadAll drains it
		req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}
}
