package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"path"
	"testing"
	"time"

	"github.com/cloudflare/cloudflare-go"
	"github.com/juju/errors"
	"github.com/mattolenik/cloudflare-ddns-client/ip"
	"github.com/stretchr/testify/suite"
)

type EndToEndSuite struct {
	suite.Suite
	Token      string
	ZoneID     string
	Domain     string
	IP         string
	CF         *cloudflare.API
	TestBinary string
	ctx        context.Context
	rand       *rand.Rand
}

func TestEndToEndSuite(t *testing.T) {
	suite.Run(t, new(EndToEndSuite))
}

func (s *EndToEndSuite) SetupSuite() {
	s.rand = rand.New(rand.NewSource(time.Now().UnixNano()))
	s.ctx = context.Background()
	require := s.Require()
	var err error
	s.TestBinary = os.Getenv("TEST_BINARY")
	require.NotEmpty(s.TestBinary, "End-to-end tests require compiled binary of cloudflare-ddns at path specified by TEST_BINARY")

	s.Token = os.Getenv("CLOUDFLARE_TOKEN")
	require.NotEmpty(s.Token, "End-to-end tests require an API token specified by the CLOUDFLARE_TOKEN env var")

	s.Domain = os.Getenv("TEST_DOMAIN")
	require.NotEmpty(s.Domain, "End-to-end tests require a domain specified by the TEST_DOMAIN env var")

	s.IP, err = ip.GetPublicIP()
	require.NoError(err, "unable to get public IP for tests")

	s.CF, err = cloudflare.NewWithAPIToken(s.Token)
	if err != nil {
		require.NoError(err, "Unable to connect to CloudFlare, token may be invalid")
	}

	// Verify token before tests start
	_, err = s.CF.VerifyAPIToken(s.ctx)
	require.NoError(err, "CloudFlare token is not valid")

	s.ZoneID, err = s.CF.ZoneIDByName(s.Domain)
	require.NoErrorf(err, "Failed to get zone ID, error: %+v", err)
}

func (s *EndToEndSuite) TestWithConfigFile() {
	assert := s.Assert()
	require := s.Require()

	record := s.randomDNSRecord()
	defer s.deleteRecord(record)

	tmp := s.T().TempDir()
	configFile := path.Join(tmp, "config.toml")
	config := fmt.Sprintf(`
	domain="%s"
	record="%s"
	token="%s"`, s.Domain, record, s.Token)
	err := ioutil.WriteFile(configFile, []byte(config), 0644)
	assert.NoErrorf(err, "Could not write test config file")

	out, err := s.runProgram(nil, "--config", configFile)
	log.Println(out)
	require.NoError(err, "Expected no error when running cloudflare-ddns")
	assert.True(s.ipMatches(record), "Expected IP to have been updated")
}

func (s *EndToEndSuite) TestWithArguments() {
	assert := s.Assert()
	require := s.Require()

	record := s.randomDNSRecord()
	defer s.deleteRecord(record)

	out, err := s.runProgram(nil, "--domain", s.Domain, "--record", record, "--token", s.Token)
	log.Println(out)
	require.NoError(err, "Expected no error when running cloudflare-ddns")
	assert.True(s.ipMatches(record), "Expected IP to have been updated")
}

func (s *EndToEndSuite) TestExistingRecord() {
	assert := s.Assert()
	require := s.Require()

	record := s.createRandomDNSRecord("10.0.0.0")
	defer s.deleteRecord(record)

	out, err := s.runProgram(nil, "--domain", s.Domain, "--record", record, "--token", s.Token)
	log.Println(out)
	require.NoError(err, "Expected no error when running cloudflare-ddns")
	assert.True(s.ipMatches(record), "Expected IP to have been updated")
}

func (s *EndToEndSuite) TestWithEnvVars() {
	assert := s.Assert()
	require := s.Require()

	record := s.randomDNSRecord()
	defer s.deleteRecord(record)

	out, err := s.runProgram([]string{"DOMAIN=" + s.Domain, "RECORD=" + record, "TOKEN=" + s.Token})
	log.Println(out)
	require.NoError(err, "Expected no error when running cloudflare-ddns")
	assert.True(s.ipMatches(record), "Expected IP to have been updated")
}

func (s *EndToEndSuite) deleteRecord(record string) {
	r, err := s.getDNSRecordByName(record)
	s.Assert().NoErrorf(err, "Could not find DNS record of name '%s' in zone ID '%s', cannot clean up", record, s.ZoneID)
	err = s.CF.DeleteDNSRecord(s.ctx, cloudflare.ZoneIdentifier(s.ZoneID), r.ID)
	s.Assert().NoErrorf(err, "Failed to remove DNS record of name '%s' in zone ID '%s'", record, s.ZoneID)
}

func (s *EndToEndSuite) ipMatches(record string) bool {
	r, err := s.getDNSRecordByName(record)
	s.Require().NotNilf(r, "Expected record for '%s' to be not nil", record)
	s.Require().NoError(err, "Failed to get record ID of '%s'", record)
	return r.Content == s.IP
}

func (s *EndToEndSuite) getDNSRecordByName(record string) (*cloudflare.DNSRecord, error) {
	records, _, err := s.CF.ListDNSRecords(s.ctx, cloudflare.ZoneIdentifier(s.ZoneID), cloudflare.ListDNSRecordsParams{Type: "A"})
	if err != nil {
		return nil, errors.Trace(err)
	}
	for _, r := range records {
		if r.Name == record {
			return &r, nil
		}
	}
	return nil, errors.NotFoundf("no record '%s' found in zone ID '%s'", record, s.ZoneID)
}

func (s *EndToEndSuite) createRandomDNSRecord(ip string) string {
	record := s.randomDNSRecord()
	_, err := s.CF.CreateDNSRecord(s.ctx, cloudflare.ZoneIdentifier(s.ZoneID), cloudflare.CreateDNSRecordParams{
		Content: ip,
		Type:    "A",
		Name:    record,
	})
	s.Require().NoErrorf(err, "failed to create DNS record '%s' on domain '%s'", record, s.Domain)
	return record
}

func (s *EndToEndSuite) randomDNSRecord() string {
	return fmt.Sprintf("ddns-e2e-test-%d.%s", s.rand.Intn(999999)+100000, s.Domain)
}

func (s *EndToEndSuite) runProgram(envVars []string, args ...string) (string, error) {
	cmd := exec.Command(s.TestBinary, args...)
	cmd.Env = envVars
	out, err := cmd.CombinedOutput()
	return string(out), err
}
