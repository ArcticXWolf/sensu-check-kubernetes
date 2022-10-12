
[![Bonsai Asset Badge](https://img.shields.io/badge/Bonsai-Download%20Me-brightgreen.svg?colorB=89C967&logo=sensu)](https://bonsai.sensu.io/assets/ArcticXWolf/sensu-check-kubernetes)

# Sensu Go Kubernetes Status Check

- [Overview](#overview)
- [Usage examples](#usage-examples)
- [Configuration](#configuration)
  - [Asset registration](#asset-registration)
  - [Asset configuration](#asset-configuration)
  - [Check configuration](#resource-configuration)
- [Functionality](#functionality)
- [Installation from source and contributing](#installation-from-source-and-contributing)

## Overview

Sensu-Check-Kubernetes is a Sensu Go Asset that aims to replace the old ruby
asset
[sensu-plugins-kubernetes](https://bonsai.sensu.io/assets/sensu-plugins/sensu-plugins-kubernetes).
Currently it can count any Kubernetes resource. See usage examples for drop-in replacements of 
[sensu-plugins-kubernetes](https://bonsai.sensu.io/assets/sensu-plugins/sensu-plugins-kubernetes).

## Usage examples

```
$ ./sensu-check-kubernetes --help
Kubernetes checks for Sensu

Usage:
  sensu-check-kubernetes [flags]
  sensu-check-kubernetes [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  version     Print the version number of this plugin

Flags:
  -f, --field-selector string     Field selector to filter resources
  -h, --help                      help for sensu-check-kubernetes
  -l, --label-selector string     Label selector to filter resources
  -n, --namespace string          Name of the namespace to query from (leave empty to check clusterwide)
  -t, --resource-kind string      Resource to query (e.g. Pod) (default "Pod")
  -c, --threshold-critical int    Threshold for critical status (default 1)
      --threshold-direction int   Direction of the thresholds (-1 = critical if metric_value < threshold-critical, 1 = critical if value > threshold-critical, 0 = critical if value != threshold-critical). A zero value disables warnings. (default -1)
  -w, --threshold-warning int     Threshold for warning status (default 1)

Use "sensu-check-kubernetes [command] --help" for more information about a command.
```

### Default values

By default (i.e. with no flags given) it counts the amount of Pods
(`ResourceKind = Pod`) clusterwide (`Namespace = ""`) and returns critical, if
the amount of pods is zero (`ThresholdDirection = -1, ThresholdCritical = 1`).
Otherwise it returns Ok.

### Amount of pods running

This check returns Ok if the number of running pods in namespace "default" is exactly 5 and Critical otherwise.

```
sensu-check-kubernetes 
  --resource-kind "Pod"
  --namespace "default"
  --field-selector "status.phase=Running"
  --threshold-direction 0
  --threshold-critical 5
```

### Replacements for sensu-plugins-kubernetes


#### check-kube-pods-running.rb

```
TODO
```

#### check-kube-service-available.rb

```
TODO
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
  name: check-kubernetes-pods
spec:
  command: >-
    sensu-check-kubernetes -n "default"
  runtime_assets:
    - ArcticXWolf/sensu-check-kubernetes
  subscriptions:
    - system
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
go build -o /usr/local/bin/sensu-check-kubernetes main.go
```

For more information about contributing to this plugin, see https://github.com/sensu/sensu-go/blob/master/CONTRIBUTING.md

[1]: https://github.com/ArcticXWolf/sensu-check-kubernetes/releases
[2]: #asset-registration
