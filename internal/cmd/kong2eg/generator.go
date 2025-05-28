// Copyright (c) Tetrate, Inc All Rights Reserved.

package kong2eg

import (
	"crypto/sha256"
	_ "embed"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
)

//go:embed templates/kong2envoy.yaml.tpl
var tmplStr string

//go:embed templates/demo-kong-config.yaml
var defaultKongConfig string

func kong2envoyConfig(p kong2envoyParameters) (string, error) {
	if p.KongConfig != "" {
		userConfig, err := os.ReadFile(p.KongConfig)
		if err != nil {
			return "", fmt.Errorf("failed to read kong config file %s: %w", p.KongConfig, err)
		}
		p.KongConfig = string(userConfig)
	} else {
		p.KongConfig = defaultKongConfig
		fmt.Fprintln(os.Stderr, "No kong config file provided, using the embedded demo configuration.")
	}
	if p.EnvoyGatewayDeployMode != "ControllerNamespace" && p.EnvoyGatewayDeployMode != "GatewayNamespace" {
		return "", fmt.Errorf("invalid Envoy Gateway deployment mode: %s, must be either ControllerNamespace or GatewayNamespace", p.EnvoyGatewayDeployMode)
	}
	p.EnvoyServiceAccount = gwServiceAccountName(p.Namespace, p.GatewayName)

	tmpl := template.Must(template.New("kong2envoy").Funcs(sprig.TxtFuncMap()).Parse(tmplStr))

	var output strings.Builder
	if err := tmpl.Execute(&output, p); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}
	return output.String(), nil
}

func gwServiceAccountName(namespace, gatewayName string) string {
	if namespace == "" {
		namespace = "default"
	}
	if gatewayName == "" {
		gatewayName = "eg"
	}
	return fmt.Sprintf("envoy-%s", hashedName(fmt.Sprintf("%s/%s", namespace, gatewayName), 48))
}

func hashedName(nsName string, length int) string {
	hashedName := digest256(nsName)
	// replace `/` with `-` to create a valid K8s resource name
	resourceName := strings.ReplaceAll(nsName, "/", "-")
	if length > 0 && len(resourceName) > length {
		// resource name needs to be trimmed, as container port name must not contain consecutive hyphens
		trimmedName := strings.TrimSuffix(resourceName[0:length], "-")
		return fmt.Sprintf("%s-%s", trimmedName, hashedName[0:8])
	}
	// Ideally we should use 32-bit hash instead of 64-bit hash and return the first 8 characters of the hash.
	// However, we are using 64-bit hash to maintain backward compatibility.
	return fmt.Sprintf("%s-%s", resourceName, hashedName[0:8])
}

// Digest256 returns a sha256 hash of the input string.
// The hash is represented as a hexadecimal string of length 64.
func digest256(str string) string {
	h := sha256.New() // Using sha256 instead of sha1 due to Blocklisted import crypto/sha1: weak cryptographic primitive (gosec)
	h.Write([]byte(str))
	return strings.ToLower(fmt.Sprintf("%x", h.Sum(nil)))
}
