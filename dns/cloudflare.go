package dns

import (
	"github.com/cloudflare/cloudflare-go"
	"github.com/juju/errors"
	"github.com/rs/zerolog/log"
)

// UpdateCloudFlare updates the CloudFlare DNS record
func UpdateCloudFlare(token, domain, record, ip string) error {
	api, err := cloudflare.NewWithAPIToken(token)
	if err != nil {
		return errors.Annotate(err, "unable to connect to CloudFlare, token may be invalid")
	}
	// Get the zone ID for the domain
	zoneID, err := api.ZoneIDByName(domain)
	if err != nil {
		return errors.Annotatef(err, "unable to retrieve zone ID for domain '%s' from CloudFlare", domain)
	}
	// Get the record ID
	records, err := api.DNSRecords(zoneID, cloudflare.DNSRecord{Type: "A"})
	if err != nil {
		return errors.Annotate(err, "unable to retrieve zone ID from CloudFlare")
	}
	// Find the specific record
	var recordID string
	for _, r := range records {
		log.Debug().Msgf("Examining DNS record ID '%s' with name '%s'", r.ID, r.Name)
		if r.Name == record {
			recordID = r.ID
			if r.Content == ip {
				log.Info().Msgf("DNS record '%s' is already set to IP '%s'", record, ip)
				return nil
			}
			break
		}
	}
	// Create the record if it's not already there
	if recordID == "" {
		log.Info().Msgf("No DNS '%s' found for domain '%s', creating now", record, domain)
		resp, err := api.CreateDNSRecord(zoneID, cloudflare.DNSRecord{
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
	err = api.UpdateDNSRecord(zoneID, recordID, cloudflare.DNSRecord{
		Content: ip,
		Type:    "A",
	})
	if err != nil {
		return errors.Annotatef(err, "failed to update DNS record '%s' to IP address '%s'", record, ip)
	}

	log.Info().Msgf("Successfully updated DNS record '%s' to point to '%s'", record, ip)
	return nil
}
