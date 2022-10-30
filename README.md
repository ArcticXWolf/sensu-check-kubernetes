
[![Bonsai Asset Badge](https://img.shields.io/badge/Bonsai-Download%20Me-brightgreen.svg?colorB=89C967&logo=sensu)](https://bonsai.sensu.io/assets/ArcticXWolf/sensu-check-kubernetes)

# Sensu Go Kubernetes Status Check

- [Overview](#overview)
- [Usage examples](#usage-examples)
- [Configuration](#configuration)
  - [Asset registration](#asset-registration)
  - [Check configuration](#resource-configuration)
- [Functionality](#functionality)
- [Installation from source and contributing](#installation-from-source-and-contributing)

## Overview

Sensu-Check-Kubernetes is a Sensu Go Asset that aims to replace the old ruby asset [sensu-plugins-kubernetes](https://bonsai.sensu.io/assets/sensu-plugins/sensu-plugins-kubernetes). Currently it provides one binary (`sensu-check-kubernetes-metrics`) which provides generic metrics about a kubernetes query (resourcekind, label/field-selector) and one binary (`sensu-check-kubernetes-query`) which queries JSON about a resource from the kubeapi, transforms it via a user-specified `jq` query and compares that against a user-specified assertion.

## Usage examples

```
$ sensu-check-kubernetes-metrics --help
Kubernetes checks for Sensu

Usage:
  sensu-check-kubernetes-metrics [flags]
  sensu-check-kubernetes-metrics [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  version     Print the version number of this plugin

Flags:
  -f, --field-selector string   Field selector to filter resources
  -h, --help                    help for sensu-check-kubernetes-metrics
  -l, --label-selector string   Label selector to filter resources
  -n, --namespace string        Name of the namespace to query from (leave empty to check clusterwide)
  -t, --resource-kind string    Resource to query (e.g. Pod) (default "Pod")

Use "sensu-check-kubernetes-metrics [command] --help" for more information about a command.
```

```
$ sensu-check-kubernetes-query --help
Kubernetes checks for Sensu

Usage:
  sensu-check-kubernetes-query [flags]
  sensu-check-kubernetes-query [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  version     Print the version number of this plugin

Flags:
  -a, --assertion string       Result of jq query to compare against
  -h, --help                   help for sensu-check-kubernetes-query
  -n, --namespace string       Name of the namespace to query from (leave empty to check clusterwide)
  -q, --query string           Query on resource json in jq format
  -t, --resource-kind string   Resource to query (e.g. Pod) (default "Pod")
  -r, --resource-name string   Resource to query (e.g. Pod) (default "Pod")

Use "sensu-check-kubernetes-query [command] --help" for more information about a command.
```

## Configuration

### Asset Registration

Assets are the best way to make use of this plugin. If you're not using an asset, please consider doing so! If you're using sensuctl 5.13 or later, you can use the following command to add the asset: 

`sensuctl asset add ArcticXWolf/sensu-check-kubernetes`

If you're using an earlier version of sensuctl, you can find the asset on the [Bonsai Asset Index](https://bonsai.sensu.io/assets/ArcticXWolf/sensu-check-kubernetes).

### Check configuration

```yml
---
type: CheckConfig
api_version: core/v2
metadata:
  name: check-kubeapi-pods-query
spec:
  command: >-
    sensu-check-kubernetes-query
    --resource-kind "deployment.apps"
    --namespace "default"
    --resource-name "testdeployment"
    --query ".status.readyReplicas"
    --assertion "1"
  cron: "* * * * *"
  publish: true
  round_robin: true
  runtime_assets:
    - ArcticXWolf/sensu-check-kubernetes
  
---
type: CheckConfig
api_version: core/v2
metadata:
  name: check-kubeapi-pods-metrics
spec:
  command: >-
    sensu-check-kubernetes-metrics
    --resource-kind "Pod"
    --namespace "default"
  cron: "* * * * *"
  publish: true
  round_robin: true
  runtime_assets:
    - ArcticXWolf/sensu-check-kubernetes
  output_metric_format: prometheus_text
  output_metric_thresholds:
    - name: kubernetes_query_resources_total
      null_status: 2
      thresholds:
        - min: "1"
          max: ""
          status: 2
```

### Functionality

For now the asset will use the in cluster configuration to access the kube API.
Thus the agent executing the check MUST run in a pod inside the cluster and
have a serviceaccount configured with the correct permissions to make the check.

#### Example RBAC configuration

This kubernetes configuration allows checking pod count in a single namespace.
Remember to apply the service account to the sensu-agent pod.

```yml
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: sensu-agent

---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: sensu-agent
rules:
- apiGroups:
    - ""
  resources:
    - pods
  verbs:
    - list
    - get
  # add more rules as needed (for example for statefulsets.apps)

---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: sensu-agent
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: sensu-agent
subjects:
  - kind: ServiceAccount
    name: sensu-agent
```

## Installation from source and contributing

The preferred way of installing and deploying this plugin is to use it as an [asset][2]. If you would like to compile and install the plugin from source or contribute to it, download the latest version of the sensu-check-kubernetes from [releases][1]
or create an executable script from this source.

From the local path of the sensu-check-kubernetes repository:

```
go build -o /usr/local/bin/sensu-check-kubernetes-metrics cmd/sensu-check-kubernetes-metrics/main.go
go build -o /usr/local/bin/sensu-check-kubernetes-query cmd/sensu-check-kubernetes-query/main.go
```

For more information about contributing to this plugin, see https://github.com/sensu/sensu-go/blob/master/CONTRIBUTING.md

[1]: https://github.com/ArcticXWolf/sensu-check-kubernetes/releases
[2]: #asset-registration
