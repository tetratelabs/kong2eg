apiVersion: v1
kind: ConfigMap
metadata:
  annotations:
    extension.tetrate.io/generator: kong2eg
    tetrate.io/generated-by: kong2eg
  name: kong-config
  {{- if eq .EnvoyGatewayDeployMode "GatewayNamespace" }}
  namespace: {{ default "default" .Namespace }}
  {{- else }}
  namespace: "envoy-gateway-system"
  {{- end }}
  labels:
    extension.tetrate.io/kong-config: "true"
    app: kong2envoy
data:
  config: |+
    kong.yaml: |+
{{ .KongConfig | trimSuffix "\n" | indent 6 }}
---
apiVersion: gateway.networking.k8s.io/v1
kind: GatewayClass
metadata:
  annotations:
    tetrate.io/generated-by: kong2eg
  name: {{ default "eg" .GatewayClassName }}
spec:
  controllerName: gateway.envoyproxy.io/gatewayclass-controller
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyExtensionPolicy
metadata:
  annotations:
    tetrate.io/generated-by: kong2eg
  name: kong2envoy
  namespace: {{ default "default" .Namespace }}
spec:
  targetRefs:
    - group: gateway.networking.k8s.io
      kind: Gateway
      name: {{ default "eg" .GatewayName }}
  extProc:
  - backendRefs:
    - name: kong2envoy
      kind: Backend
      group: gateway.envoyproxy.io
    messageTimeout: 10s
    processingMode:
      request:
        body: Buffered
      response:
        body: Buffered
---
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  annotations:
    tetrate.io/generated-by: kong2eg
  name: {{ default "eg" .GatewayName }}
  namespace: {{ default "default" .Namespace }}
spec:
  gatewayClassName: {{ default "eg" .GatewayClassName }}
  infrastructure:
    parametersRef:
      group: gateway.envoyproxy.io
      kind: EnvoyProxy
      name: kong2eg-proxy-config
  listeners:
    - name: http
      protocol: HTTP
      port: 80
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  annotations:
    tetrate.io/generated-by: kong2eg
  name: kong2eg-proxy-config
  namespace: {{ default "default" .Namespace }}
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
            mountPath: /var/sock/kong
        initContainers:
        - name: ext-proc
          restartPolicy: Always
          image: tetrate/kong2envoy:v0.3.3
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
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: Backend
metadata:
  annotations:
    tetrate.io/generated-by: kong2eg
  name: kong2envoy
  namespace: {{ default "default" .Namespace }}
spec:
  endpoints:
  - unix:
      path: /var/sock/kong/ext-proc.sock
---
# RoleBinding is created to grant kong2envoy ConfigMap read permissions.
# This is necessary for kong2envoy to load the Kong configuration from ConfigMaps.
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  annotations:
    tetrate.io/generated-by: kong2eg
  name: {{ .EnvoyServiceAccount }}
  {{- if eq .EnvoyGatewayDeployMode "GatewayNamespace" }}
  namespace: {{ default "default" .Namespace }}
  {{- else }}
  namespace: envoy-gateway-system
  {{- end }}
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: {{ .EnvoyServiceAccount }}
subjects:
- kind: ServiceAccount
  name: {{ .EnvoyServiceAccount }}
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  annotations:
    tetrate.io/generated-by: kong2eg
  name: {{ .EnvoyServiceAccount }}
  {{- if eq .EnvoyGatewayDeployMode "GatewayNamespace" }}
  namespace: {{ default "default" .Namespace }}
  {{- else }}
  namespace: envoy-gateway-system
  {{- end }}
rules:
- apiGroups:
  - ""
  resources:
  - configmaps
  verbs:
  - get
  - list
  - watch
