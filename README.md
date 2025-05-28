# Kong2Envoy

This repository is a demo of using kong2envoy to migrate Kong to Envoy Gateway.

## Install Envoy Gateway with Backend resource enabled:

```bash
helm install eg oci://docker.io/envoyproxy/gateway-helm \
  --version v1.4.0 \
  --set config.envoyGateway.extensionApis.enbaleBackend=true \
  -n envoy-gateway-system \
  --create-namespace
```

## Deploy the demo app

```bash
kubectl apply -f https://github.com/envoyproxy/gateway/releases/download/v1.4.0/quickstart.yaml
```

## Install kong2envoy

This will deploy kong2envoy as a sidecar to Envoy proxy, and create an EnvoyExtensionPolicy to use kong2envoy to process requests and responses.

kong2envoy runs a kong process and process requests and responses. The mutated requests and responses are sent to Envoy proxy before Envoy proxy forwards them to the backend/client.

The kong config can be specified via a ConfigMap wit the label `extension.tetrate.io/kong-config: "true"`, kong2envoy will watch the ConfigMap and reload the kong process when the ConfigMap changes.

```bash
kubectl apply -f manifests
```

## Test

```bash
curl http://172.18.0.200 -H "Host: www.example.com"
```

You should see the headers added by kong in the response:
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

1. kong2envoy requires ConfigMap permissions to read the kong config, as a temporary workaround, this demo creates a RoleBinding to grant the permissions. This needs to be fixed in the Envoy Gateway.
