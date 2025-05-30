// Copyright (c) Tetrate, Inc All Rights Reserved.

package main

import (
	"crypto/sha256"
	_ "embed"
	"fmt"
	"os"
	"strings"

	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/spf13/cobra"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

func main() {
	if err := GetRootCommand().Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func GetRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:               "kong2eg",
		Long:              "A command line tool to print Envoy Gateway configuration to run Kong plugins with kong2envoy.",
		Short:             "kong2eg",
		Example:           "kong2eg print --kong-config kong.yaml --namespace default --gateway eg --gatewayclass eg",
		SilenceErrors:     true,
		SilenceUsage:      true,
		DisableAutoGenTag: true,
	}

	rootCmd.AddCommand(printCmd())

	return rootCmd
}

func printCmd() *cobra.Command {
	p := kong2envoyParameters{}
	printCmd := &cobra.Command{
		Use:     "print ",
		Aliases: []string{"p"},
		Short:   "Prints Envoy Gateway configuration to run Kong plugins with kong2envoy",
		Run: func(c *cobra.Command, args []string) {
			cmdutil.CheckErr(printKong2envoy(p))
		},
	}
	printCmd.PersistentFlags().StringVar(&p.KongConfig, "kong-config", "", "Kong configuration file to use. Defaults to the embedded demo configuration.")
	printCmd.PersistentFlags().StringVar(&p.Namespace, "namespace", "default", "Kubernetes namespace to use for the Envoy Gateway resources.")
	printCmd.PersistentFlags().StringVar(&p.GatewayName, "gateway", "eg", "Name of the Envoy Gateway resource.")
	printCmd.PersistentFlags().StringVar(&p.GatewayClassName, "gatewayclass", "eg", "Name of the GatewayClass resource.")

	return printCmd
}

type kong2envoyParameters struct {
	KongConfig          string
	Namespace           string
	GatewayName         string
	GatewayClassName    string
	EnvoyServiceAccount string
}

//go:embed templates/kong2envoy.yaml.tpl
var tmplStr string

//go:embed templates/demo-kong-config.yaml
var defaultKongConfig string

func printKong2envoy(p kong2envoyParameters) error {
	if p.KongConfig != "" {
		userConfig, err := os.ReadFile(p.KongConfig)
		if err != nil {
			return fmt.Errorf("failed to read kong config file %s: %w", p.KongConfig, err)
		}
		p.KongConfig = string(userConfig)
	} else {
		p.KongConfig = defaultKongConfig
		fmt.Fprintln(os.Stderr, "No kong config file provided, using the embedded demo configuration.")
	}
	p.EnvoyServiceAccount = gwServiceAccountName(p.Namespace, p.GatewayName)

	var tmpl = template.Must(template.New("kong2envoy").Funcs(sprig.TxtFuncMap()).Parse(tmplStr))

	return tmpl.Execute(os.Stdout, p)
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
