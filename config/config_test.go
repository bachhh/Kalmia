package config_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"git.difuse.io/Difuse/kalmia/config"
	"github.com/stretchr/testify/assert"
)

func TestCookieConfig_UnmarshalJSON(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name          string
		jsonData      []byte
		expected      config.CookieConfig
		expectErr     bool
		expectedError string
	}{
		{
			name:     "Valid SameSite Lax",
			jsonData: []byte(`{"domain": "example.com", "path": "/", "age": 7, "secure": true, "sameSite": "Lax", "httpOnly": true}`),
			expected: config.CookieConfig{
				Domain:   "example.com",
				Path:     "/",
				AgeDays:  7,
				Secure:   true,
				SameSite: config.SameSite(http.SameSiteLaxMode),
				HttpOnly: true,
			},
			expectErr: false,
		},
		{
			name:     "Valid SameSite Strict",
			jsonData: []byte(`{"domain": "example.com", "path": "/", "age": 7, "secure": true, "sameSite": "Strict", "httpOnly": true}`),
			expected: config.CookieConfig{
				Domain:   "example.com",
				Path:     "/",
				AgeDays:  7,
				Secure:   true,
				SameSite: config.SameSite(http.SameSiteStrictMode),
				HttpOnly: true,
			},
			expectErr: false,
		},
		{
			name:     "Valid SameSite None",
			jsonData: []byte(`{"domain": "example.com", "path": "/", "age": 7, "secure": true, "sameSite": "None", "httpOnly": true}`),
			expected: config.CookieConfig{
				Domain:   "example.com",
				Path:     "/",
				AgeDays:  7,
				Secure:   true,
				SameSite: config.SameSite(http.SameSiteNoneMode),
				HttpOnly: true,
			},
			expectErr: false,
		},
		{
			name:          "Invalid SameSite",
			jsonData:      []byte(`{"sameSite": "Invalid"}`),
			expectErr:     true,
			expectedError: `unknown SameSite value: "Invalid"`,
		},
		{
			name:      "Case Insensitive SameSite",
			jsonData:  []byte(`{"sameSite": "lAx"}`),
			expected:  config.CookieConfig{SameSite: config.SameSite(http.SameSiteLaxMode)},
			expectErr: false,
		},
		{
			name:      "Empty JSON",
			jsonData:  []byte(`{}`),
			expected:  config.CookieConfig{},
			expectErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var cfg config.CookieConfig
			err := json.Unmarshal(tc.jsonData, &cfg)

			if tc.expectErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, cfg)
			}
		})
	}
}
