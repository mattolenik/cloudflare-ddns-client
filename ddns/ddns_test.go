package ddns

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/mattolenik/cloudflare-ddns-client/test"
)

func TestDDNSDaemon(t *testing.T) {
	assert, _, ctrl, cleanup := test.NewTools(t)
	defer cleanup()

	domain := "abc.com"
	record := "xyz.abc.com"
	expectedIP := "1.1.1.1"

	ddnsProvider := NewMockDDNSProvider(ctrl)
	ipProvider := NewMockIPProvider(ctrl)
	configProvider := NewMockConfigProvider(ctrl)
	ddnsDaemon := NewDDNSDaemon(ddnsProvider, ipProvider, configProvider)

	configProvider.EXPECT().Get().Return(domain, record, nil)
	ipProvider.EXPECT().Get().Return(expectedIP, nil)
	ddnsProvider.EXPECT().Update(gomock.Eq(domain), gomock.Eq(record), gomock.Eq(expectedIP)).Return(nil).Times(1)
	ddnsProvider.EXPECT().Get(domain, record).Return(expectedIP, nil).Times(1)
	assert.NoError(ddnsDaemon.Update())

	actualIP, err := ddnsProvider.Get(domain, record)
	assert.NoError(err)
	assert.Equal(expectedIP, actualIP)
}
