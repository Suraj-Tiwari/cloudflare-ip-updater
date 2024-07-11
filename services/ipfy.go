package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type IpifyResponse struct {
	IP string `json:"ip"`
}

func GetIPAddress() (string, error) {
	request, err := http.NewRequest("GET", "https://api.ipify.org?format=json", nil)
	if err != nil {
		return "", err
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return "", err
	}
	defer func() {
		if closeErr := response.Body.Close(); closeErr != nil {
			fmt.Printf("Error closing response body: %v\n", closeErr)
		}
	}()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	var ipifyResponse IpifyResponse
	err = json.Unmarshal(body, &ipifyResponse)
	if err != nil {
		return "", err
	}

	return ipifyResponse.IP, nil
}
