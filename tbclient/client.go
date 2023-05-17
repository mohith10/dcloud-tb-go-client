/**
Copyright (c) 2023 Cisco Systems, Inc. and its affiliates
*/

package tbclient

import (
	"fmt"
	"github.com/go-resty/resty/v2"
)

// HostURL Defaulted to local service instance
const HostURL = "http://localhost:8080"
const etagHeader = "ETag"
const ifMatch = "IF-MATCH"

type Client struct {
	HostURL     string
	Token       string
	Debug       bool
	UserAgent   string
	DisableGzip bool
}

func NewClient(host, authToken *string) *Client {
	c := Client{
		HostURL: HostURL,
		Token:   *authToken,
	}

	if host != nil {
		c.HostURL = *host
	}

	return &c
}

type resourceService[R any, RC embeddedData[R]] struct {
	collectionService[R, RC]
}

type collectionService[R any, RC embeddedData[R]] struct {
	client         *Client
	resourcePath   string
	topologyUid    string
	requestHeaders map[string]string
}

func (s *collectionService[R, RC]) getAll() ([]R, error) {

	topologyPath := getTopologyPath(s.topologyUid)

	rest := s.createRestClient()
	url := fmt.Sprintf("%s%s%s", s.client.HostURL, topologyPath, s.resourcePath)

	resp, err := executeGet(rest, s.client.Token, url, collection[RC]{})
	if err != nil {
		return nil, err
	}

	return resp.Result().(*collection[RC]).Embedded.getData(), nil
}

func (s *resourceService[R, RC]) getOne(uid string) (*R, error) {

	rest := s.createRestClient()
	url := singleResourceUrl(s.client.HostURL, s.resourcePath, uid)

	resp, err := executeGet(rest, s.client.Token, url, new(R))
	if err != nil {
		return nil, err
	}

	return resp.Result().(*R), nil
}

func (s *resourceService[R, RC]) create(resource R) (*R, error) {

	rest := s.createRestClient()
	rest.SetHeader("Content-Type", "application/json")

	url := fmt.Sprintf("%s%s", s.client.HostURL, s.resourcePath)

	resp, err := rest.R().
		SetBody(resource).
		SetResult(new(R)).
		SetError([]vndError{}).
		SetAuthToken(s.client.Token).
		Post(url)

	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, generateError(*resp)
	}

	return resp.Result().(*R), nil
}

func (s *resourceService[R, RC]) update(uid string, resource R) (*R, error) {

	rest := s.createRestClient()
	rest.SetHeader("Content-Type", "application/json")

	url := singleResourceUrl(s.client.HostURL, s.resourcePath, uid)

	current, err := executeGet(rest, s.client.Token, url, new(R))

	if err != nil {
		return nil, err
	}

	etag := current.Header().Get(etagHeader)

	updated, err := executePut(rest, s.client.Token, url, resource, new(R), etag)

	if err != nil {
		return nil, err
	}

	return updated.Result().(*R), nil
}

func (s *resourceService[R, RC]) delete(uid string) error {

	rest := s.createRestClient()

	url := singleResourceUrl(s.client.HostURL, s.resourcePath, uid)

	resp, err := rest.R().
		SetError([]vndError{}).
		SetAuthToken(s.client.Token).
		Delete(url)

	if err != nil {
		return err
	}

	if resp.IsError() && resp.StatusCode() != 404 {
		return generateError(*resp)
	}

	return nil
}

func singleResourceUrl(hostUrl, resourcePath, uid string) string {
	singlePath := fmt.Sprintf("%s/%s", resourcePath, uid)
	url := fmt.Sprintf("%s%s", hostUrl, singlePath)
	return url
}

func (s *collectionService[R, RC]) createRestClient() *resty.Client {
	rest := s.client.createRestClient()
	rest.SetHeaders(s.requestHeaders)
	return rest
}

func (c *Client) createRestClient() *resty.Client {
	rest := resty.New()
	rest.SetDebug(c.Debug)
	if c.UserAgent != "" {
		rest.SetHeader("User-Agent", c.UserAgent)
	}
	if c.DisableGzip {
		rest.SetHeader("Accept-Encoding", "")
	}
	return rest
}

func getTopologyPath(topologyUid string) string {
	if topologyUid != "" {
		return fmt.Sprintf("/topologies/%s", topologyUid)
	} else {
		return ""
	}
}

func executeGet(rest *resty.Client, authToken string, url string, result interface{}) (*resty.Response, error) {

	resp, err := rest.R().
		SetResult(result).
		SetError([]vndError{}).
		SetAuthToken(authToken).
		Get(url)

	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, generateError(*resp)
	}
	return resp, nil
}

func executePut(rest *resty.Client, authToken string, url string, body interface{}, result interface{}, etag string) (*resty.Response, error) {
	resp, err := rest.R().
		SetHeader(ifMatch, etag).
		SetBody(body).
		SetResult(result).
		SetError([]vndError{}).
		SetAuthToken(authToken).
		Put(url)

	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, generateError(*resp)
	}
	return resp, nil
}

var emptyHeaders = map[string]string{}
