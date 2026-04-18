package models

type Action string

const (
	ActionBuild   Action = "BUILD"
	ActionDeploy  Action = "DEPLOY"
	ActionRestart Action = "RESTART"
	ActionStatus  Action = "GET_STATUS"
)

type TaskPayload struct {
	ID          string            `json:"id"`
	Action      Action            `json:"action"`
	RepoURL     string            `json:"repo_url"`
	ImageName   string            `json:"image_name"`
	ContainerID string            `json:"container_id"`
	EnvVars     map[string]string `json:"env_vars"`
}
