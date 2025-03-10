package kontrolerclient

import (
	"fmt"
	"regexp"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

type PVC struct {
	AccessModes      []corev1.PersistentVolumeAccessMode `json:"accessModes"`
	Selector         *metav1.LabelSelector               `json:"selector,omitempty"`
	Resources        corev1.ResourceRequirements         `json:"resources,omitempty"`
	StorageClassName *string                             `json:"storageClassName,omitempty"`
	VolumeMode       *corev1.PersistentVolumeMode        `json:"volumeMode,omitempty"`
}

type Workspace struct {
	Enabled bool `json:"enable"`
	PvcSpec PVC  `json:"pvc"`
}

type Dag struct {
	Name       string             `json:"name"`
	Schedule   string             `json:"schedule,omitempty"`
	Tasks      []TaskSpec         `json:"tasks"`
	Parameters []DagParameterSpec `json:"parameters,omitempty"`
	Namespace  string             `json:"namespace"`
	Webhook    Webhook            `json:"webhook"`
	Workspace  *Workspace         `json:"workspace,omitempty"`
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
