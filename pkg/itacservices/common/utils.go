package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"text/template"
	"time"
)

type APIError struct {
	Code    int           `json:"code"`
	Message string        `json:"message"`
	Details []interface{} `json:"details"`
}

const (
	DefaultRetryInterval = time.Duration(5 * time.Second)
)

// ParseString parses the given template string with the provided data.
func ParseString(templateString string, data interface{}) (string, error) {
	t, err := template.New("generic").Parse(templateString)
	if err != nil {
		return "", err
	}

	var result bytes.Buffer
	err = t.Execute(&result, data)
	if err != nil {
		return "", err
	}

	return result.String(), nil
}

func MapHttpError(code int, retval []byte) error {
	switch code {
	case http.StatusUnauthorized:
		return fmt.Errorf("Unauthorized")
	case http.StatusBadRequest:
		return fmt.Errorf("Bad Request, message: %v", mapAPIErrorMessage(retval))
	case http.StatusInternalServerError:
		return fmt.Errorf("Internal Server Error, message: %v", mapAPIErrorMessage(retval))
	default:
		return fmt.Errorf("error calling API, message: %v", mapAPIErrorMessage(retval))
	}
}

func mapAPIErrorMessage(retval []byte) error {
	apiError := APIError{}
	if err := json.Unmarshal(retval, &apiError); err != nil {
		return fmt.Errorf("error parsing iks error response")
	}
	return fmt.Errorf("%v", apiError.Message)
}

func GetAvailabiltyZoneAndVnet(region string) (string, string) {
	availabilityZone := fmt.Sprintf("%sa", region)
	vnetName := fmt.Sprintf("%s-default", availabilityZone)

	return availabilityZone, vnetName

}
