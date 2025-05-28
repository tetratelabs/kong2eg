# Kong2eg

This repository demonstrates how to use kong2envoy to migrate from Kong to Envoy Gateway.
Most existing Kong plugins are supported — except those that require a database.

# How it works

Envoy Gateway runs Kong as an external processing extension. Kong intercepts requests and responses before Envoy Proxy forwards them to the backend or client.

You define the Kong configuration in a ConfigMap labeled:

```yaml
extension.tetrate.io/kong-config: "true"
```

All routing decisions are made by Envoy Gateway.
Kong is only responsible for modifying requests and responses (e.g., adding headers, modifying body, etc).

## Install Envoy Gateway with Backend resource enabled:

```bash
helm install eg oci://docker.io/envoyproxy/gateway-helm \
  --version v1.4.0 \
  --set config.envoyGateway.extensionApis.enableBackend=true \
  -n envoy-gateway-system \
  --create-namespace
```

## Deploy the demo app

```bash
kubectl apply -f https://github.com/envoyproxy/gateway/releases/download/v1.4.0/quickstart.yaml
```

## Install kong2envoy

This step deploys kong2envoy as a sidecar to Envoy Proxy and sets up an EnvoyExtensionPolicy to use it for request and response processing.

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

Caveats:

* kong2envoy needs ConfigMap read permissions to load Kong configuration. As a temporary workaround, this demo creates a RoleBinding to grant the necessary permissions. This is a known limitation and should be addressed properly in future versions of Envoy Gateway.
