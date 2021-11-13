package ddns

import (
	"time"

	"github.com/juju/errors"
	"github.com/mattolenik/cloudflare-ddns-client/conf"
	"github.com/mattolenik/cloudflare-ddns-client/ip"
	"github.com/rs/zerolog/log"
)

type DDNSProvider interface {
	Get(domain, record string) (string, error)
	Update(domain, record, ip string) error
}

// Run performs a one time DDNS update.
func Run(provider DDNSProvider) error {
	ip, err := ip.GetPublicIPWithRetry(10, 5*time.Second)
	if err != nil {
		return errors.Annotate(err, "unable to retrieve public IP")
	}
	log.Info().Msgf("Found public IP '%s'", ip)
	err = provider.Update(conf.Domain.Get(), conf.Record.Get(), ip)
	return errors.Annotatef(err, "failed to update DNS")
}

// DaemonWithDefaults calls Daemon but with default values
func DaemonWithDefaults(provider DDNSProvider) error {
	d := 10 * time.Second
	return errors.Trace(Daemon(provider, d, d))
}

// Daemon continually keeps DDNS up to date.
// updatePeriod      - how often to check for updates
// failureRetryDelay - how long to wait until retry after a failure
func Daemon(provider DDNSProvider, updatePeriod, failureRetryDelay time.Duration) error {
	var lastIP string
	var lastIPUpdate time.Time

	log.Info().Msgf("Daemon running, will now monitor for IP updates every %d seconds", int(updatePeriod.Seconds()))

	for {
		dnsRecordIP, err := provider.Get(conf.Domain.Get(), conf.Record.Get())
		if err != nil {
			log.Error().Msgf("Unable to look up current DNS record, will retry in %d seconds. Error was:\n%v", int(updatePeriod.Seconds()), err)
			time.Sleep(failureRetryDelay)
			continue
		}
		newIP, err := ip.GetPublicIPWithRetry(10, 5*time.Second)
		if err != nil {
			log.Error().Msgf("Unable to retrieve public IP, will retry in %d seconds. Error was:\n%v", int(updatePeriod.Seconds()), err)
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

		err = provider.Update(conf.Domain.Get(), conf.Record.Get(), lastIP)
		if err != nil {
			log.Error().Msgf("Unable to update DNS, will retry in %d seconds. Erorr was:\n%v", updatePeriod/time.Second, err)
			time.Sleep(failureRetryDelay)
			continue
		}
		time.Sleep(updatePeriod)
	}
}
