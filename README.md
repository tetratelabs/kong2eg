# Kong2eg

This repository demonstrates how to use kong2envoy to migrate from Kong to Envoy Gateway.
Most existing Kong plugins are supported — except those that require a database.

# How it works

This diagram shows how Envoy Gateway, Envoy, Kong2envoy,and Kong work together:

![kong2envoy](images/kong2eg.png)

* Envoy Gateway deploys kong2envoy as a sidecar to Envoy Proxy and sets up an EnvoyExtensionPolicy to use it for request and response processing.
* Kong2envoy is implemented as an [External Processing Extension](https://gateway.envoyproxy.io/docs/tasks/extensibility/ext-proc/)
that communicates with Envoy Proxy using gRPC over an Unix domain socket.
* Envoy Proxy forwards requests and responses to Kong2envoy for processing before forwarding them to the backend or client.
* Kong2envoy runs a Kong instance in its own container and forwards requests and responses to Kong for processing.

You define the Kong configuration in a ConfigMap labeled:

```yaml
extension.tetrate.io/kong-config: "true"
app: kong2envoy
```

All routing decisions are made by Envoy Gateway.
Kong is only responsible for modifying requests and responses (e.g., adding headers, modifying body, etc).

# Demo

This repository contains a quickstart demo that you can use to test kong2envoy.

## Install Envoy Gateway with Backend resource enabled:

```bash
helm install eg oci://docker.io/envoyproxy/gateway-helm \
  --version v1.4.0 \
  --set config.envoyGateway.extensionApis.enableBackend=true \
  -n envoy-gateway-system \
  --create-namespace
```

## Deploy the demo app

Deploy the demo app from the Envoy Gateway quickstart.

```bash
kubectl apply -f https://github.com/envoyproxy/gateway/releases/download/v1.4.0/quickstart.yaml
```

## Install kong2envoy

This step deploys kong2envoy as a sidecar to Envoy Proxy and sets up an EnvoyExtensionPolicy to use it for request and response processing.

It also creates a Role and RoleBinding to grant kong2envoy ConfigMap read permissions. This is necessary for kong2envoy to load the Kong configuration from configMaps.

```bash
kubectl apply -f manifests
```

## Test the Setup

```bash
curl http://172.18.0.200 -H "Host: www.example.com"
```

You should see response headers added by Kong:
```
"Via": [
   "1.1 kong/3.9.0"
  ],
  "X-Kong-Proxy-Latency": [
   "1"
  ],
  "X-Kong-Request-Id": [
   "f5fe5d3dfcf3a06452b66b33c2fa1c1b"
  ],
  "X-Kong-Response-Header-1": [
   "foo"
  ],
  "X-Kong-Response-Header-2": [
   "bar"
  ],
  "X-Kong-Upstream-Latency": [
   "8"
  ]
```

## Migrate Your Kong Gateway to Envoy Gateway

You can use the sample configuration in the manifests directory as a reference for migrating your Kong Gateway to Envoy Gateway.

1. Create a `EnvoyProxy` resource to deploy kong2envoy as a sidecar to Envoy Proxy.

```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: custom-proxy-config
spec:
  logging:
    level:
      default: info
  provider:
    type: Kubernetes
    kubernetes:
      envoyDeployment:
        container:
          volumeMounts:
          - name: socket-dir
            mountPath: /var/sock/kong  # uds socket for envoy to connect to kong2envoy
        initContainers:
        - name: ext-proc
          restartPolicy: Always
          image: zhaohuabing/kong2envoy:latest
          readinessProbe:
            exec:
              command: ["kong", "health"]
            initialDelaySeconds: 5
          livenessProbe:
            exec:
              command: ["kong", "health"]
            initialDelaySeconds: 10
          ports:
          - containerPort: 6060 # pprof
          env:
          - name: CPU_REQUEST
            valueFrom:
              resourceFieldRef:
                containerName: ext-proc
                resource: requests.cpu
          - name: NAMESPACE
            valueFrom:
              fieldRef:
                fieldPath: metadata.namespace
          - name: DEBUG
            value: "true"
          - name: APP_LABEL
            valueFrom:
              fieldRef:
                fieldPath: metadata.labels['app']
          volumeMounts:
          - name: socket-dir
            mountPath: /var/sock/kong
          - name: kong-config
            mountPath: /usr/local/share/kong2envoy/
          - name: podinfo
            mountPath: /etc/podinfo
          resources:
            limits:
              cpu: "6"
              memory: "2Gi"
            requests:
              cpu: "6"
              memory: "2Gi"
          securityContext:
            runAsUser: 65532
            runAsGroup: 65532
            runAsNonRoot: true
        pod:
          labels:
            app: kong2envoy  # this label is used by kong2envoy to match the ConfigMap that contains the Kong configuration
          volumes:
            - name: podinfo
              downwardAPI:
                items:
                - path: "labels"
                  fieldRef:
                    fieldPath: metadata.labels
            - name: socket-dir
              emptyDir: {}
            - name: kong-config
              emptyDir: {}
```

2. Create a `GatewayClass` resource. GatewayClass is used to define the controller that will be used to manage the Gateway resource.

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: GatewayClass
metadata:
  name: eg
spec:
  controllerName: gateway.envoyproxy.io/gatewayclass-controller
```

2. Create a `Gateway` resource. Envoy Gateway will watch the Gateway resource and deploy Envoy Proxy for it. Reference the `EnvoyProxy` resource that you created in step 1 in the `infrastructure` field.

Note: `EnvoyProxy` resource can also be [associated with a `GatewayClass` to apply it to all `Gateway` resources of that `GatewayClass`](https://gateway.envoyproxy.io/docs/tasks/operations/customize-envoyproxy/).

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: eg
spec:
  gatewayClassName: eg
  infrastructure:
    parametersRef:
      group: gateway.envoyproxy.io
      kind: EnvoyProxy
      name: custom-proxy-config
  listeners:
    - name: http
      protocol: HTTP
      port: 80
```

2. Create a `Role` and `RoleBinding` to grant kong2envoy ConfigMap read permissions. Please change the name of the Role and RoleBinding to match your Envoy service account name.

The Envoy service account name is the same as the Envoy deployment name. It's usually in the format of `envoy-default-eg-xxxxxxxx`.

You can find the Envoy service account name by running the following command:

```bash
kubectl get sa -n envoy-gateway-system --selector=gateway.envoyproxy.io/owning-gateway-name=${GATEWAY_NAME}
```

${GATEWAY_NAME} is the name of the Gateway resource that you created in step 1.

For example:

```bash
kubectl get sa -n envoy-gateway-system --selector=gateway.envoyproxy.io/owning-gateway-name=eg
NAME                        SECRETS   AGE
envoy-default-eg-e41e7b31   0         14m
```

```yaml
# RoleBinding is created to grant kong2envoy ConfigMap read permissions.
# This is necessary for kong2envoy to load the Kong configuration from ConfigMaps.
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: envoy-default-eg-e41e7b31  # change this name to match your Envoy service account name
  namespace: envoy-gateway-system
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: envoy-default-eg-e41e7b31 # change this name to match your Envoy service account name
subjects:
- kind: ServiceAccount
  name: envoy-default-eg-e41e7b31
  namespace: envoy-gateway-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: envoy-default-eg-e41e7b31 # change this name to match your Envoy service account name
  namespace: envoy-gateway-system
rules:
- apiGroups:
  - ""
  resources:
  - configmaps
  verbs:
  - get
  - list
  - watch
```

3. Create one or more `HTTPRoute` resources to define the routing rules for your application.

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: httproute
spec:
  hostnames:
  - www.example.com
  parentRefs:
  - group: gateway.networking.k8s.io
    kind: Gateway
    name: eg
  rules:
  - backendRefs:
    - group: ""
      kind: Service
      name: backend
      port: 3000
      weight: 1
    matches:
    - path:
        type: PathPrefix
        value: /
```

3. Create a `Backend` resource to define the backend service that will be used by Envoy Proxy to connect to kong2envoy.

```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: Backend
metadata:
  name: kong2envoy
spec:
  endpoints:
  - unix:
      path: /var/sock/kong/ext-proc.sock
```

4. Create an `EnvoyExtensionPolicy` resource to tell Envoy Proxy to use kong2envoy as an external processing extension.

```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyExtensionPolicy
metadata:
  name: kong2envoy
spec:
  targetRefs:
    - group: gateway.networking.k8s.io
      kind: Gateway
      name: eg
  extProc:
  - backendRefs:
    - name: kong2envoy
      kind: Backend
      group: gateway.envoyproxy.io
    processingMode:
      request: {}
      response: {}
```

5. Configure the Kong configuration in a ConfigMap. You can migrate your existing Kong configuration to the ConfigMap.

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: kong-config
  namespace: envoy-gateway-system
  labels:
    extension.tetrate.io/kong-config: "true"
    app: kong2envoy
data:
  config: |+
    kong.yaml: |+
      _format_version: "3.0"
      _transform: true
      services:
      - name: example
        url: http://www.example.com
        routes:
        - name: my-route-0
          paths:
          - /
          plugins:
          - name: request-transformer
            config:
              add:
                headers:
                - "x-kong-request-header-1:foo"
                - "x-kong-request-header-2:bar"
          - name: response-transformer
            config:
              add:
                headers:
                - "x-kong-response-header-1:foo"
                - "x-kong-response-header-2:bar"

```
