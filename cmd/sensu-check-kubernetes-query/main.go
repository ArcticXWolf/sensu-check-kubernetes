package main

import (
	"context"
	"fmt"

	pkg "github.com/ArcticXWolf/sensu-check-kubernetes/pkg"
	"github.com/PaesslerAG/gval"
	"github.com/itchyny/gojq"
	"github.com/sensu/sensu-go/types"
	"github.com/sensu/sensu-plugin-sdk/sensu"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// Config represents the check plugin config.
type Config struct {
	sensu.PluginConfig
	Namespace     string
	ResourceKind  string
	ResourceName  string
	LabelSelector string
	FieldSelector string
	Query         string
	Expression    string
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
			Usage:     "Resource kind to query (e.g. Pod)",
			Value:     &plugin.ResourceKind,
		},
		&sensu.PluginConfigOption[string]{
			Path:      "resource-name",
			Env:       "",
			Argument:  "resource-name",
			Shorthand: "r",
			Default:   "",
			Usage:     "Resource name to query",
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
			Path:      "expression",
			Env:       "",
			Argument:  "expression",
			Shorthand: "e",
			Default:   "",
			Usage:     "Expression for comparing result of query",
			Value:     &plugin.Expression,
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
		fmt.Printf("failed to parse query %q, error: %v", plugin.Query, err)
		return sensu.CheckStateCritical, err
	}

	code, err := gojq.Compile(query)
	if err != nil {
		fmt.Printf("failed to compile query %q, error: %v", plugin.Query, err)
		return sensu.CheckStateCritical, nil
	}

	iter := code.Run(result.Object)

	var value interface{}
	for {
		var ok bool
		v, ok := iter.Next()
		if !ok {
			break
		}

		if _, ok := v.(error); ok {
			continue
		}

		value = v
	}

	found, err := evaluateExpression(value, plugin.Expression)
	if err != nil {
		return sensu.CheckStateCritical, fmt.Errorf("error evaluating expression: %v", err)
	}
	if found {
		fmt.Printf("%s OK:  The value %v found at %s matched with expression %q and returned true\n", plugin.PluginConfig.Name, value, plugin.Query, plugin.Expression)
		return sensu.CheckStateOK, nil
	}

	fmt.Printf("%s CRITICAL: The value %v found at %s did not match with expression %q and returned false\n", plugin.PluginConfig.Name, value, plugin.Query, plugin.Expression)
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

func evaluateExpression(actualValue interface{}, expression string) (bool, error) {
	evalResult, err := gval.Evaluate("value "+expression, map[string]interface{}{"value": actualValue})
	if err != nil {
		return false, err
	}
	return evalResult.(bool), nil
}
