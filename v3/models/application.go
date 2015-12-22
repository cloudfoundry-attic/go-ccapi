package models

import (
	"encoding/json"
	"time"
)

type Application struct {
	CreatedAt             time.Time
	DesiredState          DesiredState
	EnvironmentVariables  map[string]interface{}
	GUID                  GUID
	Included              json.RawMessage
	Lifecycle             Lifecycle
	Links                 Links
	Name                  string
	TotalDesiredInstances uint
	UpdatedAt             time.Time
}
