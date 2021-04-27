package ddns

import (
	"time"

	"github.com/juju/errors"
	"github.com/mattolenik/cloudflare-ddns-client/conf"
	"github.com/mattolenik/cloudflare-ddns-client/dns"
	"github.com/mattolenik/cloudflare-ddns-client/ip"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func Run() error {
	ip, err := ip.GetPublicIPWithRetry(10, 5*time.Second)
	if err != nil {
		return errors.Annotate(err, "unable to retrieve public IP")
	}
	log.Info().Msgf("Found public IP '%s'", ip)
	return dns.UpdateCloudFlare(
		viper.GetString(conf.Token),
		viper.GetString(conf.Domain),
		viper.GetString(conf.Record),
		ip)
}
