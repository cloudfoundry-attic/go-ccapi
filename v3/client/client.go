package client

import (
	"github.com/cloudfoundry/ccapi/models"
	"github.com/tedsuo/rata"
)

type Client interface {
	GetApplication(models.GUID) models.Application
}

type client struct {
	requestGenerator rata.RequestGenerator
}

func NewClient(target string) Client {
	return &client{
		requestGenerator: rata.NewRequestGenerator(target, routing.routes),
	}
}

type EndpointOptions struct {
	Included []string
	OrderBy  string
	Page     uint
	PerPage  uint
}

func (c Client) GetApplication(GUID models.GUID, options EndpointOptions) (models.Application, error) {
	req, err := c.requestGenerator.CreateRequest("get_application", rata.Params{"guid": "app_guid"}, nil)
	if err != nil {
		return models.Application{}, err
	}
	res, err := gateway.Do(req)
	if err != nil {
		return models.Application{}, err
	}
}
