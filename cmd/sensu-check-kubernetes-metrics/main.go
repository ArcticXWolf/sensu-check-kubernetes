package main

import (
	"context"
	"time"

	pkg "github.com/ArcticXWolf/sensu-check-kubernetes/pkg"
	dto "github.com/prometheus/client_model/go"
	"github.com/sensu/sensu-go/types"
	"github.com/sensu/sensu-plugin-sdk/sensu"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// Config represents the check plugin config.
type Config struct {
	sensu.PluginConfig
	Namespace          string
	ResourceKind       string
	ThresholdWarning   int
	ThresholdCritical  int
	ThresholdDirection int
	LabelSelector      string
	FieldSelector      string
}

var (
	plugin = Config{
		PluginConfig: sensu.PluginConfig{
			Name:     "sensu-check-kubernetes-metrics",
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
	resources, err := getResources(plugin.Namespace, plugin.ResourceKind, metav1.ListOptions{LabelSelector: plugin.LabelSelector, FieldSelector: plugin.FieldSelector})
	if err != nil {
		return sensu.CheckStateCritical, err
	}

	metrics := extractMetrics(plugin.ResourceKind, resources)
	pkg.PrintMetrics(metrics)

	return sensu.CheckStateOK, nil
}

func getResources(namespace string, resourcekind string, opts metav1.ListOptions) ([]unstructured.Unstructured, error) {
	api_client, err := pkg.GetKubeApiClient()
	if err != nil {
		return []unstructured.Unstructured{}, err
	}

	mapping, err := pkg.GetResourceMapping(resourcekind)
	if err != nil {
		return []unstructured.Unstructured{}, err
	}

	list, err := api_client.Resource(mapping.Resource).Namespace(namespace).List(context.TODO(), opts)
	if err != nil {
		return []unstructured.Unstructured{}, err
	}

	return list.Items, nil
}

func extractMetrics(resourcekind string, resourceslist []unstructured.Unstructured) []*dto.MetricFamily {
	metrics := make([]*dto.MetricFamily, 0, 7)
	nowMS := time.Now().UnixMilli()

	metrics = extractGenericMetrics(metrics, resourcekind, resourceslist, nowMS)

	// TODO: Add more metrics (also specific to different resourcekinds)

	return metrics
}

func extractGenericMetrics(metrics []*dto.MetricFamily, resourcekind string, resourceslist []unstructured.Unstructured, timestampMS int64) []*dto.MetricFamily {
	metrics = pkg.AddNewMetric(metrics, "kubernetes_query_resources_total", uint64(len(resourceslist)), timestampMS)
	return metrics
}
