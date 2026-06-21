/*
Copyright 2015 The Kubernetes Authors.

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

package ipallocator

import (
	"errors"
	"fmt"
	"math/big"
	"net"

	netutils "k8s.io/utils/net"
)

var (
	ErrFull      = errors.New("range is full")
	ErrAllocated = errors.New("provided IP is already allocated")
)

type ErrNotInRange struct {
	IP         net.IP
	ValidRange string
}

func (e *ErrNotInRange) Error() string {
	return fmt.Sprintf("the provided IP (%v) is not in the valid range. The range of valid IPs is %s", e.IP, e.ValidRange)
}

// Range is a contiguous block of IPs that can be allocated atomically.
//
// The internal structure of the range is:
//
//	For CIDR 10.0.0.0/24
//	254 addresses usable out of 256 total (minus base and broadcast IPs)
//	  The number of usable addresses is r.max
//
//	CIDR base IP          CIDR broadcast IP
//	10.0.0.0                     10.0.0.255
//	|                                     |
//	0 1 2 3 4 5 ...         ... 253 254 255
//	  |                              |
//	r.base                     r.base + r.max
//	  |                              |
//	offset #0 of r.allocated   last offset of r.allocated
type Range struct {
	net  *net.IPNet
	base *big.Int
	max  int

	alloc allocatorInterface
}

// NewInMemory creates an in-memory IP allocator over a net.IPNet.
func NewInMemory(cidr *net.IPNet) (*Range, error) {
	max := netutils.RangeSize(cidr)
	base := netutils.BigForIP(cidr.IP)
	rangeSpec := cidr.String()

	if netutils.IsIPv6CIDR(cidr) {
		if max > 65536 {
			max = 65536
		}
	} else {
		// Don't use the IPv4 network's broadcast address.
		max--
	}

	// Don't use the network's ".0" address.
	base.Add(base, big.NewInt(1))
	max--

	if max < 0 {
		max = 0
	}

	r := Range{
		net:  cidr,
		base: base,
		max:  maximum(0, int(max)),
	}

	offset := calculateRangeOffset(cidr)
	r.alloc = newAllocationMapWithOffset(r.max, rangeSpec, offset)
	return &r, nil
}

func maximum(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Free returns the count of IP addresses left in the range.
func (r *Range) Free() int {
	return r.alloc.Free()
}

// Allocate attempts to reserve the provided IP. ErrNotInRange or
// ErrAllocated will be returned if the IP is not valid for this range
// or has already been reserved. ErrFull will be returned if there
// are no addresses left.
func (r *Range) Allocate(ip net.IP) error {
	ok, offset := r.contains(ip)
	if !ok {
		return &ErrNotInRange{ip, r.net.String()}
	}

	allocated, err := r.alloc.Allocate(offset)
	if err != nil {
		return err
	}
	if !allocated {
		return ErrAllocated
	}
	return nil
}

// AllocateNext reserves one of the IPs from the pool. ErrFull may
// be returned if there are no addresses left.
func (r *Range) AllocateNext() (net.IP, error) {
	offset, ok, err := r.alloc.AllocateNext()
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrFull
	}
	return netutils.AddIPOffset(r.base, offset), nil
}

// Release releases the IP back to the pool. Releasing an
// unallocated IP or an IP out of the range is a no-op and
// returns no error.
func (r *Range) Release(ip net.IP) error {
	ok, offset := r.contains(ip)
	if !ok {
		return nil
	}
	return r.alloc.Release(offset)
}

// Has returns true if the provided IP is already allocated and a call
// to Allocate(ip) would fail with ErrAllocated.
func (r *Range) Has(ip net.IP) bool {
	ok, offset := r.contains(ip)
	if !ok {
		return false
	}
	return r.alloc.Has(offset)
}

// contains returns true and the offset if the ip is in the range, and false
// and 0 otherwise. The first and last addresses of the CIDR are omitted.
func (r *Range) contains(ip net.IP) (bool, int) {
	if !r.net.Contains(ip) {
		return false, 0
	}

	offset := calculateIPOffset(r.base, ip)
	if offset < 0 || offset >= r.max {
		return false, 0
	}
	return true, offset
}

// calculateIPOffset calculates the integer offset of ip from base such that
// base + offset = ip. It requires ip >= base.
func calculateIPOffset(base *big.Int, ip net.IP) int {
	return int(big.NewInt(0).Sub(netutils.BigForIP(ip), base).Int64())
}

// calculateRangeOffset estimates the offset used on the range for static allocation based on
// the following formula `min(max($min, cidrSize/$step), $max)`, described as ~never less than
// $min or more than $max, with a graduated step function between them~. The function returns 0
// if any of the parameters is invalid.
func calculateRangeOffset(cidr *net.IPNet) int {
	const (
		min  = 16
		max  = 256
		step = 16
	)

	cidrSize := netutils.RangeSize(cidr)
	if cidrSize <= min {
		return 0
	}

	offset := cidrSize / step
	if offset < min {
		return min
	}
	if offset > max {
		return max
	}
	return int(offset)
}
