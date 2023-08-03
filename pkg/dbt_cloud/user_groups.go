package dbt_cloud

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type UserGroupsResponse struct {
	Data   UserGroups     `json:"data"`
	Status ResponseStatus `json:"status"`
}

type UserGroups struct {
	Permissions []Permission `json:"permissions"`
}

type Permission struct {
	AccountID int     `json:"account_id"`
	Groups    []Group `json:"groups"`
}

type UserGroupsCurrentAccount struct {
	Groups []Group
}

type UserGroupsBody struct {
	UserID   int   `json:"user_id"`
	GroupIDs []int `json:"desired_group_ids"`
}

// Group is already defined in group.go

type AssignUserGroupsResponse struct {
	Data   []Group        `json:"data"`
	Status ResponseStatus `json:"status"`
}

// the 2 following types are used to parse groups from /users/
// they could be removed once we can call /users/{user_id}/
type UserWithGroups struct {
	ID     int          `json:"id"`
	Email  string       `json:"email"`
	Groups []Permission `json:"permissions"`
}

type UserListResponseWithGroups struct {
	Data   []UserWithGroups `json:"data"`
	Status ResponseStatus   `json:"status"`
}

func (c *Client) GetUserGroups(userId int) (*UserGroupsCurrentAccount, error) {

	// TODO: the current endpoint is not working for service tokens, we could use it when it is updated
	// in the meantime we use the /users/ endpoint and parse the groups from there

	// req, err := http.NewRequest("GET", fmt.Sprintf("%s/v2/accounts/%s/users/%s/", c.HostURL, strconv.Itoa(c.AccountID), strconv.Itoa(userId)), nil)
	// if err != nil {
	// 	return nil, err
	// }

	// body, err := c.doRequest(req)
	// if err != nil {
	// 	return nil, err
	// }

	// userGroupsResponse := UserGroupsResponse{}
	// err = json.Unmarshal(body, &userGroupsResponse)
	// if err != nil {
	// 	return nil, err
	// }

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/v3/accounts/%s/users/?limit=1000", c.HostURL, strconv.Itoa(c.AccountID)), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	userListResponse := UserListResponseWithGroups{}
	err = json.Unmarshal(body, &userListResponse)
	if err != nil {
		return nil, err
	}

	userGroupsPermissions := []Permission{}
	for i, user := range userListResponse.Data {
		if user.ID == userId {
			userGroupsPermissions = userListResponse.Data[i].Groups
		}
	}

	// the API returns the permissions for all accounts
	// we just want to get the ones of the current account
	userGroupsCurrentAccount := UserGroupsCurrentAccount{}
	for _, permission := range userGroupsPermissions {
		if permission.AccountID == c.AccountID {
			userGroupsCurrentAccount.Groups = append(userGroupsCurrentAccount.Groups, permission.Groups...)
		}
	}

	return &userGroupsCurrentAccount, nil
}

func (c *Client) AssignUserGroups(userId int, groupIDs []int) (*AssignUserGroupsResponse, error) {

	userGroupsBody := UserGroupsBody{
		UserID:   userId,
		GroupIDs: groupIDs,
	}

	userGroupsData, err := json.Marshal(userGroupsBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/v3/accounts/%s/assign-groups/", c.HostURL, strconv.Itoa(c.AccountID)), strings.NewReader(string(userGroupsData)))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	userGroupsResponse := AssignUserGroupsResponse{}
	err = json.Unmarshal(body, &userGroupsResponse)
	if err != nil {
		return nil, err
	}

	return &userGroupsResponse, nil
}
