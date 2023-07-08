/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace = "storage_migrator"
	subsystem = "core_migrator"
)

var (
	// Metrics provides access to all core migrator metrics.
	Metrics = newCoreMigratorMetrics()
)

// CoreMigratorMetrics instruments core migrator with prometheus metrics.
type CoreMigratorMetrics struct {
	objectsMigrated  *prometheus.CounterVec
	objectsRemaining *prometheus.GaugeVec
	migration        *prometheus.CounterVec
}

// newCoreMigratorMetrics create a new CoreMigratorMetrics, configured with default metric names.
func newCoreMigratorMetrics() *CoreMigratorMetrics {
	objectsMigrated := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "migrated_objects",
			Help:      "The number of objects that have been migrated, labeled with the full resource name.",
		}, []string{"resource"})
	prometheus.MustRegister(objectsMigrated)

	objectsRemaining := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "remaining_objects",
			Help:      "The number of objects that still require migration, labeled with the full resource name",
		}, []string{"resource"})
	prometheus.MustRegister(objectsRemaining)

	migration := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "migrations",
			Help:      "The number of completed migration, labeled with the full resource name, and the status of the migration (failed or succeeded)",
		}, []string{"resource", "status"})
	prometheus.MustRegister(migration)

	return &CoreMigratorMetrics{
		objectsMigrated:  objectsMigrated,
		objectsRemaining: objectsRemaining,
		migration:        migration,
	}
}

func (m *CoreMigratorMetrics) Reset() {
	m.objectsMigrated.Reset()
	m.objectsRemaining.Reset()
	m.migration.Reset()
}

// ObserveObjectsMigrated adds the number of migrated objects for a resource type..
func (m *CoreMigratorMetrics) ObserveObjectsMigrated(added int, resource string) {
	m.objectsMigrated.WithLabelValues(resource).Add(float64(added))
}

// ObserveObjectsRemaining records the number of objects pending-migration for a particular resource type.
func (m *CoreMigratorMetrics) ObserveObjectsRemaining(count int, resource string) {
	m.objectsRemaining.WithLabelValues(resource).Set(float64(count))
}

// ObserveSucceededMigration increments the number of successful migrations for a resource type..
func (m *CoreMigratorMetrics) ObserveSucceededMigration(resource string) {
	m.migration.WithLabelValues(resource, "Succeeded").Add(float64(1))
}

// ObserveFailedMigration increments the number of failed migrations for a resource type..
func (m *CoreMigratorMetrics) ObserveFailedMigration(resource string) {
	m.migration.WithLabelValues(resource, "Failed").Add(float64(1))
}
