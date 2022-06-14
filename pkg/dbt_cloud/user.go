package dbt_cloud

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

type User struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
}

type UserListResponse struct {
	Data   []User         `json:"data"`
	Status ResponseStatus `json:"status"`
}

func (c *Client) GetUser(email string) (*User, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/v2/accounts/%s/users/", c.HostURL, strconv.Itoa(c.AccountID)), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	userListResponse := UserListResponse{}
	err = json.Unmarshal(body, &userListResponse)
	if err != nil {
		return nil, err
	}

	for i, user := range userListResponse.Data {
		if user.Email == email {
			return &userListResponse.Data[i], nil
		}
	}

	return nil, fmt.Errorf("Did not find user with email %s", email)
}
