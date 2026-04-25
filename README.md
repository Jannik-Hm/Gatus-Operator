# Gatus Operator
This operator is not affiliated with gatus.io.
It is purely community maintained.

> [!IMPORTANT]  
> This operator is still in early alpha and CRDs are therefore subject to change.

## Description
This Operator aims to simplify Gatus Deployment and Management.
It aims to exceed capabilities of Gatus Sidecar by implementing CRDs for announcements, suites, etc.
In addition to that, it also allows running multiple seperate instances in the cluster.

## Installation

This part is work in progress, goal is to have a helm chart ready for installation

## Usage

### Creating a Gatus Instance

To create a gatus instance, use the `Instance` CRD like the example below.

```yaml
apiVersion: gatus.io/v1alpha1
kind: Instance
metadata:
  name: instance-sample
spec:
  replicas: 1 # optional, defaults to 1, currently limited to 1 as gatus does not support HA yet
  image: "ghcr.io/twin/gatus:stable" # optional, defaults to your configured DEFAULT_GATUS_IMAGE
  service:
    enable: true # optional, defaults to true, wether to create a service
    type: ClusterIP # optional, defaults to `ClusterIP`, type of the service
    annotations: # optional, custom service annotations
      foo: bar
    labels: # optional, custom service labels
      foo: bar
    ipFamilyPolicy: "PreferDualStack" # optional, ipFamilyPolicy of the Service
    ipFamilies: # optional, ipFamilies of the Service
      - "IPv4"
      - "IPv6"
  gatus-config: # custom gatus config
    metrics:
      # optional, gatus metrics config (see https://github.com/TwiN/gatus#configuration)
    alerting:
      # optional, gatus alerting config (see https://github.com/TwiN/gatus#alerting)
    concurrency:
      # optional, gatus concurrency config (see https://github.com/TwiN/gatus#concurrency)
    connectivity:
      # optional, gatus connectivity config (see https://github.com/TwiN/gatus#connectivity)
    storage:
      # optional, gatus storage config (see https://github.com/TwiN/gatus#storage)
    security:
      # optional, gatus security config (see https://github.com/TwiN/gatus#security)
    web:
      # optional, gatus web config (see https://github.com/TwiN/gatus#web)
    ui:
      # optional, gatus ui config (see https://github.com/TwiN/gatus#ui)
    maintenance:
      # optional, gatus maintenance config (see https://github.com/TwiN/gatus#maintenance)
```

### CRDs

#### Endpoint

For further details check [Gatus Endpoint Config](https://github.com/TwiN/gatus#endpoints)

```yaml
apiVersion: gatus.io/v1alpha1
kind: Endpoint
metadata:
  name: endpoint-sample
spec:
  instances: # list of instances this should be attached to
    - name: instance-sample # required, name of the gatus instance to attach to
      namespace: default # namespace of the instance, can be omitted if instance is in same namespace
  config:
    enabled: true # optional
    name: Test # required
    group: Test # optional
    url: http://test-service # required
    method: GET # optional
    conditions:
      - "[STATUS] == 200" # atleast 1 required
    interval: 60s # optional
    graphql: false # optional
    body: "" # optional
    headers: # optional
      foo: bar
    dns: # optional
      query-type: A # optional
      query-name: google.com # optional
    ssh: # optional
      username: "" # subject to change into secret ref
      password: "" # subject to change into secret ref
    alerts:
      # optional, see https://github.com/TwiN/gatus#alerting
    maintenance-windows:
      # optional, see https://github.com/TwiN/gatus#maintenance
    client:
      # optional, see https://github.com/TwiN/gatus#client-configuration
    ui:
      # optional, see https://github.com/TwiN/gatus#endpoints
    extra-labels:
      # optional
```

#### External Endpoint

For further details check [Gatus External Endpoint Config](https://github.com/TwiN/gatus#external-endpoints)

```yaml
apiVersion: gatus.io/v1alpha1
kind: ExternalEndpoint
metadata:
  name: externalendpoint-sample
spec:
  instances: # list of instances this should be attached to
    - name: instance-sample # required, name of the gatus instance to attach to
      namespace: default # namespace of the instance, can be omitted if instance is in same namespace
  config:
    enabled: true # optional
    name: Test # required
    group: Test # optional
    token: foobar # required
    alerts:
      # optional, see https://github.com/TwiN/gatus#alerting
    heartbeat: # optional
      interval: 60s # defaults to 0 (disabled)
```

#### Announcement

For further details check [Gatus Announcement Config](https://github.com/TwiN/gatus#announcements)

```yaml
apiVersion: gatus.io/v1alpha1
kind: Announcement
metadata:
  name: announcement-sample
spec:
  instances: # list of instances this should be attached to
    - name: instance-sample # required, name of the gatus instance to attach to
      namespace: default # namespace of the instance, can be omitted if instance is in same namespace
  config:
    timestamp: 2025-11-07T14:00:00Z # required, UTC timestamp when the announcement was made (RFC3339 format)
    type: none # optional, defaults to `none`, one of outage;warning;information;operational;none
    message: "This is a Message" # required
    archived: false # optional
```

#### Suite

For further details check [Gatus Suite Config](https://github.com/TwiN/gatus#suites-alpha)

```yaml
apiVersion: gatus.io/v1alpha1
kind: Suite
metadata:
  name: suite-sample
spec:
  instances: # list of instances this should be attached to
    - name: instance-sample # required, name of the gatus instance to attach to
      namespace: default # namespace of the instance, can be omitted if instance is in same namespace
  config:
    enabled: true # optional
    ...

```

### Annotations

Endpoints can be discovered via Annotations to one of the following objects: `HTTPRoute, Gateway, Ingress, IngressClass`.

Annotations to Gateways and IngressClasses are passed on to attached HTTPRoutes and Ingresses.

| Annotation | Sample | Description |
| :---       | :--- | :---
| `gatus.io/disabled` | `true` | optional flag to temporarily remove this endpoint from attached instances |
| `gatus.io/name` | `Test-Backend` | required display name of the endpoint in Gatus |
| `gatus.io/instances` | `default/instance-sample` or `instance-sample` | required *Namespace/Name* of the instance this endpoint should be attached to, namespace can be omitted if ressource is in same namespace as instance (except for ingressclasses) |
| `gatus.io/group` | `Test` | optional, Group Name |
| `gatus.io/config` | | optional, Endpoint Config Override (see https://github.com/TwiN/gatus#endpoints) |

## Development

### Prerequisites
- go version v1.24.6+
- docker version 17.03+.
- kubectl version v1.11.3+.
- Access to a Kubernetes v1.11.3+ cluster.

### To Deploy on the cluster
**Build and push your image to the location specified by `IMG`:**

```sh
make docker-build docker-push IMG=<some-registry>/projects:tag
```

**NOTE:** This image ought to be published in the personal registry you specified.
And it is required to have access to pull the image from the working environment.
Make sure you have the proper permission to the registry if the above commands don’t work.

**Install the CRDs into the cluster:**

```sh
make install
```

**Deploy the Manager to the cluster with the image specified by `IMG`:**

```sh
make deploy IMG=<some-registry>/projects:tag
```

> **NOTE**: If you encounter RBAC errors, you may need to grant yourself cluster-admin
privileges or be logged in as admin.

**Create instances of your solution**
You can apply the samples (examples) from the config/sample:

```sh
kubectl apply -k config/samples/
```

>**NOTE**: Ensure that the samples has default values to test it out.

### To Uninstall
**Delete the instances (CRs) from the cluster:**

```sh
kubectl delete -k config/samples/
```

**Delete the APIs(CRDs) from the cluster:**

```sh
make uninstall
```

**UnDeploy the controller from the cluster:**

```sh
make undeploy
```

## Project Distribution

Following the options to release and provide this solution to the users.

### By providing a bundle with all YAML files

1. Build the installer for the image built and published in the registry:

```sh
make build-installer IMG=<some-registry>/projects:tag
```

**NOTE:** The makefile target mentioned above generates an 'install.yaml'
file in the dist directory. This file contains all the resources built
with Kustomize, which are necessary to install this project without its
dependencies.

2. Using the installer

Users can just run 'kubectl apply -f <URL for YAML BUNDLE>' to install
the project, i.e.:

```sh
kubectl apply -f https://raw.githubusercontent.com/<org>/projects/<tag or branch>/dist/install.yaml
```

### By providing a Helm Chart

1. Build the chart using the optional helm plugin

```sh
kubebuilder edit --plugins=helm/v2-alpha
```

2. See that a chart was generated under 'dist/chart', and users
can obtain this solution from there.

**NOTE:** If you change the project, you need to update the Helm Chart
using the same command above to sync the latest changes. Furthermore,
if you create webhooks, you need to use the above command with
the '--force' flag and manually ensure that any custom configuration
previously added to 'dist/chart/values.yaml' or 'dist/chart/manager/manager.yaml'
is manually re-applied afterwards.

**NOTE:** Run `make help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)
