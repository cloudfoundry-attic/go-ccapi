package routing

import "github.com/tedsuo/rata"

var Routes = rata.Routes{
	{Name: "apps", Method: "GET", Path: "/v3/apps"},
}
