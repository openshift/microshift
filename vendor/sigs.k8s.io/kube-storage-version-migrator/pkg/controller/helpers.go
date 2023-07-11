/*
Copyright 2018 The Kubernetes Authors.

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

package controller

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	migrationv1alpha1 "sigs.k8s.io/kube-storage-version-migrator/pkg/apis/migration/v1alpha1"
)

func HasCondition(m *migrationv1alpha1.StorageVersionMigration, conditionType migrationv1alpha1.MigrationConditionType) bool {
	return indexOfCondition(m, conditionType) != -1
}

func indexOfCondition(m *migrationv1alpha1.StorageVersionMigration, conditionType migrationv1alpha1.MigrationConditionType) int {
	for i, c := range m.Status.Conditions {
		if c.Type == conditionType && c.Status == corev1.ConditionTrue {
			return i
		}
	}
	return -1
}

func resource(m *migrationv1alpha1.StorageVersionMigration) schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    m.Spec.Resource.Group,
		Version:  m.Spec.Resource.Version,
		Resource: m.Spec.Resource.Resource,
	}
}
