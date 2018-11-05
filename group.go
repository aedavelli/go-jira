package jira

import (
	"errors"
	"fmt"
	"net/url"
)

// GroupService handles Groups for the JIRA instance / API.
//
// JIRA API docs: https://docs.atlassian.com/jira/REST/server/#api/2/group
type GroupService struct {
	client *Client
}

// groupMembersResult is only a small wrapper around the Group* methods
// to be able to parse the results
type groupMembersResult struct {
	StartAt    int           `json:"startAt"`
	MaxResults int           `json:"maxResults"`
	Total      int           `json:"total"`
	Members    []GroupMember `json:"values"`
}

// Group represents a JIRA group
type Group struct {
	ID                   string          `json:"id"`
	Title                string          `json:"title"`
	Type                 string          `json:"type"`
	Properties           groupProperties `json:"properties"`
	AdditionalProperties bool            `json:"additionalProperties"`
}

type groupProperties struct {
	Name groupPropertiesName `json:"name"`
}

type groupPropertiesName struct {
	Type string `json:"type"`
}

// GroupMember reflects a single member of a group
type GroupMember struct {
	Self         string `json:"self,omitempty"`
	Name         string `json:"name,omitempty"`
	Key          string `json:"key,omitempty"`
	EmailAddress string `json:"emailAddress,omitempty"`
	DisplayName  string `json:"displayName,omitempty"`
	Active       bool   `json:"active,omitempty"`
	TimeZone     string `json:"timeZone,omitempty"`
}

// GroupSearchOptions specifies the optional parameters for the Get Group methods
type GroupSearchOptions struct {
	StartAt              int64
	MaxResults           int32
	IncludeInactiveUsers bool
}

type GroupLabel struct {
	Text  string `json:"text"`
	Title string `json:"title"`
	Type  string `json:"type"`
}
type GroupDetails struct {
	Name   string       `json:"name"`
	Html   string       `json:"html"`
	Labels []GroupLabel `json:"labels"`
}
type GroupList struct {
	Header string         `json:"header"`
	Total  int32          `json:"total"`
	Groups []GroupDetails `json:"groups"`
}

// Get returns a paginated list of users who are members of the specified group and its subgroups.
// Users in the page are ordered by user names.
// User of this resource is required to have sysadmin or admin permissions.
//
// JIRA API docs: https://developer.atlassian.com/cloud/jira/platform/rest/v3/#api-api-3-group-member-get
//
// WARNING: This API only returns the first page of group members
func (s *GroupService) Get(name string) ([]GroupMember, *Response, error) {
	return s.GetWithOptions(name, nil)
}

// GetWithOptions returns a paginated list of members of the specified group and its subgroups.
// Users in the page are ordered by user names.
// User of this resource is required to have sysadmin or admin permissions.
//
// JIRA API docs: https://developer.atlassian.com/cloud/jira/platform/rest/v3/#api-api-3-group-member-get
func (s *GroupService) GetWithOptions(name string, options *GroupSearchOptions) ([]GroupMember, *Response, error) {
	var apiEndpoint string
	if options == nil {
		apiEndpoint = fmt.Sprintf("%s/group/member?groupname=%s", restAPIBase, url.QueryEscape(name))
	} else {
		apiEndpoint = fmt.Sprintf(
			"%s/group/member?groupname=%s&startAt=%d&maxResults=%d&includeInactiveUsers=%t",
			restAPIBase,
			url.QueryEscape(name),
			options.StartAt,
			options.MaxResults,
			options.IncludeInactiveUsers,
		)
	}
	req, err := s.client.NewRequest("GET", apiEndpoint, nil)
	if err != nil {
		return nil, nil, err
	}

	group := new(groupMembersResult)
	resp, err := s.client.Do(req, group)
	if err != nil {
		return nil, resp, err
	}
	return group.Members, resp, nil
}

// Add adds user to group
//
// JIRA API docs: https://developer.atlassian.com/cloud/jira/platform/rest/v3/#api-api-3-group-user-post
func (s *GroupService) Add(groupname string, userParams ...string) (*Group, *Response, error) {
	if len(userParams) != 1 && len(userParams) != 2 {
		// First string is username and second string is accountId
		return nil, nil, errors.New("Invalid User add parameters")
	}

	apiEndpoint := fmt.Sprintf("%s/group/user?groupname=%s", restAPIBase, url.QueryEscape(groupname))
	var user struct {
		Name      string `json:"name"`
		AccountId string `json:"accountId,omitempty"`
	}

	user.Name = userParams[0]
	if len(userParams) == 2 {
		user.AccountId = userParams[1]
	}

	req, err := s.client.NewRequest("POST", apiEndpoint, &user)
	if err != nil {
		return nil, nil, err
	}

	if len(userParams) == 2 {
		req.Header.Add("force-account-id", "true")
	}

	responseGroup := new(Group)
	resp, err := s.client.Do(req, responseGroup)
	if err != nil {
		jerr := NewJiraError(resp, err)
		return nil, resp, jerr
	}

	return responseGroup, resp, nil
}

// Remove removes user from group
//
// JIRA API docs: https://docs.atlassian.com/jira/REST/cloud/#api/2/group-removeUserFromGroup
func (s *GroupService) Remove(groupname string, username string) (*Response, error) {
	apiEndpoint := fmt.Sprintf("/rest/api/2/group/user?groupname=%s&username=%s",
		url.QueryEscape(groupname), url.QueryEscape(username))
	req, err := s.client.NewRequest("DELETE", apiEndpoint, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req, nil)
	if err != nil {
		jerr := NewJiraError(resp, err)
		return resp, jerr
	}

	return resp, nil
}

func (s *GroupService) GetList() (*GroupList, *Response, error) {
	return s.GetListWithOptions(nil)
}

func (s *GroupService) GetListWithOptions(v url.Values) (*GroupList, *Response, error) {
	apiEndPoint := fmt.Sprintf("%s/groups/picker", restAPIBase)
	if len(v) > 0 {
		apiEndPoint = fmt.Sprintf("%s?%s", apiEndPoint, v.Encode())
	}
	req, err := s.client.NewRequest("GET", apiEndPoint, nil)
	if err != nil {
		return nil, nil, err
	}

	gl := new(GroupList)
	resp, err := s.client.Do(req, gl)
	if err != nil {
		jerr := NewJiraError(resp, err)
		return nil, resp, jerr
	}

	return gl, resp, nil
}
