package providers

import (
	"context"

	"github.com/cloudflare/cloudflare-go"
	"github.com/juju/errors"
	"github.com/mattolenik/cloudflare-ddns-client/task"
)

type CloudFlareProvider struct {
	client *cloudflare.API
	ctx    context.Context
	status task.StatusStream[any]
}

func NewCloudFlareProvider(ctx context.Context, status task.StatusStream[any], apiToken string) (*CloudFlareProvider, error) {
	api, err := cloudflare.NewWithAPIToken(apiToken)
	if err != nil {
		return nil, errors.Annotate(err, "unable to connect to CloudFlare, token may be invalid")
	}
	return &CloudFlareProvider{client: api, ctx: ctx}, nil
}

// Get fetches the IP of the given record, returning empty string if it doesn't exist
func (p *CloudFlareProvider) Get(domain, record string) (string, error) {
	// Get the zone ID for the domain
	zoneID, err := p.client.ZoneIDByName(domain)
	if err != nil {
		return "", errors.Annotatef(err, "unable to retrieve zone ID for domain '%s' from CloudFlare", domain)
	}
	// Get the record ID
	records, err := p.client.DNSRecords(p.ctx, zoneID, cloudflare.DNSRecord{Type: "A"})
	if err != nil {
		return "", errors.Annotate(err, "unable to retrieve zone ID from CloudFlare")
	}
	// Find the specific record
	for _, r := range records {
		p.status.Infof("Examining DNS record ID '%s' with name '%s'", r.ID, r.Name)
		if r.Name == record {
			return r.Content, nil
		}
	}
	return "", nil
}

// Update updates the CloudFlare DNS record
func (p *CloudFlareProvider) Update(domain, record, ip string) error {
	// Get the zone ID for the domain
	zoneID, err := p.client.ZoneIDByName(domain)
	if err != nil {
		return errors.Annotatef(err, "unable to retrieve zone ID for domain '%s' from CloudFlare", domain)
	}
	// Get the record ID
	records, err := p.client.DNSRecords(p.ctx, zoneID, cloudflare.DNSRecord{Type: "A"})
	if err != nil {
		return errors.Annotate(err, "unable to retrieve zone ID from CloudFlare")
	}
	// Find the specific record
	var recordID string
	for _, r := range records {
		p.status.Infof("Examining DNS record ID '%s' with name '%s'", r.ID, r.Name)
		if r.Name == record {
			recordID = r.ID
			if r.Content == ip {
				p.status.Infof("DNS record '%s' is already set to IP '%s'", record, ip)
				return nil
			}
			break
		}
	}
	// Create the record if it's not already there
	if recordID == "" {
		p.status.Infof("No DNS record '%s' found for domain '%s', creating now", record, domain)
		resp, err := p.client.CreateDNSRecord(p.ctx, zoneID, cloudflare.DNSRecord{
			Content: ip,
			Type:    "A",
			Name:    record,
		})
		if err != nil {
			return errors.Annotatef(err, "failed to create DNS record '%s' on domain '%s'", record, domain)
		}
		recordID = resp.Result.ID
	}
	// Update the record
	err = p.client.UpdateDNSRecord(p.ctx, zoneID, recordID, cloudflare.DNSRecord{
		Content: ip,
		Type:    "A",
	})
	if err != nil {
		return errors.Annotatef(err, "failed to update DNS record '%s' to IP address '%s'", record, ip)
	}

	p.status.Infof("Successfully updated DNS record '%s' to point to '%s'", record, ip)
	return nil
}
