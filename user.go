package jira

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
)

// UserService handles users for the JIRA instance / API.
//
// JIRA API docs: https://docs.atlassian.com/jira/REST/cloud/#api/2/user
type UserService struct {
	client *Client
}

// User represents a JIRA user.
type User struct {
	Self            string     `json:"self,omitempty" structs:"self,omitempty"`
	Name            string     `json:"name,omitempty" structs:"name,omitempty"`
	Password        string     `json:"-"`
	Key             string     `json:"key,omitempty" structs:"key,omitempty"`
	EmailAddress    string     `json:"emailAddress,omitempty" structs:"emailAddress,omitempty"`
	AvatarUrls      AvatarUrls `json:"avatarUrls,omitempty" structs:"avatarUrls,omitempty"`
	DisplayName     string     `json:"displayName,omitempty" structs:"displayName,omitempty"`
	Active          bool       `json:"active,omitempty" structs:"active,omitempty"`
	Notification    bool       `json:"notification ,omitempty" structs:"notification ,omitempty"`
	TimeZone        string     `json:"timeZone,omitempty" structs:"timeZone,omitempty"`
	Groups          UserGroups `json:"groups,omitempty" structs:"groups,omitempty"`
	ApplicationKeys []string   `json:"applicationKeys,omitempty" structs:"applicationKeys,omitempty"`
}

// UserGroup represents the group list
type UserGroup struct {
	Self string `json:"self,omitempty" structs:"self,omitempty"`
	Name string `json:"name,omitempty" structs:"name,omitempty"`
}

type UserGroups struct {
	Size  int         `json:"size,omitempty" structs:"size,omitempty"`
	Items []UserGroup `json:"items,omitempty" structs:"items,omitempty"`
}

type userSearchParam struct {
	name  string
	value string
}

type userSearch []userSearchParam

type userSearchF func(userSearch) userSearch

// Get gets user info from JIRA
//
// JIRA API docs: https://docs.atlassian.com/jira/REST/cloud/#api/2/user-getUser
func (s *UserService) Get(username string) (*User, *Response, error) {
	qp := url.Values{}
	qp["username"] = []string{username}
	return s.GetWithQueryParams(qp)
}

// Create creates an user in JIRA.
//
// JIRA API docs: https://docs.atlassian.com/jira/REST/cloud/#api/2/user-createUser
func (s *UserService) Create(user *User) (*User, *Response, error) {
	const apiEndpoint = restAPIBase + "/user"
	req, err := s.client.NewRequest("POST", apiEndpoint, user)
	if err != nil {
		return nil, nil, err
	}

	resp, err := s.client.Do(req, nil)
	if err != nil {
		return nil, resp, err
	}

	responseUser := new(User)
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		e := fmt.Errorf("Could not read the returned data")
		return nil, resp, NewJiraError(resp, e)
	}
	err = json.Unmarshal(data, responseUser)
	if err != nil {
		e := fmt.Errorf("Could not unmarshall the data into struct")
		return nil, resp, NewJiraError(resp, e)
	}
	return responseUser, resp, nil
}

// Delete deletes an user from JIRA.
// Returns http.StatusNoContent on success.
//
// JIRA API docs: https://developer.atlassian.com/cloud/jira/platform/rest/#api-api-2-user-delete
func (s *UserService) Delete(username string) (*Response, error) {
	apiEndpoint := fmt.Sprintf("/rest/api/2/user?username=%s", url.QueryEscape(username))
	req, err := s.client.NewRequest("DELETE", apiEndpoint, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req, nil)
	if err != nil {
		return resp, NewJiraError(resp, err)
	}
	return resp, nil
}

// GetGroups returns the groups which the user belongs to
//
// JIRA API docs: https://docs.atlassian.com/jira/REST/cloud/#api/2/user-getUserGroups
func (s *UserService) GetGroups(username string) (*[]UserGroup, *Response, error) {
	apiEndpoint := fmt.Sprintf("/rest/api/2/user/groups?username=%s", url.QueryEscape(username))
	req, err := s.client.NewRequest("GET", apiEndpoint, nil)
	if err != nil {
		return nil, nil, err
	}

	userGroups := new([]UserGroup)
	resp, err := s.client.Do(req, userGroups)
	if err != nil {
		return nil, resp, NewJiraError(resp, err)
	}
	return userGroups, resp, nil
}

// Get information about the current logged-in user
//
// JIRA API docs: https://developer.atlassian.com/cloud/jira/platform/rest/#api-api-2-myself-get
func (s *UserService) GetSelf() (*User, *Response, error) {
	const apiEndpoint = "rest/api/2/myself"
	req, err := s.client.NewRequest("GET", apiEndpoint, nil)
	if err != nil {
		return nil, nil, err
	}
	var user User
	resp, err := s.client.Do(req, &user)
	if err != nil {
		return nil, resp, NewJiraError(resp, err)
	}
	return &user, resp, nil
}

// WithMaxResults sets the max results to return
func WithMaxResults(maxResults int) userSearchF {
	return func(s userSearch) userSearch {
		s = append(s, userSearchParam{name: "maxResults", value: fmt.Sprintf("%d", maxResults)})
		return s
	}
}

// WithStartAt set the start pager
func WithStartAt(startAt int) userSearchF {
	return func(s userSearch) userSearch {
		s = append(s, userSearchParam{name: "startAt", value: fmt.Sprintf("%d", startAt)})
		return s
	}
}

// WithActive sets the active users lookup
func WithActive(active bool) userSearchF {
	return func(s userSearch) userSearch {
		s = append(s, userSearchParam{name: "includeActive", value: fmt.Sprintf("%t", active)})
		return s
	}
}

// WithInactive sets the inactive users lookup
func WithInactive(inactive bool) userSearchF {
	return func(s userSearch) userSearch {
		s = append(s, userSearchParam{name: "includeInactive", value: fmt.Sprintf("%t", inactive)})
		return s
	}
}

// WithQuery sets the query string
func WithQuery(query string) userSearchF {
	return func(s userSearch) userSearch {
		s = append(s, userSearchParam{name: "query", value: fmt.Sprintf("%s", query)})
		return s
	}
}

// WithUsername sets the username
func WithUsername(username string) userSearchF {
	return func(s userSearch) userSearch {
		s = append(s, userSearchParam{name: "username", value: fmt.Sprintf("%s", url.QueryEscape(username))})
		return s
	}
}

// Find searches for user info from JIRA:
// It can find users by email, username or name
//
// JIRA API docs: https://docs.atlassian.com/jira/REST/cloud/#api/2/user-findUsers
func (s *UserService) Find(tweaks ...userSearchF) ([]User, *Response, error) {
	search := []userSearchParam{}
	for _, f := range tweaks {
		search = f(search)
	}

	var queryString = ""
	for _, param := range search {
		queryString += param.name + "=" + param.value + "&"
	}

	apiEndpoint := fmt.Sprintf("/rest/api/2/user/search?" + queryString[:len(queryString)-1])
	req, err := s.client.NewRequest("GET", apiEndpoint, nil)
	if err != nil {
		return nil, nil, err
	}

	users := []User{}
	resp, err := s.client.Do(req, &users)
	if err != nil {
		return nil, resp, NewJiraError(resp, err)
	}
	return users, resp, nil
}

// Returns a list of users that match the search string and property.
//
// https://developer.atlassian.com/cloud/jira/platform/rest/v3/#api-api-3-user-search-get
func (s *UserService) FindWithQueryParams(qp url.Values) ([]User, *Response, error) {
	apiEndpoint := restAPIBase + "/user/search"
	if len(qp) > 0 {
		apiEndpoint += "?" + qp.Encode()
	}
	req, err := s.client.NewRequest("GET", apiEndpoint, nil)
	if err != nil {
		return nil, nil, err
	}

	users := []User{}
	resp, err := s.client.Do(req, &users)
	if err != nil {
		return nil, resp, NewJiraError(resp, err)
	}
	return users, resp, nil
}

// Returns a user.
//
// https://developer.atlassian.com/cloud/jira/platform/rest/v3/#api-api-3-user-get
func (s *UserService) GetWithQueryParams(qp url.Values) (*User, *Response, error) {
	apiEndpoint := restAPIBase + "/user"
	if len(qp) > 0 {
		apiEndpoint += "?" + qp.Encode()
	}

	req, err := s.client.NewRequest("GET", apiEndpoint, nil)
	if err != nil {
		return nil, nil, err
	}

	user := new(User)
	resp, err := s.client.Do(req, user)
	if err != nil {
		return nil, resp, NewJiraError(resp, err)
	}
	return user, resp, nil
}
