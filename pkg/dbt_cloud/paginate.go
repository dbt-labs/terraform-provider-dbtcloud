package dbt_cloud

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
)

type Response struct {
	Data  []any `json:"data"`
	Extra Extra `json:"extra"`
}

type Extra struct {
	Pagination Pagination `json:"pagination"`
}

type Pagination struct {
	Count      int `json:"count"`
	TotalCount int `json:"total_count"`
}

var log = logrus.New()

func (c *Client) GetEndpoint(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalf("Error creating a new request: %v", err)
	}

	resp, err := c.doRequest(req)
	if err != nil {
		log.Fatalf("Error fetching URL %v: %v", url, err)
	}

	return resp, err
}

func (c *Client) GetData(url string) []any {

	// get the first page
	jsonPayload, err := c.GetEndpoint(url)
	if err != nil {
		log.Fatal(err)
	}

	var response Response

	err = json.Unmarshal(jsonPayload, &response)
	if err != nil {
		log.Fatal(err)
	}

	allResponses := response.Data

	count := response.Extra.Pagination.Count
	for count < response.Extra.Pagination.TotalCount {
		// get the next page

		var newURL string
		lastPartURL, _ := lo.Last(strings.Split(url, "/"))
		if strings.Contains(lastPartURL, "?") {
			newURL = fmt.Sprintf("%s&offset=%d", url, count)
		} else {
			newURL = fmt.Sprintf("%s?offset=%d", url, count)
		}

		jsonPayload, err := c.GetEndpoint(newURL)
		if err != nil {
			log.Fatal(err)
		}
		var response Response

		err = json.Unmarshal(jsonPayload, &response)
		if err != nil {
			log.Fatal(err)
		}

		if response.Extra.Pagination.Count == 0 {
			// Unlucky! one object might have been deleted since the first call
			// if we don't stop here we will loop forever!
			break
		} else {
			count += response.Extra.Pagination.Count
		}
		allResponses = append(allResponses, response.Data...)
	}

	return allResponses
}

func (c *Client) GetAllGroupIDsByName(groupName string) []int {
	url := fmt.Sprintf("%s/v3/accounts/%d/groups/", c.HostURL, c.AccountID)

	allGroupsRaw := c.GetData(url)

	return lo.FilterMap(allGroupsRaw, func(group any, _ int) (int, bool) {
		if group.(map[string]any)["name"].(string) == groupName {
			return int(group.(map[string]any)["id"].(float64)), true
		}
		return 0, false
	})
}

func (c *Client) GetAllEnvironments(projectID int) ([]Environment, error) {
	url := fmt.Sprintf("%s/v3/accounts/%d/environments/", c.HostURL, c.AccountID)

	if projectID != 0 {
		url = fmt.Sprintf("%s?project_id=%d", url, projectID)
	}

	allEnvironmentsRaw := c.GetData(url)

	allEnvs := []Environment{}
	for _, env := range allEnvironmentsRaw {

		data, _ := json.Marshal(env)
		currentEnv := Environment{}
		err := json.Unmarshal(data, &currentEnv)
		if err != nil {
			return nil, err
		}
		allEnvs = append(allEnvs, currentEnv)
	}
	return allEnvs, nil
}

func (c *Client) GetAllNotifications() ([]Notification, error) {
	url := fmt.Sprintf("%s/v2/accounts/%d/notifications/", c.HostURL, c.AccountID)

	allNotificationsRaw := c.GetData(url)

	allNotifications := []Notification{}
	for _, notification := range allNotificationsRaw {

		data, _ := json.Marshal(notification)
		currentNotification := Notification{}
		err := json.Unmarshal(data, &currentNotification)
		if err != nil {
			return nil, err
		}
		allNotifications = append(allNotifications, currentNotification)
	}
	return allNotifications, nil
}
