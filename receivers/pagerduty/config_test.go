package pagerduty

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/grafana/alerting/receivers"
	testing2 "github.com/grafana/alerting/receivers/testing"
	"github.com/grafana/alerting/templates"
)

func TestValidateConfig(t *testing.T) {
	hostName := "Grafana-TEST-host"
	provideHostName := func() (string, error) {
		return hostName, nil
	}
	original := getHostname
	t.Cleanup(func() {
		getHostname = original
	})
	getHostname = provideHostName

	cases := []struct {
		name              string
		settings          string
		secureSettings    map[string][]byte
		expectedConfig    Config
		expectedInitError string
		hostnameOverride  func() (string, error)
	}{
		{
			name:              "Error if empty",
			settings:          "",
			expectedInitError: `failed to unmarshal settings`,
		},
		{
			name:              "Error if empty JSON object",
			settings:          `{}`,
			expectedInitError: `could not find integration key property in settings`,
		},
		{
			name:     "Minimal valid configuration",
			settings: `{"integrationKey": "test-api-key" }`,
			expectedConfig: Config{
				Key:           "test-api-key",
				Severity:      DefaultSeverity,
				CustomDetails: defaultCustomDetails(),
				Class:         DefaultClass,
				Component:     "Grafana",
				Group:         DefaultGroup,
				Summary:       templates.DefaultMessageTitleEmbed,
				Source:        hostName,
				Client:        DefaultClient,
				ClientURL:     "{{ .ExternalURL }}",
			},
		},
		{
			name:     "Minimal valid configuration",
			settings: `{}`,
			secureSettings: map[string][]byte{
				"integrationKey": []byte("test-api-key"),
			},
			expectedConfig: Config{
				Key:           "test-api-key",
				Severity:      DefaultSeverity,
				CustomDetails: defaultCustomDetails(),
				Class:         DefaultClass,
				Component:     "Grafana",
				Group:         DefaultGroup,
				Summary:       templates.DefaultMessageTitleEmbed,
				Source:        hostName,
				Client:        DefaultClient,
				ClientURL:     "{{ .ExternalURL }}",
			},
		},
		{
			name:     "Should overwrite token from secrets",
			settings: `{ "integrationKey": "test" }`,
			secureSettings: map[string][]byte{
				"integrationKey": []byte("test-api-key"),
			},
			expectedConfig: Config{
				Key:           "test-api-key",
				Severity:      DefaultSeverity,
				CustomDetails: defaultCustomDetails(),
				Class:         DefaultClass,
				Component:     "Grafana",
				Group:         DefaultGroup,
				Summary:       templates.DefaultMessageTitleEmbed,
				Source:        hostName,
				Client:        DefaultClient,
				ClientURL:     "{{ .ExternalURL }}",
			},
		},
		{
			name: "All empty fields = minimal valid configuration",
			secureSettings: map[string][]byte{
				"integrationKey": []byte("test-api-key"),
			},
			settings: `{
				"integrationKey": "", 
				"severity" : "", 
				"class" : "", 
				"component": "", 
				"group": "", 
				"summary": "", 
				"source": "",
				"client" : "",
				"client_url": ""
			}`,
			expectedConfig: Config{
				Key:           "test-api-key",
				Severity:      DefaultSeverity,
				CustomDetails: defaultCustomDetails(),
				Class:         DefaultClass,
				Component:     "Grafana",
				Group:         DefaultGroup,
				Summary:       templates.DefaultMessageTitleEmbed,
				Source:        hostName,
				Client:        DefaultClient,
				ClientURL:     "{{ .ExternalURL }}",
			},
		},
		{
			name: "All empty fields = minimal valid configuration",
			secureSettings: map[string][]byte{
				"integrationKey": []byte("test-api-key"),
			},
			settings: `{
				"integrationKey": "", 
				"severity" : "test-severity", 
				"class" : "test-class", 
				"component": "test-component", 
				"group": "test-group", 
				"summary": "test-summary", 
				"source": "test-source",
				"client" : "test-client",
				"client_url": "test-client-url"
			}`,
			expectedConfig: Config{
				Key:           "test-api-key",
				Severity:      "test-severity",
				CustomDetails: defaultCustomDetails(),
				Class:         "test-class",
				Component:     "test-component",
				Group:         "test-group",
				Summary:       "test-summary",
				Source:        "test-source",
				Client:        "test-client",
				ClientURL:     "test-client-url",
			},
		},
		{
			name: "Should ignore custom details",
			secureSettings: map[string][]byte{
				"integrationKey": []byte("test-api-key"),
			},
			settings: `{
				"custom_details" : {
					"test" : "test"
				},
				"CustomDetails" : {
					"test" : "test"
				}
			}`,
			expectedConfig: Config{
				Key:           "test-api-key",
				Severity:      DefaultSeverity,
				CustomDetails: defaultCustomDetails(),
				Class:         DefaultClass,
				Component:     "Grafana",
				Group:         DefaultGroup,
				Summary:       templates.DefaultMessageTitleEmbed,
				Source:        hostName,
				Client:        DefaultClient,
				ClientURL:     "{{ .ExternalURL }}",
			},
		},
		{
			name: "Source should fallback to client if hostname cannot be resolved",
			secureSettings: map[string][]byte{
				"integrationKey": []byte("test-api-key"),
			},
			settings: `{
				"client" : "test-client"
			}`,
			hostnameOverride: func() (string, error) {
				return "", errors.New("test")
			},
			expectedConfig: Config{
				Key:           "test-api-key",
				Severity:      DefaultSeverity,
				CustomDetails: defaultCustomDetails(),
				Class:         DefaultClass,
				Component:     "Grafana",
				Group:         DefaultGroup,
				Summary:       templates.DefaultMessageTitleEmbed,
				Source:        "test-client",
				Client:        "test-client",
				ClientURL:     "{{ .ExternalURL }}",
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if c.hostnameOverride != nil {
				getHostname = c.hostnameOverride
				t.Cleanup(func() {
					getHostname = provideHostName
				})
			}
			m := &receivers.NotificationChannelConfig{
				Settings:       json.RawMessage(c.settings),
				SecureSettings: c.secureSettings,
			}
			fc, err := testing2.NewFactoryConfigForValidateConfigTesting(t, m)
			require.NoError(t, err)

			actual, err := ValidateConfig(fc)

			if c.expectedInitError != "" {
				require.ErrorContains(t, err, c.expectedInitError)
				return
			}
			require.NoError(t, err)
			require.Equal(t, c.expectedConfig, *actual)
		})
	}
}
