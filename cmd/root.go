package cmd

import (
	"os"
	"strings"
	"time"

	"github.com/juju/errors"
	"github.com/mattolenik/cloudflare-ddns-client/conf"
	"github.com/mattolenik/cloudflare-ddns-client/dns"
	"github.com/mattolenik/cloudflare-ddns-client/errhandler"
	"github.com/mattolenik/cloudflare-ddns-client/ip"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var configFile string

// Root is the root command of the program
var Root = &cobra.Command{
	Use:   "cloudflare-ddns",
	Short: "Update a CloudFlare DNS record with your public IP",
	Long: `A dynamic DNS client for CloudFlare. Automatically detects your public IP and
creates/updates a DNS record in CloudFlare.

Configuration flags can be set by defining an environment variable of the same name.
For example:
` + "DOMAIN=mydomain.com RECORD=sub.mydomain.com TOKEN=<api-token> cloudflare-ddns" + `
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ip, err := ip.GetPublicIPWithRetry(5, 30*time.Second)
		if err != nil {
			return errors.Annotate(err, "unable to retrieve public IP")
		}
		log.Info().Msgf("Found public IP '%s'", ip)
		return dns.UpdateCloudFlare(
			viper.GetString(conf.Token),
			viper.GetString(conf.Domain),
			viper.GetString(conf.Record),
			ip)
	},
	Version: conf.Version,
}

func init() {
	Root.PersistentFlags().StringVar(&configFile, conf.Config, "", "Path to config file (default is /etc/cloudflare-ddns.toml)")
	Root.PersistentFlags().String(conf.Domain, "", "Domain name in CloudFlare, e.g. example.com")
	Root.PersistentFlags().String(conf.Record, "", "DNS record name in CloudFlare, may be subdomain or same as domain")
	Root.PersistentFlags().String(conf.Token, "", "CloudFlare API token with permissions Zone:Zone:Read and Zone:DNS:Edit")
	Root.PersistentFlags().Bool(conf.JSONOutput, false, "Log format, either pretty or json, defaults to pretty")
	Root.PersistentFlags().BoolP(conf.Verbose, conf.VerboseShort, false, "Verbose logging, prints additional log output")
	Root.SetVersionTemplate("{{.Version}}\n")

	viper.BindPFlag(conf.Config, Root.PersistentFlags().Lookup(conf.Config))
	viper.BindPFlag(conf.Domain, Root.PersistentFlags().Lookup(conf.Domain))
	viper.BindPFlag(conf.Record, Root.PersistentFlags().Lookup(conf.Record))
	viper.BindPFlag(conf.Token, Root.PersistentFlags().Lookup(conf.Token))
	viper.BindPFlag(conf.JSONOutput, Root.PersistentFlags().Lookup(conf.JSONOutput))

	viper.SetDefault(conf.JSONOutput, false)
	viper.SetDefault(conf.Config, "/etc/cloudflare-ddns.toml")
	viper.SetDefault(conf.Verbose, false)

	cobra.OnInitialize(initConfig)
}

func initConfig() {
	if configFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(configFile)
	} else {
		viper.AddConfigPath(".")
		viper.AddConfigPath("$HOME/.config")
		viper.AddConfigPath("$HOME")
		viper.AddConfigPath("/etc")
		viper.SetConfigName("cloudflare-ddns.toml")
	}
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	// Config file is optional, ignore errors
	err := viper.ReadInConfig()
	if !viper.GetBool(conf.JSONOutput) {
		writer := zerolog.ConsoleWriter{Out: os.Stderr}
		log.Logger = log.Output(writer)
	}
	if viper.GetBool(conf.Verbose) {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
	// Config file is optional, continue if not found, unless config was specified by user and still not found
	_, notFound := err.(viper.ConfigFileNotFoundError)
	if !(notFound && configFile == "") {
		errhandler.Handle(err)
	}
}
