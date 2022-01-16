package ddns

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestDDNSDaemon(t *testing.T) {
	assert := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	d := NewMockDaemon(ctrl)
	p := NewMockDDNSProvider(ctrl)
	domain := "abc.com"
	record := "xyz.abc.com"
	ip := "1.1.1.1"
	p.EXPECT().Update(gomock.Eq(domain), gomock.Eq(record), gomock.Eq(ip)).Return(nil).Times(1)
	p.EXPECT().Get(domain, record).Return(ip, nil).Times(1)
	assert.NoError(d.Update(p))
	ip, err := p.Get(domain)
	assert.NoError(err)
	assert.Equal(ip, ip)
}
