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
	"fmt"
	"math/big"
	"math/rand"
	"sync"
	"time"
)

// allocatorInterface manages the allocation of items out of a range.
type allocatorInterface interface {
	Allocate(int) (bool, error)
	AllocateNext() (int, bool, error)
	Release(int) error
	Has(int) bool
	Free() int
}

// allocationBitmap is a contiguous block of resources that can be allocated atomically.
//
// Each resource has an offset. The internal structure is a bitmap, with a bit for each offset.
//
// If a resource is taken, the bit at that offset is set to one.
// r.count is always equal to the number of set bits and can be recalculated at any time
// by counting the set bits in r.allocated.
type allocationBitmap struct {
	strategy  bitAllocator
	max       int
	rangeSpec string

	lock      sync.Mutex
	count     int
	allocated *big.Int
}

var _ allocatorInterface = &allocationBitmap{}

// bitAllocator represents a search strategy in the allocation map for a valid item.
type bitAllocator interface {
	AllocateBit(allocated *big.Int, max, count int) (int, bool)
}

// newAllocationMapWithOffset creates an allocation bitmap using a random scan strategy that
// allows to pass an offset that divides the allocation bitmap in two blocks.
// The first block of values will not be used for random value assigned by the AllocateNext()
// method until the second block of values has been exhausted.
func newAllocationMapWithOffset(max int, rangeSpec string, offset int) *allocationBitmap {
	return &allocationBitmap{
		strategy: randomScanStrategyWithOffset{
			rand:   rand.New(rand.NewSource(time.Now().UnixNano())),
			offset: offset,
		},
		allocated: big.NewInt(0),
		count:     0,
		max:       max,
		rangeSpec: rangeSpec,
	}
}

// Allocate attempts to reserve the provided item.
// Returns true if it was allocated, false if it was already in use.
func (r *allocationBitmap) Allocate(offset int) (bool, error) {
	r.lock.Lock()
	defer r.lock.Unlock()

	if offset < 0 || offset >= r.max {
		return false, fmt.Errorf("offset %d out of range [0,%d]", offset, r.max)
	}
	if r.allocated.Bit(offset) == 1 {
		return false, nil
	}
	r.allocated = r.allocated.SetBit(r.allocated, offset, 1)
	r.count++
	return true, nil
}

// AllocateNext reserves one of the items from the pool.
// (0, false, nil) may be returned if there are no items left.
func (r *allocationBitmap) AllocateNext() (int, bool, error) {
	r.lock.Lock()
	defer r.lock.Unlock()

	next, ok := r.strategy.AllocateBit(r.allocated, r.max, r.count)
	if !ok {
		return 0, false, nil
	}
	r.count++
	r.allocated = r.allocated.SetBit(r.allocated, next, 1)
	return next, true, nil
}

// Release releases the item back to the pool. Releasing an
// unallocated item or an item out of the range is a no-op and
// returns no error.
func (r *allocationBitmap) Release(offset int) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	if r.allocated.Bit(offset) == 0 {
		return nil
	}

	r.allocated = r.allocated.SetBit(r.allocated, offset, 0)
	r.count--
	return nil
}

// Has returns true if the provided item is already allocated and a call
// to Allocate(offset) would fail.
func (r *allocationBitmap) Has(offset int) bool {
	r.lock.Lock()
	defer r.lock.Unlock()

	return r.allocated.Bit(offset) == 1
}

// Free returns the count of items left in the range.
func (r *allocationBitmap) Free() int {
	r.lock.Lock()
	defer r.lock.Unlock()
	return r.max - r.count
}

// randomScanStrategyWithOffset chooses a random address from the provided big.Int and then scans
// forward looking for the next available address. The big.Int range is subdivided so it will try
// to allocate first from the reserved upper range of addresses (it will wrap the upper subrange if necessary).
// If there is no free address it will try to allocate one from the lower range too.
type randomScanStrategyWithOffset struct {
	rand   *rand.Rand
	offset int
}

func (rss randomScanStrategyWithOffset) AllocateBit(allocated *big.Int, max, count int) (int, bool) {
	if count >= max {
		return 0, false
	}
	subrangeMax := max - rss.offset
	start := rss.rand.Intn(subrangeMax)
	for i := 0; i < subrangeMax; i++ {
		at := rss.offset + ((start + i) % subrangeMax)
		if allocated.Bit(at) == 0 {
			return at, true
		}
	}

	// Guard against rand.Intn(0) panic when offset is 0.
	if rss.offset > 0 {
		start = rss.rand.Intn(rss.offset)
		for i := 0; i < rss.offset; i++ {
			at := (start + i) % rss.offset
			if allocated.Bit(at) == 0 {
				return at, true
			}
		}
	}
	return 0, false
}

var _ bitAllocator = randomScanStrategyWithOffset{}
