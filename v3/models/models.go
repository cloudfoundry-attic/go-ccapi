package models

type GUID string

type LifecycleData struct {
	Buildpack string
	Stack     string
}

type Lifecycle struct {
	Type string
	Data LifecycleData
}

type DesiredState string

const DesiredStateStopped DesiredState = "STOPPED"
