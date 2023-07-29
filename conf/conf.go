package conf

import (
	"fmt"

	"github.com/mattolenik/cloudflare-ddns-client/meta"
)

var (
	ConfigFile            string // Path to the config file, if any
	DefaultConfigFilename = "cloudflare-ddns.toml"

	Config = StringOption{
		Name:        "config",
		Description: fmt.Sprintf("Path to config file. If not specified will look for %s in the program dir (%s), $HOME/.config, or /etc, in that order", DefaultConfigFilename, meta.ProgramDir),
	}
	Daemon = BoolOption{
		Name:        "daemon",
		Default:     false,
		Description: "Run as a service, continually monitoring for IP changes",
	}
	Domain = StringOption{
		Name:        "domain",
		Description: "Domain name in CloudFlare, e.g. example.com",
	}
	IP = StringOption{
		Name:        "ip",
		Description: "An already known WAN IP, will not perform lookup",
	}
	Record = StringOption{
		Name:        "record",
		Description: "DNS record name in CloudFlare, may be subdomain or same as domain",
	}
	Token = StringOption{
		Name:        "token",
		Description: "CloudFlare API token with permissions Zone:Zone:Read and Zone:DNS:Edit",
	}
	JSONOutput = StringOption{
		Name:        "log-format",
		Default:     "pretty",
		Description: "Log format, either pretty or json, defaults to pretty",
	}
	Verbose = BoolOptionP{
		Name:        "verbose",
		ShortName:   "v",
		Description: "Verbose logging, prints additional log output",
		Default:     true,
	}
)
