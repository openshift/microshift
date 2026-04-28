package c2cc

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/ovn-kubernetes/libovsdb/client"
	"github.com/ovn-kubernetes/libovsdb/model"
	"k8s.io/klog/v2"
)

const (
	ovnNBSocketPath = "/var/run/ovn/ovnnb_db.sock"
	ovnNBEndpoint   = "unix:" + ovnNBSocketPath
	ovnNBDatabase   = "OVN_Northbound"

	socketPollInterval = 5 * time.Second
	connectTimeout     = 30 * time.Second
)

// LogicalRouter is a minimal OVN NB model for the Logical_Router table.
type LogicalRouter struct {
	UUID         string   `ovsdb:"_uuid"`
	Name         string   `ovsdb:"name"`
	StaticRoutes []string `ovsdb:"static_routes"`
}

// LogicalRouterStaticRoute is a minimal OVN NB model for the Logical_Router_Static_Route table.
type LogicalRouterStaticRoute struct {
	UUID        string            `ovsdb:"_uuid"`
	IPPrefix    string            `ovsdb:"ip_prefix"`
	Nexthop     string            `ovsdb:"nexthop"`
	ExternalIDs map[string]string `ovsdb:"external_ids"`
	Policy      *string           `ovsdb:"policy"`
}

func nbdbModel() (model.ClientDBModel, error) {
	dbModel, err := model.NewClientDBModel(ovnNBDatabase, map[string]model.Model{
		"Logical_Router":              &LogicalRouter{},
		"Logical_Router_Static_Route": &LogicalRouterStaticRoute{},
	})
	if err != nil {
		return dbModel, err
	}
	dbModel.SetIndexes(map[string][]model.ClientIndex{
		"Logical_Router": {{Columns: []model.ColumnKey{{Column: "name"}}}},
	})
	return dbModel, nil
}

func waitForOVNSocket(ctx context.Context) error {
	for {
		if _, err := os.Stat(ovnNBSocketPath); err == nil {
			return nil
		}
		klog.V(2).Infof("Waiting for OVN NB socket at %s", ovnNBSocketPath)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(socketPollInterval):
		}
	}
}

func connectOVNNB(ctx context.Context) (client.Client, error) {
	if err := waitForOVNSocket(ctx); err != nil {
		return nil, fmt.Errorf("waiting for OVN NB socket: %w", err)
	}

	dbModel, err := nbdbModel()
	if err != nil {
		return nil, fmt.Errorf("building OVN NB database model: %w", err)
	}

	nbClient, err := client.NewOVSDBClient(
		dbModel,
		client.WithEndpoint(ovnNBEndpoint),
		client.WithReconnect(connectTimeout, backoff.NewExponentialBackOff()),
	)
	if err != nil {
		return nil, fmt.Errorf("creating OVN NB client: %w", err)
	}

	if err := nbClient.Connect(ctx); err != nil {
		nbClient.Close()
		return nil, fmt.Errorf("connecting to OVN NB: %w", err)
	}

	_, err = nbClient.MonitorAll(ctx)
	if err != nil {
		nbClient.Close()
		return nil, fmt.Errorf("setting up OVN NB monitor: %w", err)
	}

	klog.Infof("Connected to OVN NB database at %s", ovnNBEndpoint)
	return nbClient, nil
}
