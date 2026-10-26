package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateHost(t *testing.T) {
	valid := map[string]string{
		"https://api.nowsecure.com":  "https://api.nowsecure.com",
		"https://foo.nowsecure.io/":  "https://foo.nowsecure.io",
		"https://nowsecure.com":      "https://nowsecure.com",
		"https://API.NowSecure.com":  "https://API.NowSecure.com",
		"http://localhost:8080":      "http://localhost:8080",
		"https://127.0.0.1":          "https://127.0.0.1",
		"http://[::1]:9000":          "http://[::1]:9000",
		"https://api.nowsecure.com/": "https://api.nowsecure.com",
	}

	for input, expected := range valid {
		t.Run("accepts "+input, func(t *testing.T) {
			actual, err := validateHost("api_host", input)
			require.NoError(t, err)
			assert.Equal(t, expected, actual)
		})
	}

	invalid := []string{
		"",
		"api.nowsecure.com",
		"http://api.nowsecure.com",
		"ftp://api.nowsecure.com",
		"https://evil.com",
		"https://evilnowsecure.com",
		"https://nowsecure.com.evil.com",
		"https://user:pw@api.nowsecure.com",
		"https://user@localhost",
	}

	for _, input := range invalid {
		t.Run("rejects "+input, func(t *testing.T) {
			_, err := validateHost("api_host", input)
			require.ErrorContains(t, err, "api_host")
		})
	}
}
