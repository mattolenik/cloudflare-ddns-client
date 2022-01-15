package ddns

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

// mockDaemon is a mock DDNSDaemon
type mockDaemon struct {
	DDNSDaemon
	domain, record, ip string
}

func (p *mockDaemon) Update(provider DDNSProvider) error {
	return provider.Update(p.domain, p.record, p.ip)
}

func (p *mockDaemon) Start(provider DDNSProvider, updatePeriod, failureRetryDelay time.Duration) error {
	return nil
}

func (p *mockDaemon) Stop() error {
	return nil
}

func TestDDNSDaemon(t *testing.T) {
	assert := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	d := mockDaemon{}
	p := NewMockDDNSProvider(ctrl)
	d.domain = "abc.com"
	d.record = "xyz.abc.com"
	d.ip = "1.1.1.1"
	p.EXPECT().Update(d.domain, d.record, d.ip).Return(nil).Times(1)
	p.EXPECT().Get(d.domain, d.record).Return(d.ip, nil).Times(1)
	assert.NoError(d.Update(p))
	ip, err := p.Get(d.domain, d.record)
	assert.NoError(err)
	assert.Equal(d.ip, ip)
}
