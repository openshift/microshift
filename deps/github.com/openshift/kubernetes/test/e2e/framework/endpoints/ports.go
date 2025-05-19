/*
Copyright 2019 The Kubernetes Authors.

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

package endpoints

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

// PortsByPodUID is a map that maps pod UID to container ports.
type PortsByPodUID map[types.UID][]int

// FullPortsByPodUID is a map that maps pod UID to container ports.
type FullPortsByPodUID map[types.UID][]v1.ContainerPort

// GetContainerPortsByPodUID returns a PortsByPodUID map on the given endpoints.
func GetContainerPortsByPodUID(ep *v1.Endpoints) PortsByPodUID {
	m := PortsByPodUID{}
	for _, ss := range ep.Subsets {
		for _, port := range ss.Ports {
			for _, addr := range ss.Addresses {
				containerPort := port.Port
				if _, ok := m[addr.TargetRef.UID]; !ok {
					m[addr.TargetRef.UID] = make([]int, 0)
				}
				m[addr.TargetRef.UID] = append(m[addr.TargetRef.UID], int(containerPort))
			}
		}
	}
	return m
}

// GetFullContainerPortsByPodUID returns a FullPortsByPodUID map on the given endpoints with all the port data.
func GetFullContainerPortsByPodUID(ep *v1.Endpoints) FullPortsByPodUID {
	m := FullPortsByPodUID{}
	for _, ss := range ep.Subsets {
		for _, port := range ss.Ports {
			containerPort := v1.ContainerPort{
				Name:          port.Name,
				ContainerPort: port.Port,
				Protocol:      port.Protocol,
			}
			for _, addr := range ss.Addresses {
				if _, ok := m[addr.TargetRef.UID]; !ok {
					m[addr.TargetRef.UID] = make([]v1.ContainerPort, 0)
				}
				m[addr.TargetRef.UID] = append(m[addr.TargetRef.UID], containerPort)
			}
		}
	}
	return m
}
