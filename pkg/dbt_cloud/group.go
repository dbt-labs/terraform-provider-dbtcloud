package dbt_cloud

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type GroupPermission struct {
	ID          *int   `json:"id"`
	AccountID   int    `json:"account_id"`
	GroupID     int    `json:"group_id"`
	ProjectID   int    `json:"project_id"`
	AllProjects bool   `json:"all_projects"`
	State       int    `json:"state"`
	Set         string `json:"permission_set"`
	Level       string `json:"permission_level"`
}

type Group struct {
	ID               *int              `json:"id"`
	Name             string            `json:"name"`
	AccountID        int               `json:"account_id"`
	State            int               `json:"state"`
	AssignByDefault  bool              `json:"assign_by_default"`
	SSOMappingGroups []string          `json:"sso_mapping_groups"`
	Permissions      []GroupPermission `json:"group_permissions"`
}

type GroupResponse struct {
	Data   Group          `json:"data"`
	Status ResponseStatus `json:"status"`
}

type GroupListResponse struct {
	Data   []Group        `json:"data"`
	Status ResponseStatus `json:"status"`
}

func (c *Client) GetGroup(groupID int) (*Group, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/v3/accounts/%s/groups/", c.HostURL, strconv.Itoa(c.AccountID)), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	groupListResponse := GroupListResponse{}
	err = json.Unmarshal(body, &groupListResponse)
	if err != nil {
		return nil, err
	}

	for i, group := range groupListResponse.Data {
		if *group.ID == groupID {
			return &groupListResponse.Data[i], nil
		}
	}

	return nil, fmt.Errorf("Did not find group with ID %d", groupID)
}

func (c *Client) CreateGroup(name string, assignByDefault bool, ssoMappingGroups []string) (*Group, error) {
	newGroup := Group{
		AccountID:        c.AccountID,
		Name:             name,
		AssignByDefault:  assignByDefault,
		SSOMappingGroups: ssoMappingGroups,
		State:            STATE_ACTIVE, // TODO: make variable
	}
	newGroupData, err := json.Marshal(newGroup)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/v3/accounts/%d/groups/", c.HostURL, c.AccountID), strings.NewReader(string(newGroupData)))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	groupResponse := GroupResponse{}
	err = json.Unmarshal(body, &groupResponse)
	if err != nil {
		return nil, err
	}

	return &groupResponse.Data, nil
}

func (c *Client) UpdateGroup(groupID int, group Group) (*Group, error) {
	groupData, err := json.Marshal(group)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/v3/accounts/%s/groups/%s/", c.HostURL, strconv.Itoa(c.AccountID), groupID), strings.NewReader(string(groupData)))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	groupResponse := GroupResponse{}
	err = json.Unmarshal(body, &groupResponse)
	if err != nil {
		return nil, err
	}

	return &groupResponse.Data, nil
}
