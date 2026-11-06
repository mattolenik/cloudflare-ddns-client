package conf

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// useMockRouter puts the mock nvram and logger scripts first on PATH, mimicking an Asus router.
// Skipped on Windows, which can't run the shell script mocks (and isn't an Asus router).
func useMockRouter(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Asus router mocks are shell scripts")
	}
	dir, err := filepath.Abs("../test/mocks/scripts")
	require.NoError(t, err)
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

func TestLoadAsusNVRAM(t *testing.T) {
	useMockRouter(t)
	tests := []struct {
		name, hostname, token string
		wantDomain, wantErr   string
	}{
		{name: "subdomain", hostname: "abc.example.com", token: "tok", wantDomain: "example.com"},
		{name: "bare domain", hostname: "example.com", token: "tok", wantDomain: "example.com"},
		{name: "multi-part suffix", hostname: "abc.example.co.uk", token: "tok", wantDomain: "example.co.uk"},
		{name: "whitespace trimmed", hostname: " abc.example.com \n", token: "tok\n", wantDomain: "example.com"},
		{name: "hostname unset", token: "tok", wantErr: "nvram value 'ddns_hostname_x' is not set"},
		{name: "token unset", hostname: "abc.example.com", wantErr: "nvram set ddns_cloudflare_api_token="},
		{name: "both unset", wantErr: "is not set"},
		{name: "bare suffix", hostname: "com", token: "tok", wantErr: "could not determine domain"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()
			t.Setenv("NVRAM_ddns_hostname_x", tt.hostname)
			t.Setenv("NVRAM_ddns_cloudflare_api_token", tt.token)
			err := LoadAsusNVRAM()
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, strings.TrimSpace(tt.hostname), Record.Get())
			assert.Equal(t, "tok", Token.Get())
			assert.Equal(t, tt.wantDomain, Domain.Get())
		})
	}
}

func TestAsusLogWriter(t *testing.T) {
	useMockRouter(t)
	out := filepath.Join(t.TempDir(), "syslog")
	t.Setenv("LOGGER_OUT", out)
	_, err := AsusLogWriter{}.Write([]byte("hello world\n"))
	require.NoError(t, err)
	b, err := os.ReadFile(out)
	require.NoError(t, err)
	assert.Equal(t, "cloudflare-ddns-client hello world\n", string(b))
}
