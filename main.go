package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/cloudflare/cloudflare-go"
	"github.com/juju/errors"
	"github.com/mattolenik/cloudflare-ddns-client/ip"
	"github.com/pelletier/go-toml"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// version is populated by the ldflags argument during build.
var version string

const configFileName = "cloudflare-ddns.conf"

var flagLogFormat string

// Config represents TOML program configuration
type Config struct {
	// Domain name of the DNS record
	Domain string `toml:"domain"`

	// DNS record to update
	Record string `toml:"record"`

	// CloudFlare API token, must have permissions Zone:Zone:Read, Zone:DNS:Edit
	Token string `toml:"token"`

	// Log output, either "pretty" or "json"
	LogFormat string `toml:"log_format" default:"pretty"`

	// Print verbose debug logs
	LogVerbose bool `toml:"log_verbose" default:"false"`
}

func main() {
	err := mainE()
	if err != nil {
		var msg string
		if e, ok := err.(*errors.Err); ok {
			trace := e.StackTrace()
			if flagLogFormat == "pretty" {
				msg = strings.Join(trace, "\n")
			} else {
				msg = strings.Join(trace, " \\ ")
			}
		}
		log.Error().Msg(msg)
		os.Exit(1)
	}
}

func mainE() error {
	// Setting arg 0 makes sure that -help output has the correct program name when being invoked with "go run"
	os.Args[0] = "cloudflare-ddns"
	var flagVersion bool
	var flagVerbose bool
	var flagConfigPath string

	flag.BoolVar(&flagVersion, "version", false, "Print the program version")
	flag.BoolVar(&flagVerbose, "verbose", false, "Print debug logs")
	flag.StringVar(&flagConfigPath, "config", "/etc/"+configFileName, "Path to configuration file")
	flag.StringVar(&flagLogFormat, "log-format", "pretty", "Log output format, either json or pretty")

	flag.Parse()

	if flagLogFormat == "pretty" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}
	if flagVersion {
		printVersion()
	}
	c, err := loadConfig(flagConfigPath)
	if err != nil {
		return errors.Trace(err)
	}

	// Set log level to verbose if either flag or config values are set
	flagVerbose = flagVerbose || c.LogVerbose
	if flagVerbose {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	ip, err := ip.GetExternalIP()
	if err != nil {
		return errors.Annotate(err, "unable to retrieve external IP")
	}
	log.Info().Msgf("Found external IP '%s'", ip)
	err = UpdateCloudFlare(c)
	return errors.Annotate(err, "failed to update CloudFlare")
}

// UpdateCloudFlare updates the CloudFlare DNS record
func UpdateCloudFlare(c *Config) error {
	api, err := cloudflare.NewWithAPIToken(c.Token)
	if err != nil {
		return errors.Annotate(err, "unable to connect to CloudFlare, token may be invalid")
	}
	// Get the zone ID for the domain
	zoneID, err := api.ZoneIDByName(c.Domain)
	if err != nil {
		return errors.Annotatef(err, "unable to retrieve zone ID for domain '%s' from CloudFlare", c.Domain)
	}
	// Get the record ID
	records, err := api.DNSRecords(zoneID, cloudflare.DNSRecord{Type: "A"})
	if err != nil {
		return errors.Annotate(err, "unable to retrieve zone ID from CloudFlare")
	}
	// Find the specific record
	var recordID string
	for _, record := range records {
		log.Debug().Msgf("Examining DNS record ID '%s' with name '%s'", record.ID, record.Name)
		if record.Name == c.Record {
			recordID = record.ID
			if record.Content == ip {
				log.Info().Msgf("DNS record '%s' is already set to IP '%s'", c.Record, ip)
				return nil
			}
			break
		}
	}
	// Create the record if it's not already there
	if recordID == "" {
		log.Info().Msgf("No DNS '%s' found for domain '%s', creating now", c.Record, c.Domain)
		resp, err := api.CreateDNSRecord(zoneID, cloudflare.DNSRecord{
			Content: ip,
			Type:    "A",
			Name:    c.Record,
		})
		if err != nil {
			return errors.Annotatef(err, "failed to create DNS record '%s' on domain '%s'", c.Record, c.Domain)
		}
		recordID = resp.Result.ID
	}
	// Update the record
	err = api.UpdateDNSRecord(zoneID, recordID, cloudflare.DNSRecord{
		Content: ip,
		Type:    "A",
	})
	if err != nil {
		return errors.Annotatef(err, "failed to update DNS record '%s' to IP address '%s'", c.Record, ip)
	}

	log.Info().Msgf("Successfully updated DNS record '%s' to point to '%s'", c.Record, ip)
	return nil
}

// loadConfig loads the TOML configuration at the specified path.
func loadConfig(path string) (*Config, error) {
	log.Debug().Msgf("Loading config file '%s'", path)
	config := &Config{}
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Annotatef(err, "failed to read config file '%s", path)
	}
	err = toml.Unmarshal(contents, config)
	if err != nil {
		return nil, errors.Annotatef(err, "failed to parse config file '%s", path)
	}
	return config, nil
}

// printVersion prints the program version and exits.
func printVersion() {
	fmt.Println(version)
	os.Exit(0)
}
