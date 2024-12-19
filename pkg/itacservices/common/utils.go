package common

import (
	"bytes"
	"fmt"
	"net/http"
	"text/template"
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

func MapHttpError(code int) error {
	switch code {
	case http.StatusUnauthorized:
		return fmt.Errorf("Unauthorized")
	case http.StatusBadRequest:
		return fmt.Errorf("Bad Request")
	case http.StatusInternalServerError:
		return fmt.Errorf("Internal Server Error")
	default:
		return fmt.Errorf("error calling API")
	}
}
