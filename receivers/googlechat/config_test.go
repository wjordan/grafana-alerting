package googlechat

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/grafana/alerting/receivers"
	testing2 "github.com/grafana/alerting/receivers/testing"
	"github.com/grafana/alerting/templates"
)

func TestValidateConfig(t *testing.T) {
	cases := []struct {
		name              string
		settings          string
		expectedConfig    Config
		expectedInitError string
	}{
		{
			name:              "Error if empty",
			settings:          "",
			expectedInitError: `failed to unmarshal settings`,
		},
		{
			name:              "Error if empty JSON object",
			settings:          `{}`,
			expectedInitError: `could not find url property in settings`,
		},
		{
			name:              "Error if URL is empty",
			settings:          `{ "url": "" }`,
			expectedInitError: `could not find url property in settings`,
		},
		{
			name:     "Minimal valid configuration",
			settings: `{"url": "http://localhost"}`,
			expectedConfig: Config{
				Title:   templates.DefaultMessageTitleEmbed,
				Message: templates.DefaultMessageEmbed,
				URL:     "http://localhost",
			},
		},
		{
			name:     "All empty fields = minimal valid configuration",
			settings: `{"url": "http://localhost", "title": "", "message": "", "avatar_url" : "", "use_discord_username": null}`,
			expectedConfig: Config{
				Title:   templates.DefaultMessageTitleEmbed,
				Message: templates.DefaultMessageEmbed,
				URL:     "http://localhost",
			},
		},
		{
			name:     "Extracts all fields",
			settings: `{"url": "http://localhost", "title": "test-title", "message": "test-message", "avatar_url" : "http://avatar", "use_discord_username": true}`,
			expectedConfig: Config{
				Title:   "test-title",
				Message: "test-message",
				URL:     "http://localhost",
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			m := &receivers.NotificationChannelConfig{
				Settings: json.RawMessage(c.settings),
			}
			fc, err := testing2.NewFactoryConfigForValidateConfigTesting(t, m)
			require.NoError(t, err)

			actual, err := ValidateConfig(fc)

			if c.expectedInitError != "" {
				require.ErrorContains(t, err, c.expectedInitError)
				return
			}
			require.Equal(t, c.expectedConfig, *actual)
		})
	}
}
