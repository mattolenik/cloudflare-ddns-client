package ddns

//go:generate mockgen -destination=../mocks/mock_ddns.go -package=mocks -source=ddns.go

import (
	"time"

	"github.com/juju/errors"
	"github.com/mattolenik/cloudflare-ddns-client/conf"
	"github.com/mattolenik/cloudflare-ddns-client/ip"
	"github.com/mattolenik/cloudflare-ddns-client/task"
)

type DDNSProvider interface {
	Get(domain, record string) (string, error)
	Update(domain, record, ip string) error
}

type IPProvider interface {
	Get() (ip string, err error)
}

type DefaultIPProvider struct{}

func (p *DefaultIPProvider) Get() (string, error) {
	ip, err := ip.GetPublicIPWithRetry(10, 5*time.Second)
	if err != nil {
		return "", errors.Trace(err)
	}
	return ip, nil
}

func NewDefaultIPProvider() *DefaultIPProvider {
	return &DefaultIPProvider{}
}

type ConfigProvider interface {
	Get() (domain, record string, err error)
}

type DefaultConfigProvider struct{}

func (p *DefaultConfigProvider) Get() (domain, record string, err error) {
	return conf.Domain.Get(), conf.Record.Get(), nil
}

func NewDefaultConfigProvider() *DefaultConfigProvider {
	return &DefaultConfigProvider{}
}

type Daemon[T any] interface {
	Update() error
	Start(updatePeriod, retryDelay time.Duration) task.StatusStream[T]
	Stop()
}

type DDNSDaemon struct {
	Daemon[StatusInfo]
	shouldRun      bool
	ddnsProvider   DDNSProvider
	ipProvider     IPProvider
	configProvider ConfigProvider

	BeforeUpdate func()
	AfterUpdate  func()
}

// NewDefaultDaemon creates a new DDNSDaemon
func NewDefaultDaemon(ddnsProvider DDNSProvider, ipProvider IPProvider, configProvider ConfigProvider) *DDNSDaemon {
	if ddnsProvider == nil {
		panic("ddnsProvider must not be nil")
	}
	if ipProvider == nil {
		panic("ipProvider must not be nil")
	}
	if configProvider == nil {
		panic("configProvider must not be nil")
	}
	return &DDNSDaemon{
		shouldRun:      true,
		ddnsProvider:   ddnsProvider,
		ipProvider:     ipProvider,
		configProvider: configProvider,
	}
}

// Update performs a one time DDNS update.
func (d *DDNSDaemon) Update() error {
	ip, err := d.ipProvider.Get()
	if err != nil {
		return errors.Annotate(err, "unable to retrieve public IP")
	}
	domain, record, err := d.configProvider.Get()
	if err != nil {
		return errors.Annotate(err, "unable to find domain or record in configuration")
	}
	err = d.ddnsProvider.Update(domain, record, ip)
	return errors.Annotatef(err, "failed to update DNS")
}

type StatusInfo struct {
	AtEndOfUpdate bool
}

// Start continually keeps DDNS up to date, asynchronously in a new goroutine.
// updatePeriod - how often to check for updates
// retryDelay   - how long to wait until retry after a failure
func (d *DDNSDaemon) Start(updatePeriod, retryDelay time.Duration) (status task.StatusStream[StatusInfo]) {
	var lastIP string
	var lastIPUpdate time.Time

	status.Infof("Daemon running, will now monitor for IP updates every %d seconds", int(updatePeriod.Seconds()))

	func() {
		defer close(status)
		for d.shouldRun {
			domain, record, err := d.configProvider.Get()
			if err != nil {
				status.Fatal(errors.Annotate(err, "unable to find domain or record in configuration"))
				return
			}
			dnsRecordIP, err := d.ddnsProvider.Get(domain, record)
			if err != nil {
				status.Error(errors.Annotatef(err, "Unable to look up current DNS record, will retry in %d seconds. Error was:\n%v", int(updatePeriod.Seconds())))
				time.Sleep(retryDelay)
				continue
			}
			newIP, err := d.ipProvider.Get()
			if err != nil {
				status.Errorf("Unable to retrieve public IP, will retry in %d seconds. Error was:\n%v", int(updatePeriod.Seconds()), err)
				time.Sleep(retryDelay)
				continue
			}

			// Nothing has changed, log and move on
			if newIP == lastIP && newIP == dnsRecordIP {
				status.Infof(
					"No IP change detected since %s (%d seconds ago)",
					lastIPUpdate.Format(time.RFC1123Z),
					int(time.Since(lastIPUpdate).Seconds()))
				time.Sleep(updatePeriod)
				continue
			}

			// IP has changed, log depending on how it has changed
			if lastIP == "" {
				// Log line for first time
				status.Infof("Found public IP '%s'", newIP)
			} else if newIP != lastIP {
				// Log line for IP change
				status.Infof("Detected new public IP address, it changed from '%s' to '%s'", lastIP, newIP)
			} else if dnsRecordIP != newIP {
				// Log line for no new IP, but mismatch with DNS record
				status.Infof("Public IP address did not change, but DNS record did match, is '%s' but expected '%s', correcting", dnsRecordIP, newIP)
			}

			lastIP = newIP
			lastIPUpdate = time.Now()

			// Reach out to the actual DDNS provider and make the update
			err = d.ddnsProvider.Update(domain, record, lastIP)
			if err != nil {
				status.Errorf("Unable to update DNS, will retry in %d seconds. Erorr was:\n%v", updatePeriod/time.Second, err)
				time.Sleep(retryDelay)
				continue
			}

			status.Msgf(StatusInfo{AtEndOfUpdate: true}, "Successfully updated DNS")

			// Do another run check before the sleep occurs so as to not draw out the stop operation
			if !d.shouldRun {
				status.Msgf(StatusInfo{AtEndOfUpdate: false}, "Daemon stopped")
				return
			}
			time.Sleep(updatePeriod)
		}
	}()
	return status
}

// StartWithDefaults calls Start but with default values
func (d *DDNSDaemon) StartWithDefaults() (status task.StatusStream[StatusInfo]) {
	t := 10 * time.Second
	return d.Start(t, t)
}

// Stop instructs the daemon to stop as soon as the current (if any) operation is finished
func (d *DDNSDaemon) Stop() {
	d.shouldRun = true
}
