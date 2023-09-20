package dbt_cloud

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type GroupPermission struct {
	ID          *int   `json:"id,omitempty"`
	AccountID   int    `json:"account_id"`
	GroupID     int    `json:"group_id"`
	ProjectID   int    `json:"project_id,omitempty"`
	AllProjects bool   `json:"all_projects"`
	State       int    `json:"state,omitempty"`
	Set         string `json:"permission_set,omitempty"`
	Level       string `json:"permission_level,omitempty"`
}

type Group struct {
	ID               *int              `json:"id"`
	Name             string            `json:"name"`
	AccountID        int               `json:"account_id"`
	State            int               `json:"state"`
	AssignByDefault  bool              `json:"assign_by_default"`
	SSOMappingGroups []string          `json:"sso_mapping_groups"`
	Permissions      []GroupPermission `json:"group_permissions,omitempty"`
}

type GroupResponse struct {
	Data   Group          `json:"data"`
	Status ResponseStatus `json:"status"`
}

type GroupListResponse struct {
	Data   []Group        `json:"data"`
	Status ResponseStatus `json:"status"`
}

type GroupPermissionListResponse struct {
	Data   []GroupPermission `json:"data"`
	Status ResponseStatus    `json:"status"`
}

func (c *Client) GetGroup(groupID int) (*Group, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/v3/accounts/%s/groups/%s/", c.HostURL, strconv.Itoa(c.AccountID), strconv.Itoa(groupID)), nil)
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

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/v3/accounts/%s/groups/%d/", c.HostURL, strconv.Itoa(c.AccountID), groupID), strings.NewReader(string(groupData)))
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

func (c *Client) UpdateGroupPermissions(groupID int, groupPermissions []GroupPermission) (*[]GroupPermission, error) {
	groupPermissionData, err := json.Marshal(groupPermissions)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/v3/accounts/%s/group-permissions/%d/", c.HostURL, strconv.Itoa(c.AccountID), groupID), strings.NewReader(string(groupPermissionData)))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	groupPermissionResponse := GroupPermissionListResponse{}
	err = json.Unmarshal(body, &groupPermissionResponse)
	if err != nil {
		return nil, err
	}

	return &groupPermissionResponse.Data, nil
}
