package c2cc

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/knftables"
)

func setupFakeNFT(t *testing.T) *knftables.Fake {
	t.Helper()
	fake := knftables.NewFake(knftables.InetFamily, nftTable)

	tx := fake.NewTransaction()
	tx.Add(&knftables.Table{})
	tx.Add(&knftables.Chain{Name: nftChain})
	require.NoError(t, fake.Run(context.Background(), tx))

	return fake
}

func parseCIDR(t *testing.T, s string) *net.IPNet {
	t.Helper()
	_, cidr, err := net.ParseCIDR(s)
	require.NoError(t, err)
	return cidr
}

func TestNftCommentForCIDR(t *testing.T) {
	assert.Equal(t, "c2cc-no-masq:10.45.0.0/16", nftCommentForCIDR("10.45.0.0/16"))
	assert.Equal(t, "c2cc-no-masq:fd01::/48", nftCommentForCIDR("fd01::/48"))
}

func TestCIDRFromNFTComment(t *testing.T) {
	tests := []struct {
		comment  string
		expected string
	}{
		{"c2cc-no-masq:10.45.0.0/16", "10.45.0.0/16"},
		{"c2cc-no-masq:fd01::/48", "fd01::/48"},
		{"some-other-comment", ""},
		{"", ""},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.expected, cidrFromNFTComment(tt.comment))
	}
}

func TestNftablesManager_ReconcileAddsMissingRules(t *testing.T) {
	fake := setupFakeNFT(t)
	ctx := context.Background()

	cidrs := []*net.IPNet{
		parseCIDR(t, "10.45.0.0/16"),
		parseCIDR(t, "10.46.0.0/16"),
	}

	mgr := &nftablesManager{
		nft:          fake,
		desiredCIDRs: make(map[string]string),
	}
	for _, cidr := range cidrs {
		mgr.desiredCIDRs[cidr.String()] = "ip daddr " + cidr.String() + " return"
	}

	err := mgr.reconcile(ctx)
	require.NoError(t, err)

	rules, err := fake.ListRules(ctx, nftChain)
	require.NoError(t, err)
	assert.Len(t, rules, 2)

	foundCIDRs := make(map[string]bool)
	for _, r := range rules {
		require.NotNil(t, r.Comment)
		cidr := cidrFromNFTComment(*r.Comment)
		foundCIDRs[cidr] = true
	}
	assert.True(t, foundCIDRs["10.45.0.0/16"])
	assert.True(t, foundCIDRs["10.46.0.0/16"])
}

func TestNftablesManager_ReconcileIsIdempotent(t *testing.T) {
	fake := setupFakeNFT(t)
	ctx := context.Background()

	mgr := &nftablesManager{
		nft:          fake,
		desiredCIDRs: map[string]string{"10.45.0.0/16": "ip daddr 10.45.0.0/16 return"},
	}

	require.NoError(t, mgr.reconcile(ctx))
	require.NoError(t, mgr.reconcile(ctx))

	rules, err := fake.ListRules(ctx, nftChain)
	require.NoError(t, err)
	assert.Len(t, rules, 1)
}

func TestNftablesManager_ReconcileRemovesStaleRules(t *testing.T) {
	fake := setupFakeNFT(t)
	ctx := context.Background()

	comment := nftCommentForCIDR("10.99.0.0/16")
	tx := fake.NewTransaction()
	tx.Add(&knftables.Rule{
		Chain:   nftChain,
		Rule:    "ip daddr 10.99.0.0/16 return",
		Comment: &comment,
	})
	require.NoError(t, fake.Run(ctx, tx))

	mgr := &nftablesManager{
		nft:          fake,
		desiredCIDRs: map[string]string{"10.45.0.0/16": "ip daddr 10.45.0.0/16 return"},
	}

	require.NoError(t, mgr.reconcile(ctx))

	rules, err := fake.ListRules(ctx, nftChain)
	require.NoError(t, err)
	assert.Len(t, rules, 1)
	require.NotNil(t, rules[0].Comment)
	assert.Equal(t, "10.45.0.0/16", cidrFromNFTComment(*rules[0].Comment))
}

func TestNftablesManager_ReconcileChainNotFound(t *testing.T) {
	fake := knftables.NewFake(knftables.InetFamily, nftTable)
	ctx := context.Background()

	tx := fake.NewTransaction()
	tx.Add(&knftables.Table{})
	require.NoError(t, fake.Run(ctx, tx))

	mgr := &nftablesManager{
		nft:          fake,
		desiredCIDRs: map[string]string{"10.45.0.0/16": "ip daddr 10.45.0.0/16 return"},
	}

	err := mgr.reconcile(ctx)
	assert.NoError(t, err)
}

func TestNftablesManager_CleanupRemovesAllC2CCRules(t *testing.T) {
	fake := setupFakeNFT(t)
	ctx := context.Background()

	tx := fake.NewTransaction()
	c1 := nftCommentForCIDR("10.45.0.0/16")
	c2 := nftCommentForCIDR("10.46.0.0/16")
	nonC2CC := "some-other-rule"
	tx.Add(&knftables.Rule{Chain: nftChain, Rule: "ip daddr 10.45.0.0/16 return", Comment: &c1})
	tx.Add(&knftables.Rule{Chain: nftChain, Rule: "ip daddr 10.46.0.0/16 return", Comment: &c2})
	tx.Add(&knftables.Rule{Chain: nftChain, Rule: "ip daddr 10.0.0.0/8 masquerade", Comment: &nonC2CC})
	require.NoError(t, fake.Run(ctx, tx))

	mgr := &nftablesManager{
		nft:          fake,
		desiredCIDRs: map[string]string{},
	}

	require.NoError(t, mgr.cleanup(ctx))

	rules, err := fake.ListRules(ctx, nftChain)
	require.NoError(t, err)
	assert.Len(t, rules, 1, "only non-c2cc rule should remain")
	assert.Equal(t, "some-other-rule", *rules[0].Comment)
}

func TestNftablesManager_IPv6RuleExpression(t *testing.T) {
	cidrs := []*net.IPNet{parseCIDR(t, "fd01::/48")}

	fake := setupFakeNFT(t)
	mgr := &nftablesManager{
		nft:          fake,
		desiredCIDRs: make(map[string]string),
	}
	for _, cidr := range cidrs {
		prefix := "ip"
		if cidr.IP.To4() == nil {
			prefix = "ip6"
		}
		mgr.desiredCIDRs[cidr.String()] = prefix + " daddr " + cidr.String() + " return"
	}

	require.NoError(t, mgr.reconcile(context.Background()))

	rules, err := fake.ListRules(context.Background(), nftChain)
	require.NoError(t, err)
	require.Len(t, rules, 1)
	assert.Contains(t, rules[0].Rule, "ip6 daddr fd01::/48 return")
}
