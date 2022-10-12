package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/sensu/sensu-go/types"
	"github.com/sensu/sensu-plugin-sdk/sensu"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
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
		&sensu.PluginConfigOption[int]{
			Path:      "threshold-critical",
			Env:       "",
			Argument:  "threshold-critical",
			Shorthand: "c",
			Default:   1,
			Usage:     "Threshold for critical status",
			Value:     &plugin.ThresholdCritical,
		},
		&sensu.PluginConfigOption[int]{
			Path:      "threshold-warning",
			Env:       "",
			Argument:  "threshold-warning",
			Shorthand: "w",
			Default:   1,
			Usage:     "Threshold for warning status",
			Value:     &plugin.ThresholdWarning,
		},
		&sensu.PluginConfigOption[int]{
			Path:      "threshold-direction",
			Env:       "",
			Argument:  "threshold-direction",
			Shorthand: "",
			Default:   -1,
			Usage:     "Direction of the thresholds (-1 = critical if metric_value < threshold-critical, 1 = critical if value > threshold-critical, 0 = critical if value != threshold-critical). A zero value disables warnings.",
			Value:     &plugin.ThresholdDirection,
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
	fmt.Printf("ResourceKind: %s", plugin.ResourceKind)
	fmt.Printf("Namespace: %s", plugin.Namespace)
	fmt.Printf("LabelSelector: %s", plugin.LabelSelector)
	fmt.Printf("FieldSelector: %s", plugin.FieldSelector)
	fmt.Printf("ThresholdCritical: %d", plugin.ThresholdCritical)
	fmt.Printf("ThresholdWarning: %d", plugin.ThresholdWarning)
	fmt.Printf("ThresholdDirection: %d", plugin.ThresholdDirection)

	amount, err := getNumResources(plugin.Namespace, plugin.ResourceKind, metav1.ListOptions{LabelSelector: plugin.LabelSelector, FieldSelector: plugin.FieldSelector})
	if err != nil {
		return sensu.CheckStateCritical, err
	}
	fmt.Printf("AmountResourcesFound: %d", amount)

	responseCode, err := getResponseCodeFromThresholds(amount, plugin.ThresholdCritical, plugin.ThresholdWarning, plugin.ThresholdDirection)
	if err != nil {
		return sensu.CheckStateCritical, err
	}
	fmt.Printf("Result: %d", responseCode)

	return responseCode, nil
}

func getNumResources(namespace string, resourcekind string, opts metav1.ListOptions) (int, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return -1, err
	}

	client, err := dynamic.NewForConfig(config)
	if err != nil {
		return -1, err
	}

	discoveryclient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return -1, err
	}

	groupResources, err := restmapper.GetAPIGroupResources(discoveryclient)
	if err != nil {
		return -1, err
	}

	mapper := restmapper.NewDiscoveryRESTMapper(groupResources)

	mapping, err := mapper.RESTMapping(schema.ParseGroupKind(resourcekind))
	if err != nil {
		return -1, err
	}

	list, err := client.Resource(mapping.Resource).Namespace(namespace).List(context.TODO(), opts)
	if err != nil {
		return -1, err
	}

	return len(list.Items), nil
}

func getResponseCodeFromThresholds(value int, thresholdCritical int, thresholdWarning int, thresholdDirection int) (int, error) {
	if thresholdDirection == 0 {
		if value == thresholdCritical {
			return sensu.CheckStateOK, nil
		} else {
			return sensu.CheckStateCritical, nil
		}
	} else if thresholdDirection < 0 {
		if thresholdWarning < thresholdCritical {
			return sensu.CheckStateCritical, errors.New("threshold direction is < 0, but warning threshold is bigger than critical threshold")
		}

		if value < thresholdCritical {
			return sensu.CheckStateCritical, nil
		} else if value < thresholdWarning {
			return sensu.CheckStateWarning, nil
		} else {
			return sensu.CheckStateOK, nil
		}
	} else {
		if thresholdWarning > thresholdCritical {
			return sensu.CheckStateCritical, errors.New("threshold direction is > 0, but warning threshold is less than critical threshold")
		}

		if value > thresholdCritical {
			return sensu.CheckStateCritical, nil
		} else if value > thresholdWarning {
			return sensu.CheckStateWarning, nil
		} else {
			return sensu.CheckStateOK, nil
		}
	}
}
