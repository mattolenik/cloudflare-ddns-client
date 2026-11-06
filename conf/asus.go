package conf

import (
	"os/exec"
	"strings"

	"github.com/juju/errors"
	"github.com/spf13/viper"
	"golang.org/x/net/publicsuffix"
)

const AsusNVRAMConfig = "asus-nvram"

// LoadAsusNVRAM sets record, domain, and token from the Asus router's DDNS nvram settings.
func LoadAsusNVRAM() error {
	record, err := nvramGet("ddns_hostname_x")
	if err != nil {
		return err
	}
	if record == "" {
		return errors.New("nvram value 'ddns_hostname_x' is not set. Set the host name in the router's WAN > DDNS settings")
	}
	token, err := nvramGet("ddns_cloudflare_api_token")
	if err != nil {
		return err
	}
	if token == "" {
		return errors.New("nvram value 'ddns_cloudflare_api_token' is not set. SSH into the router and run: nvram set ddns_cloudflare_api_token=<your-cloudflare-api-token> && nvram commit")
	}
	domain, err := publicsuffix.EffectiveTLDPlusOne(record)
	if err != nil {
		return errors.Annotatef(err, "could not determine domain from ddns_hostname_x '%s'", record)
	}
	viper.Set(Record.Name, record)
	viper.Set(Domain.Name, domain)
	viper.Set(Token.Name, token)
	return nil
}

// nvramGet reads a value from nvram. Unset keys come back blank (nvram still exits 0).
func nvramGet(key string) (string, error) {
	out, err := exec.Command("nvram", "get", key).Output()
	if err != nil {
		return "", errors.Annotatef(err, "failed to run 'nvram get %s'", key)
	}
	return strings.TrimSpace(string(out)), nil
}

// AsusLogWriter sends each log line to the Asus system log via `logger`.
type AsusLogWriter struct{}

func (AsusLogWriter) Write(p []byte) (int, error) {
	err := exec.Command("logger", "-t", "cloudflare-ddns-client", strings.TrimSpace(string(p))).Run()
	return len(p), err
}
