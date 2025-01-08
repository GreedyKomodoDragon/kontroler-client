package kontrolerclient

import (
	"fmt"
	"regexp"
)

type DagParameterSpec struct {
	Name     string `json:"name"`
	IsSecret bool   `json:"isSecret"`
	Value    string `json:"value"`
}

type Webhook struct {
	URL       string `json:"url"`
	VerifySSL bool   `json:"verifySSL"`
}

type Dag struct {
	Name       string             `json:"name"`
	Schedule   string             `json:"schedule,omitempty"`
	Tasks      []TaskSpec         `json:"tasks"`
	Parameters []DagParameterSpec `json:"parameters,omitempty"`
	Namespace  string             `json:"namespace"`
	Webhook    Webhook            `json:"webhook"`
}

func (d *Dag) Validate() error {
	nameRegex := regexp.MustCompile(`^[a-z]([-a-z0-9]*[a-z0-9])?$`)
	if len(d.Name) < 1 || len(d.Name) > 63 || !nameRegex.MatchString(d.Name) {
		return fmt.Errorf("invalid DAG name: %s", d.Name)
	}

	for _, task := range d.Tasks {
		if len(task.Name) < 1 || len(task.Name) > 63 || !nameRegex.MatchString(task.Name) {
			return fmt.Errorf("invalid task name: %s", task.Name)
		}
	}

	return nil
}

type TaskSpec struct {
	Name         string   `json:"name"`
	Command      []string `json:"command,omitempty"`
	Args         []string `json:"args,omitempty"`
	Script       string   `json:"script,omitempty"`
	Image        string   `json:"image"`
	RunAfter     []string `json:"runAfter,omitempty"`
	BackoffLimit int      `json:"backoffLimit"`
	RetryCodes   []int    `json:"retryCodes,omitempty"`
	Parameters   []string `json:"parameters,omitempty"`
	PodTemplate  string   `json:"podTemplate,omitempty"`
	TaskRef      *TaskRef `json:"taskRef,omitempty"`
}

type TaskRef struct {
	Name    string `json:"name"`
	Version int    `json:"version"`
}

type DagRun struct {
	Name       string            `json:"name"`
	RunName    string            `json:"runName"`
	Parameters map[string]string `json:"parameters"`
	Namespace  string            `json:"namespace"`
}
