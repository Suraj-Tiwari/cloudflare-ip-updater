package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type DnsRecord struct {
	Content   string `json:"content"`
	Name      string `json:"name"`
	Proxied   bool   `json:"proxied"`
	Type      string `json:"type"`
	Comment   string `json:"comment"`
	CreatedOn string `json:"created_on"`
	ID        string `json:"id"`
	Locked    bool   `json:"locked"`
	Meta      struct {
		AutoAdded bool   `json:"auto_added"`
		Source    string `json:"source"`
	} `json:"meta"`
	ModifiedOn string   `json:"modified_on"`
	Proxiable  bool     `json:"proxiable"`
	Tags       []string `json:"tags"`
	TTL        int      `json:"ttl"`
	ZoneID     string   `json:"zone_id"`
	ZoneName   string   `json:"zone_name"`
}

type GetDNSApiResponse struct {
	Errors     []string            `json:"errors"`
	Messages   []string            `json:"messages"`
	Result     []DnsRecord         `json:"result"`
	Success    bool                `json:"success"`
	ResultInfo GetDNSApiResultInfo `json:"result_info"`
}

type GetDNSApiResultInfo struct {
	Count      int `json:"count"`
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	TotalCount int `json:"total_count"`
}

type CloudFlareService struct {
	zoneIdentifier string
	authKey        string
	filter         string
}

type UpdateDNSRecordRequest struct {
	Content string `json:"content"`
}

type UpdateDNSRecordResponse struct {
	Success bool                   `json:"success"`
	Errors  []string               `json:"errors"`
	Result  UpdateDNSRecordRequest `json:"result"`
}

func (s *CloudFlareService) GetDnsRecords() ([]DnsRecord, error) {
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records?comment.contains=%s", s.zoneIdentifier, s.filter)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return []DnsRecord{}, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", s.authKey))
	//Authorization:

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return []DnsRecord{}, fmt.Errorf("error making request: %v", err)
	}
	defer func() {
		if closeErr := res.Body.Close(); closeErr != nil {
			fmt.Printf("Error closing response body: %v\n", closeErr)
		}
	}()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return []DnsRecord{}, fmt.Errorf("error reading response body: %v", err)
	}

	var apiResponse GetDNSApiResponse
	err = json.Unmarshal(body, &apiResponse)
	if err != nil {
		return []DnsRecord{}, fmt.Errorf("error decoding JSON: %v", err)
	}

	return apiResponse.Result, nil
}

func (s *CloudFlareService) UpdateDnsRecord(records []DnsRecord, ip string) error {
	for _, record := range records {
		err := s.updateDnsRecordRequest(record, ip)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *CloudFlareService) updateDnsRecordRequest(record DnsRecord, ip string) error {
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records/%s", s.zoneIdentifier, record.ID)
	payload, err := json.Marshal(UpdateDNSRecordRequest{
		Content: ip,
	})
	fmt.Printf("updating record %s with new ip %s\n", record.Name, ip)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON payload: %v", err)
	}
	req, err := http.NewRequest("PATCH", url, strings.NewReader(string(payload)))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %v", err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", s.authKey))
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to perform HTTP request: %v", err)
	}
	defer func() {
		if closeErr := response.Body.Close(); closeErr != nil {
			fmt.Printf("Error closing response body: %v\n", closeErr)
		}
	}()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %v", err)
	}

	var apiResponse UpdateDNSRecordResponse
	err = json.Unmarshal(body, &apiResponse)
	if err != nil {
		return fmt.Errorf("failed to unmarshal API response: %v", err)
	}

	if !apiResponse.Success {
		return fmt.Errorf("API request failed: %v", apiResponse.Errors)
	}
	fmt.Println("Updated DNS Record")

	return nil
}

func NewCloudflareService(zoneIdentifier string, authKeys string, filter string) *CloudFlareService {
	return &CloudFlareService{
		zoneIdentifier: zoneIdentifier,
		authKey:        authKeys,
		filter:         filter,
	}
}
