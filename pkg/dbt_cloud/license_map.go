package dbt_cloud

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type LicenseMap struct {
	ID               *int     `json:"id"`
	LicenseType      string   `json:"license_type"`
	AccountID        int      `json:"account_id"`
	State            int      `json:"state"`
	SSOMappingGroups []string `json:"sso_mapping_groups"`
}

type LicenseMapResponse struct {
	Data   LicenseMap     `json:"data"`
	Status ResponseStatus `json:"status"`
}

func (c *Client) GetLicenseMap(licenseMapId int) (*LicenseMap, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/v3/accounts/%s/license-maps/%d/", c.HostURL, strconv.Itoa(c.AccountID), licenseMapId), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	licenseMapResponse := LicenseMapResponse{}
	err = json.Unmarshal(body, &licenseMapResponse)
	if err != nil {
		return nil, err
	}

	return &licenseMapResponse.Data, nil
}

func (c *Client) CreateLicenseMap(licenseType string, ssoMappingGroups []string) (*LicenseMap, error) {
	newLicenseMap := LicenseMap{
		AccountID:        c.AccountID,
		LicenseType:      licenseType,
		SSOMappingGroups: ssoMappingGroups,
		State:            STATE_ACTIVE,
	}
	newLicenseMapData, err := json.Marshal(newLicenseMap)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/v3/accounts/%d/license-maps/", c.HostURL, c.AccountID), strings.NewReader(string(newLicenseMapData)))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	licenseMapResponse := LicenseMapResponse{}
	err = json.Unmarshal(body, &licenseMapResponse)
	if err != nil {
		return nil, err
	}

	return &licenseMapResponse.Data, nil
}

func (c *Client) UpdateLicenseMap(licenseMapID int, licenseMap LicenseMap) (*LicenseMap, error) {
	licenseMapData, err := json.Marshal(licenseMap)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/v3/accounts/%s/license-maps/%d/", c.HostURL, strconv.Itoa(c.AccountID), licenseMapID), strings.NewReader(string(licenseMapData)))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	licenseMapResponse := LicenseMapResponse{}
	err = json.Unmarshal(body, &licenseMapResponse)
	if err != nil {
		return nil, err
	}

	return &licenseMapResponse.Data, nil
}

func (c *Client) DestroyLicenseMap(licenseMapID int) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/v3/accounts/%s/license-maps/%d/", c.HostURL, strconv.Itoa(c.AccountID), licenseMapID), nil)
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	if err != nil {
		return err
	}

	return nil
}
