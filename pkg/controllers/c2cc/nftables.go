package c2cc

import (
	"context"
	"fmt"
	"net"
	"strings"
	"syscall"

	"github.com/vishvananda/netlink/nl"
	"golang.org/x/sys/unix"
	"k8s.io/klog/v2"
	"sigs.k8s.io/knftables"
)

const (
	nftTable         = "ovn-kubernetes"
	nftChain         = "ovn-kube-pod-subnet-masq"
	nftCommentPrefix = "c2cc-no-masq:"
)

type nftablesManager struct {
	nft          knftables.Interface
	desiredCIDRs map[string]string // cidr -> bypass rule expression
}

func newNftablesManager(remoteCIDRs []*net.IPNet) (*nftablesManager, error) {
	nft, err := knftables.New(knftables.InetFamily, nftTable)
	if err != nil {
		return nil, fmt.Errorf("creating knftables interface: %w", err)
	}

	desired := make(map[string]string, len(remoteCIDRs))
	for _, cidr := range remoteCIDRs {
		prefix := "ip"
		if cidr.IP.To4() == nil {
			prefix = "ip6"
		}
		desired[cidr.String()] = fmt.Sprintf("%s daddr %s return", prefix, cidr)
	}

	return &nftablesManager{
		nft:          nft,
		desiredCIDRs: desired,
	}, nil
}

func nftCommentForCIDR(cidr string) string {
	return nftCommentPrefix + cidr
}

func cidrFromNFTComment(comment string) string {
	if !strings.HasPrefix(comment, nftCommentPrefix) {
		return ""
	}
	return strings.TrimPrefix(comment, nftCommentPrefix)
}

func (m *nftablesManager) reconcile(ctx context.Context) error {
	existing, err := m.nft.ListRules(ctx, nftChain)
	if err != nil {
		if knftables.IsNotFound(err) {
			klog.V(4).Infof("nftables chain %s does not exist yet, will retry", nftChain)
			return nil
		}
		return fmt.Errorf("listing nftables rules: %w", err)
	}

	actualCIDRs := make(map[string]*knftables.Rule, len(existing))
	for _, r := range existing {
		if r.Comment == nil {
			continue
		}
		if cidr := cidrFromNFTComment(*r.Comment); cidr != "" {
			actualCIDRs[cidr] = r
		}
	}

	tx := m.nft.NewTransaction()
	changed := false

	for cidr, ruleExpr := range m.desiredCIDRs {
		if _, exists := actualCIDRs[cidr]; exists {
			continue
		}
		comment := nftCommentForCIDR(cidr)
		tx.Insert(&knftables.Rule{
			Chain:   nftChain,
			Rule:    ruleExpr,
			Comment: &comment,
		})
		changed = true
		klog.V(2).Infof("nftables: inserting bypass rule for %s", cidr)
	}

	for cidr, rule := range actualCIDRs {
		if _, desired := m.desiredCIDRs[cidr]; desired {
			continue
		}
		tx.Delete(&knftables.Rule{
			Chain:  nftChain,
			Handle: rule.Handle,
		})
		changed = true
		klog.V(2).Infof("nftables: removing stale bypass rule for %s", cidr)
	}

	if !changed {
		return nil
	}

	if err := m.nft.Run(ctx, tx); err != nil {
		return fmt.Errorf("running nftables transaction: %w", err)
	}
	return nil
}

func (m *nftablesManager) cleanup(ctx context.Context) error {
	existing, err := m.nft.ListRules(ctx, nftChain)
	if err != nil {
		if knftables.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("listing nftables rules: %w", err)
	}

	tx := m.nft.NewTransaction()
	changed := false
	for _, r := range existing {
		if r.Comment == nil {
			continue
		}
		if cidrFromNFTComment(*r.Comment) != "" {
			tx.Delete(&knftables.Rule{
				Chain:  nftChain,
				Handle: r.Handle,
			})
			changed = true
		}
	}

	if !changed {
		return nil
	}

	return m.nft.Run(ctx, tx)
}

func (m *nftablesManager) subscribe(reconcileCh chan<- string) (func(), error) {
	sock, err := nl.Subscribe(unix.NETLINK_NETFILTER, unix.NFNLGRP_NFTABLES)
	if err != nil {
		return nil, fmt.Errorf("subscribe to nftables events: %w", err)
	}

	go func() {
		for {
			msgs, _, err := sock.Receive()
			if err != nil {
				klog.V(4).Infof("nftables netlink receive error: %v", err)
				return
			}
			for _, msg := range msgs {
				if msg.Header.Type == syscall.NLMSG_DONE || msg.Header.Type == syscall.NLMSG_ERROR {
					continue
				}
				// nfnetlink message type = (subsys << 8) | msg_type
				msgType := int(msg.Header.Type) & 0xFF
				if msgType == unix.NFT_MSG_NEWRULE ||
					msgType == unix.NFT_MSG_DELRULE ||
					msgType == unix.NFT_MSG_NEWCHAIN ||
					msgType == unix.NFT_MSG_DELCHAIN {
					select {
					case reconcileCh <- "nftables-change":
					default:
					}
				}
			}
		}
	}()

	klog.V(2).Infof("Subscribed to nftables netlink events (NFNLGRP_NFTABLES)")
	return sock.Close, nil
}
