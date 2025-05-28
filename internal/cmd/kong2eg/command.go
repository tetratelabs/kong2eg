// Copyright (c) Tetrate, Inc All Rights Reserved.

package kong2eg

import (
	_ "embed"
	"fmt"

	"github.com/spf13/cobra"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

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
	printCmd.PersistentFlags().StringVar(&p.EnvoyGatewayDeployMode, "envoy-gateway-deploy-mode", "ControllerNamespace", "Envoy Gateway deployment mode. Options: ControllerNamespace, or GatewayNamespace.")

	return printCmd
}

type kong2envoyParameters struct {
	KongConfig             string
	Namespace              string
	GatewayName            string
	GatewayClassName       string
	EnvoyServiceAccount    string
	EnvoyGatewayDeployMode string
}

func printKong2envoy(p kong2envoyParameters) error {
	config, err := kong2envoyConfig(p)
	if err != nil {
		return err
	}
	fmt.Println(config)
	return nil
}
