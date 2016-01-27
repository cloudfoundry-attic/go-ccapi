package client

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/cloudfoundry/go-ccapi/v3/routing"
	"github.com/tedsuo/rata"
)

//go:generate counterfeiter -o fakes/fake_client.go . Client
type Client interface {
	GetApplications(queryParams url.Values) ([]byte, error)
	GetResource(path string) ([]byte, error)
	GetResources(path string, limit int) ([]byte, error)
	RefreshAuthToken() error
}

type client struct {
	ccRequestGenerator  *rata.RequestGenerator
	uaaRequestGenerator *rata.RequestGenerator
	httpClient          *http.Client
	ccEndpoint          string
	token               string
}

func NewClient(ccEndpoint, uaaEndpoint, token string) Client {
	return &client{
		ccRequestGenerator:  rata.NewRequestGenerator(ccEndpoint, routing.CCRoutes),
		uaaRequestGenerator: rata.NewRequestGenerator(uaaEndpoint, routing.UAARoutes),
		httpClient:          &http.Client{},
		ccEndpoint:          ccEndpoint,
		token:               token,
	}
}

func (c client) RefreshAuthToken() error {
	req, err := c.uaaRequestGenerator.CreateRequest("refresh_token", rata.Params{}, strings.NewReader(""))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte("cf:")))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// responseBytes, _ := ioutil.ReadAll(resp.Body)
	// if resp.StatusCode != http.StatusOK {
	// 	return responseBytes, fmt.Errorf("Received response with status code %d", resp.StatusCode)
	// }

	return nil
}

func (c client) SetToken(token string) {
	c.token = token
}

func (c client) GetApplications(queryParams url.Values) ([]byte, error) {
	req, err := c.ccRequestGenerator.CreateRequest("apps", rata.Params{}, strings.NewReader(""))
	if err != nil {
		return []byte{}, err
	}

	path := req.URL.Path
	encodedParams := queryParams.Encode()
	if encodedParams != "" {
		path = path + "?" + encodedParams
	}

	responseBytes, err := c.GetResources(path, 0)
	if err != nil {
		return []byte{}, err
	}

	return responseBytes, nil
}

func (c client) GetResource(path string) ([]byte, error) {
	url := c.ccEndpoint + "/" + strings.TrimLeft(path, "/")

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return []byte{}, err
	}

	responseBytes, err := c.performRequest(req)
	if err != nil {
		return responseBytes, err
	}

	return responseBytes, nil
}

func (c client) GetResources(path string, limit int) ([]byte, error) {
	var resources, rs []interface{}
	var err error

	nextPath := &path

	for nextPath != nil && (limit == 0 || len(resources) < limit) {
		rs, nextPath, err = c.getResources(nextPath)
		if err != nil {
			return []byte{}, err
		}
		resources = append(resources, rs...)
	}

	if limit > 0 {
		resources = resources[:limit]
	}

	responseJSON, err := json.Marshal(resources)
	if err != nil {
		return []byte{}, err
	}

	return responseJSON, nil
}

func (c client) getResources(path *string) ([]interface{}, *string, error) {
	u, err := url.Parse(*path)
	if err != nil {
		return []interface{}{}, nil, err
	}

	req, err := http.NewRequest("GET", c.ccEndpoint+u.Path, nil)
	if err != nil {
		return []interface{}{}, nil, err
	}

	req.URL.RawQuery = u.Query().Encode()

	responseBytes, err := c.performRequest(req)
	if err != nil {
		return []interface{}{}, nil, err
	}

	getResourcesResponse := &GetResourcesResponse{}
	err = json.Unmarshal(responseBytes, getResourcesResponse)
	if err != nil {
		return []interface{}{}, nil, err
	}

	return getResourcesResponse.Resources, getResourcesResponse.Pagination.Next, nil
}

func (c client) performRequest(req *http.Request) ([]byte, error) {
	req.Header.Set("Authorization", c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return []byte{}, err
	}
	defer resp.Body.Close()

	responseBytes, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return responseBytes, fmt.Errorf("Received response with status code %d", resp.StatusCode)
	}

	return responseBytes, nil
}
