package dbt_cloud

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type User struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
	// only the first permission is filled id, it is a list with  1 element
	Permissions []struct {
		Groups []struct {
			ID int `json:"id"`
		} `json:"groups"`
	} `json:"permissions"`
}

type UserListResponse struct {
	Data   []User         `json:"data"`
	Status ResponseStatus `json:"status"`
	Extra  ResponseExtra  `json:"extra"`
}

type CurrentUser struct {
	User User `json:"user"`
}

type CurrentUserResponse struct {
	Data   CurrentUser    `json:"data"`
	Status ResponseStatus `json:"status"`
}

func (c *Client) GetUsers() ([]User, error) {
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf("%s/v3/accounts/%s/users/", c.HostURL, strconv.FormatInt(c.AccountID, 10)),
		nil,
	)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequestWithRetry(req)
	if err != nil {
		return nil, err
	}

	userListResponse := UserListResponse{}
	err = json.Unmarshal(body, &userListResponse)
	if err != nil {
		return nil, err
	}

	listAllUsers := userListResponse.Data

	// if there are more than the limit, we need to paginate
	if userListResponse.Extra.Pagination.TotalCount > userListResponse.Extra.Filters.Limit {
		numUsers := userListResponse.Extra.Pagination.Count
		for numUsers < userListResponse.Extra.Pagination.TotalCount {

			req, err := http.NewRequest(
				"GET",
				fmt.Sprintf(
					"%s/v3/accounts/%s/users/?offset=%d",
					c.HostURL,
					strconv.FormatInt(c.AccountID, 10),
					numUsers,
				),
				nil,
			)
			if err != nil {
				return nil, err
			}

			body, err := c.doRequestWithRetry(req)
			if err != nil {
				return nil, err
			}

			userListResponse := UserListResponse{}
			err = json.Unmarshal(body, &userListResponse)
			if err != nil {
				return nil, err
			}

			numUsersLastCall := userListResponse.Extra.Pagination.Count
			if numUsersLastCall > 0 {
				listAllUsers = append(listAllUsers, userListResponse.Data...)
				numUsers += userListResponse.Extra.Pagination.Count
			} else {
				// this means that most likely one item was deleted since the first call
				// so the number of items is less than the initial total, we can break the loop
				break
			}

		}
	}

	return listAllUsers, nil
}

func (c *Client) GetUser(email string) (*User, error) {

	listAllUsers, err := c.GetUsers()
	if err != nil {
		return nil, err
	}

	for i, user := range listAllUsers {
		if strings.EqualFold(user.Email, email) {
			return &listAllUsers[i], nil
		}
	}

	return nil, fmt.Errorf("did not find user with email %s", email)
}

func (c *Client) GetConnectedUser() (*User, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/v2/whoami/", c.HostURL), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequestWithRetry(req)
	if err != nil {
		return nil, err
	}

	currentUserResponse := CurrentUserResponse{}
	err = json.Unmarshal(body, &currentUserResponse)
	if err != nil {
		return nil, err
	}

	return &currentUserResponse.Data.User, nil
}
