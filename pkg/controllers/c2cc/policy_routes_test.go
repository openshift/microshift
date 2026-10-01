package c2cc

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vishvananda/netlink"
)

func TestIPFamilyOf(t *testing.T) {
	tests := []struct {
		cidr     string
		expected int
	}{
		{"10.45.0.0/16", netlink.FAMILY_V4},
		{"192.168.1.0/24", netlink.FAMILY_V4},
		{"fd01::/48", netlink.FAMILY_V6},
		{"::1/128", netlink.FAMILY_V6},
	}

	for _, tt := range tests {
		t.Run(tt.cidr, func(t *testing.T) {
			cidr := parseCIDR(t, tt.cidr)
			assert.Equal(t, tt.expected, ipFamilyOf(cidr))
		})
	}
}
