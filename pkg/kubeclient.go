package pkg

import (
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
)

func GetKubeConfig() (*rest.Config, error) {
	return rest.InClusterConfig()
}

func GetKubeApiClient() (dynamic.Interface, error) {
	config, err := GetKubeConfig()
	if err != nil {
		return nil, err
	}

	return dynamic.NewForConfig(config)
}

func GetResourceMapping(resourcekind string) (*meta.RESTMapping, error) {
	config, err := GetKubeConfig()
	if err != nil {
		return nil, err
	}

	discoveryclient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, err
	}

	groupResources, err := restmapper.GetAPIGroupResources(discoveryclient)
	if err != nil {
		return nil, err
	}

	mapper := restmapper.NewDiscoveryRESTMapper(groupResources)

	return mapper.RESTMapping(schema.ParseGroupKind(resourcekind))
}
