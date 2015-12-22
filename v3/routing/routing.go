package routing

import "github.com/tedsuo/rata"

var routes = rata.Routes{
	{Name: "get_application", Method: "GET", Path: "/v3/apps/:guid"},
}
