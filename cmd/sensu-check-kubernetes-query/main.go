package main

import (
	"context"
	"encoding/json"
	"fmt"

	pkg "github.com/ArcticXWolf/sensu-check-kubernetes/pkg"
	"github.com/itchyny/gojq"
	"github.com/sensu/sensu-go/types"
	"github.com/sensu/sensu-plugin-sdk/sensu"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// Config represents the check plugin config.
type Config struct {
	sensu.PluginConfig
	Namespace       string
	ResourceKind    string
	ResourceName    string
	LabelSelector   string
	FieldSelector   string
	Query           string
	ResultAssertion string
}

var (
	plugin = Config{
		PluginConfig: sensu.PluginConfig{
			Name:     "sensu-check-kubernetes-query",
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
			Path:      "resource-name",
			Env:       "",
			Argument:  "resource-name",
			Shorthand: "r",
			Default:   "Pod",
			Usage:     "Resource to query (e.g. Pod)",
			Value:     &plugin.ResourceName,
		},
		&sensu.PluginConfigOption[string]{
			Path:      "query",
			Env:       "",
			Argument:  "query",
			Shorthand: "q",
			Default:   "",
			Usage:     "Query on resource json in jq format",
			Value:     &plugin.Query,
		},
		&sensu.PluginConfigOption[string]{
			Path:      "assertion",
			Env:       "",
			Argument:  "assertion",
			Shorthand: "a",
			Default:   "",
			Usage:     "Result of jq query to compare against",
			Value:     &plugin.ResultAssertion,
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
	result, err := getResource(plugin.Namespace, plugin.ResourceKind, plugin.ResourceName)
	if err != nil {
		return sensu.CheckStateCritical, err
	}

	query, err := gojq.Parse(plugin.Query)
	if err != nil {
		return sensu.CheckStateCritical, err
	}

	iter := query.Run(result.Object)

	check := true
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}

		if err, ok := v.(error); ok {
			fmt.Println(err)
			break
		}
		marsh, err := json.Marshal(v)
		if err != nil {
			fmt.Println(err)
			break
		}

		fmt.Printf("Found: %s\n", marsh)
		fmt.Printf("Assertion against: %s\n", plugin.ResultAssertion)
		if plugin.ResultAssertion != string(marsh) {
			check = false
		}
	}

	if check {
		fmt.Println("Result: OK")
		return sensu.CheckStateOK, nil
	}
	fmt.Println("Result: Critical")
	return sensu.CheckStateCritical, nil
}

func getResource(namespace string, resourcekind string, resourcename string) (*unstructured.Unstructured, error) {
	api_client, err := pkg.GetKubeApiClient()
	if err != nil {
		return nil, err
	}

	mapping, err := pkg.GetResourceMapping(resourcekind)
	if err != nil {
		return nil, err
	}

	result, err := api_client.Resource(mapping.Resource).Namespace(namespace).Get(context.TODO(), resourcename, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return result, nil
}
