// Copyright (c) Tetrate, Inc All Rights Reserved.

package kong2eg

import (
	"flag"
	"os"
	"strings"
	"testing"
)

var overwrite = flag.Bool("overwrite", false, "overwrite expected test files")

func TestKong2envoyConfig(t *testing.T) {
	tests := []struct {
		name    string
		params  kong2envoyParameters // cleanup temp file if created
		wantErr bool
		errMsg  string
	}{
		{
			name: "default",
			params: kong2envoyParameters{
				KongConfig:             "",
				Namespace:              "default",
				GatewayName:            "eg",
				GatewayClassName:       "eg",
				EnvoyGatewayDeployMode: "ControllerNamespace",
			},
			wantErr: false,
		},
		{
			name: "custom-config",
			params: kong2envoyParameters{
				KongConfig:             "testdata/custom-kong-config.yaml",
				Namespace:              "develop",
				GatewayName:            "my-gateway",
				GatewayClassName:       "my-gatewayclass",
				EnvoyGatewayDeployMode: "ControllerNamespace",
			},
			wantErr: false,
		},
		{
			name: "invalid-deploy-mode",
			params: kong2envoyParameters{
				KongConfig:             "",
				Namespace:              "default",
				GatewayName:            "eg",
				GatewayClassName:       "eg",
				EnvoyGatewayDeployMode: "InvalidMode",
			},
			wantErr: true,
			errMsg:  "invalid Envoy Gateway deployment mode: InvalidMode, must be either ControllerNamespace or GatewayNamespace",
		},
		{
			name: "gateway-namespace-mode",
			params: kong2envoyParameters{
				KongConfig:             "",
				Namespace:              "production",
				GatewayName:            "eg",
				GatewayClassName:       "eg",
				EnvoyGatewayDeployMode: "GatewayNamespace",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute the function
			output, err := kong2envoyConfig(tt.params)

			// Check error expectations
			if tt.wantErr {
				if err == nil {
					t.Errorf("kong2envoyConfig() expected error but got none")
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("kong2envoyConfig() error = %v, want error containing %q", err, tt.errMsg)
				}
				return
			}

			// Check no error expected
			if err != nil {
				t.Errorf("kong2envoyConfig() unexpected error = %v", err)
				return
			}

			// Compare with expected file
			expectedFilePath := "testdata/" + tt.name + ".expected.yaml"
			// #nosec G304
			expectedContent, readErr := os.ReadFile(expectedFilePath)
			if readErr != nil {
				t.Errorf("Failed to read expected file %s: %v (run with -overwrite to create)", expectedFilePath, readErr)
				if *overwrite {
					// Write actual output to expected file when overwrite flag is set
					// #nosec G306
					if err := os.WriteFile(expectedFilePath, []byte(output), 0o644); err != nil {
						t.Fatalf("Failed to write expected file %s: %v", expectedFilePath, err)
					}
					t.Logf("Updated expected file: %s", expectedFilePath)
					return
				}
				return
			}

			expectedStr := string(expectedContent)
			actualStr := output

			if expectedStr != actualStr {
				t.Errorf("Output does not match expected file %s", expectedFilePath)
				if *overwrite {
					// Write actual output to expected file when overwrite flag is set
					// #nosec G306
					if err := os.WriteFile(expectedFilePath, []byte(output), 0o44); err != nil {
						t.Fatalf("Failed to write expected file %s: %v", expectedFilePath, err)
					}
					t.Logf("Updated expected file: %s", expectedFilePath)
					return
				}
			}
		})
	}
}
