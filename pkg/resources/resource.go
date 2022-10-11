package resources

import (
	"context"
	"fmt"
	"k8s.io/client-go/kubernetes"
	"strings"
)

type Resource struct {
	Kind    string
	List    func(ctx context.Context, client *kubernetes.Clientset, namespace string) ([]string, error)
	GetJSON func(ctx context.Context, client *kubernetes.Clientset, namespace, name string) ([]byte, error)
	SetJSON func(ctx context.Context, client *kubernetes.Clientset, namespace, name string, data []byte) error
}

func (r Resource) GetCanonicalYAML(ctx context.Context, client *kubernetes.Clientset, namespace, name string) (data []byte, err error) {
	if data, err = r.GetJSON(ctx, client, namespace, name); err != nil {
		return
	}
	if len(data) == 0 {
		return
	}
	if data, err = defaultSanitizers.Apply(data); err != nil {
		return
	}
	if main.GateNoResourceVersion.IsOn() {
		if data, err = noResourceVersionSanitizers.Apply(data); err != nil {
			return
		}
	}
	data, err = JSON2YAML(data)
	return
}

func (r Resource) SetCanonicalYAML(ctx context.Context, client *kubernetes.Clientset, namespace, name string, data []byte) (err error) {
	if data, err = YAML2JSON(data); err != nil {
		return
	}
	if data, err = defaultSanitizers.Apply(data); err != nil {
		return
	}
	if main.GateNoResourceVersion.IsOn() {
		if data, err = noResourceVersionSanitizers.Apply(data); err != nil {
			return
		}
	}
	if err = r.SetJSON(ctx, client, namespace, name, data); err != nil {
		return
	}
	return
}

var (
	knownResources     []*Resource
	knownResourceNames []string
)

func findResource(kind string) (resource *Resource, err error) {
	for _, knownResource := range knownResources {
		if knownResource.Kind == kind {
			resource = knownResource
		}
	}
	if resource == nil {
		err = fmt.Errorf("unknown resource kind '%s', known kinds are %s", kind, strings.Join(knownResourceNames, ", "))
		return
	}
	return
}
