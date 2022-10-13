package main

import (
	"context"
	"fmt"

	pkg "github.com/ArcticXWolf/sensu-check-kubernetes/pkg"
	"github.com/sensu/sensu-go/types"
	"github.com/sensu/sensu-plugin-sdk/sensu"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Config represents the check plugin config.
type Config struct {
	sensu.PluginConfig
	Namespace     string
	ResourceKind  string
	ResourceName  string
	LabelSelector string
	FieldSelector string
}

var (
	plugin = Config{
		PluginConfig: sensu.PluginConfig{
			Name:     "sensu-check-kubernetes",
			Short:    "Kubernetes checks for Sensu",
			Keyspace: "sensu.io/plugins/sensu-check-kubernetes/config",
		},
	}

	options = []sensu.ConfigOption{
		&sensu.PluginConfigOption[string]{
			Path:      "namespace",
			Env:       "",
			Argument:  "namespace",
			Shorthand: "n",
			Default:   "",
			Usage:     "Name of the namespace to query from (leave empty to check clusterwide)",
			Value:     &plugin.Namespace,
		},
		&sensu.PluginConfigOption[string]{
			Path:      "resource-kind",
			Env:       "",
			Argument:  "resource-kind",
			Shorthand: "t",
			Default:   "Pod",
			Usage:     "Resource to query (e.g. Pod)",
			Value:     &plugin.ResourceKind,
		},
		&sensu.PluginConfigOption[string]{
			Path:      "label-selector",
			Env:       "",
			Argument:  "label-selector",
			Shorthand: "l",
			Default:   "",
			Usage:     "Label selector to filter resources",
			Value:     &plugin.LabelSelector,
		},
		&sensu.PluginConfigOption[string]{
			Path:      "field-selector",
			Env:       "",
			Argument:  "field-selector",
			Shorthand: "f",
			Default:   "",
			Usage:     "Field selector to filter resources",
			Value:     &plugin.FieldSelector,
		},
	}
)

func main() {
	check := sensu.NewCheck(&plugin.PluginConfig, options, checkArgs, executeCheck, false)
	check.Execute()
}

func checkArgs(event *types.Event) (int, error) {
	return sensu.CheckStateOK, nil
}

func executeCheck(event *types.Event) (int, error) {

	amount, err := getNumResources(plugin.Namespace, plugin.ResourceKind, metav1.ListOptions{LabelSelector: plugin.LabelSelector, FieldSelector: plugin.FieldSelector})
	if err != nil {
		return sensu.CheckStateCritical, err
	}
	fmt.Printf("AmountResourcesFound: %d\n", amount)

	return sensu.CheckStateOK, nil
}

func getNumResources(namespace string, resourcekind string, opts metav1.ListOptions) (int, error) {
	api_client, err := pkg.GetKubeApiClient()
	if err != nil {
		return -1, err
	}

	mapping, err := pkg.GetResourceMapping(resourcekind)
	if err != nil {
		return -1, err
	}

	list, err := api_client.Resource(mapping.Resource).Namespace(namespace).Get(context.TODO())
	if err != nil {
		return -1, err
	}

	return len(list.Items), nil
}
