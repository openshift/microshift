package c2cc

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNBDBModel(t *testing.T) {
	dbModel, err := nbdbModel()
	require.NoError(t, err)
	assert.Equal(t, ovnNBDatabase, dbModel.Name())
}

func TestLogicalRouterStaticRoute_FieldTags(t *testing.T) {
	route := LogicalRouterStaticRoute{
		UUID:        "test-uuid",
		IPPrefix:    "10.45.0.0/16",
		Nexthop:     "192.168.1.1",
		ExternalIDs: map[string]string{"k8s.ovn.org/owner-controller": "microshift-c2cc"},
	}

	assert.Equal(t, "test-uuid", route.UUID)
	assert.Equal(t, "10.45.0.0/16", route.IPPrefix)
	assert.Equal(t, "192.168.1.1", route.Nexthop)
	assert.Equal(t, "microshift-c2cc", route.ExternalIDs["k8s.ovn.org/owner-controller"])
	assert.Nil(t, route.Policy)
}

func TestLogicalRouter_FieldTags(t *testing.T) {
	router := LogicalRouter{
		UUID:         "router-uuid",
		Name:         "GR_test-node",
		StaticRoutes: []string{"route-uuid-1", "route-uuid-2"},
	}

	assert.Equal(t, "router-uuid", router.UUID)
	assert.Equal(t, "GR_test-node", router.Name)
	assert.Len(t, router.StaticRoutes, 2)
}
