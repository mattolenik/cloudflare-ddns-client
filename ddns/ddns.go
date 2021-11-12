package ddns

import (
	"context"
	"time"

	"github.com/juju/errors"
	"github.com/mattolenik/cloudflare-ddns-client/cloudflare"
	"github.com/mattolenik/cloudflare-ddns-client/conf"
	"github.com/mattolenik/cloudflare-ddns-client/ip"
	"github.com/rs/zerolog/log"
)

// Run performs a one time DDNS update.
func Run(ctx context.Context) error {
	ip, err := ip.GetPublicIPWithRetry(10, 5*time.Second)
	if err != nil {
		return errors.Annotate(err, "unable to retrieve public IP")
	}
	log.Info().Msgf("Found public IP '%s'", ip)
	err = cloudflare.Update(
		ctx,
		conf.Token.Get(),
		conf.Domain.Get(),
		conf.Record.Get(),
		ip)
	return errors.Trace(err)
}

// DaemonWithDefaults calls Daemon but with default values
// updatePeriod      - 10 seconds
// failureRetryDelay - 10 seconds
func DaemonWithDefaults(ctx context.Context) error {
	d := 10 * time.Second
	return errors.Trace(Daemon(ctx, d, d))
}

// Daemon continually keeps DDNS up to date.
// updatePeriod      - how often to check for updates
// failureRetryDelay - how long to wait until retry after a failure
func Daemon(ctx context.Context, updatePeriod, failureRetryDelay time.Duration) error {
	var lastIP string
	var lastIPUpdate time.Time

	log.Info().Msgf("Daemon running, will now monitor for IP updates every %d seconds", int(updatePeriod.Seconds()))

	for {
		dnsRecordIP, err := cloudflare.Get(
			ctx,
			conf.Token.Get(),
			conf.Domain.Get(),
			conf.Record.Get())
		if err != nil {
			log.Error().Msgf("unable to look up current DNS record, will retry in %d seconds", int(updatePeriod.Seconds()))
			time.Sleep(failureRetryDelay)
			continue
		}
		newIP, err := ip.GetPublicIPWithRetry(10, 5*time.Second)
		if err != nil {
			log.Error().Msgf("unable to retrieve public IP, will retry in %d seconds", int(updatePeriod.Seconds()))
			time.Sleep(failureRetryDelay)
			continue
		}
		if newIP == lastIP && newIP == dnsRecordIP {
			log.Info().Msgf(
				"No IP change detected since %s (%d seconds ago)",
				lastIPUpdate.Format(time.RFC1123Z),
				int(time.Since(lastIPUpdate).Seconds()))
			time.Sleep(updatePeriod)
			continue
		}
		if lastIP == "" {
			// Log line for first time
			log.Info().Msgf("Found public IP '%s'", newIP)
		} else if newIP != lastIP {
			// Log line for IP change
			log.Info().Msgf("Detected new public IP address, it changed from '%s' to '%s'", lastIP, newIP)
		} else if dnsRecordIP != newIP {
			// Log line for no new IP, but mismatch with DNS record
			log.Info().Msgf("Public IP address did not change, but DNS record did match, is '%s' but expected '%s', correcting", dnsRecordIP, newIP)
		}
		lastIP = newIP
		lastIPUpdate = time.Now()

		err = cloudflare.Update(
			ctx,
			conf.Token.Get(),
			conf.Domain.Get(),
			conf.Record.Get(),
			lastIP)
		if err != nil {
			log.Error().Msgf("unable to update DDNS, will retry in %d seconds", updatePeriod/time.Second)
			time.Sleep(failureRetryDelay)
			continue
		}
		time.Sleep(updatePeriod)
	}
}
