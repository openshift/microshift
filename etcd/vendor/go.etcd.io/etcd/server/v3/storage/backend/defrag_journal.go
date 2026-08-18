package backend

import "sync"

type defragOpType uint8

const (
	opPut defragOpType = iota
	opDelete
	opCreateBucket
	opDeleteBucket
)

// DefaultDefragJournalMaxOps is the default maximum number of operations
// the journal will buffer before applying backpressure to writers.
// The CLI flag --defrag-journal-max-ops defaults to this value.
const DefaultDefragJournalMaxOps = 50_000

type defragJournalOp struct {
	opType     defragOpType
	bucketName []byte
	key        []byte
	value      []byte
	seq        bool
}

type defragJournal struct {
	mu     sync.Mutex
	cond   *sync.Cond
	ops    []defragJournalOp
	closed bool
	maxOps int
}

// newDefragJournal creates a journal that buffers write operations
// during the defrag copy phase. maxOps sets the backpressure threshold:
// writers block when the journal reaches this size. A value of 0 means
// no limit.
func newDefragJournal(maxOps int) *defragJournal {
	j := &defragJournal{
		ops:    make([]defragJournalOp, 0, 1024),
		maxOps: maxOps,
	}
	j.cond = sync.NewCond(&j.mu)
	return j
}

// waitForSpace blocks until the journal has room for more operations
// or is closed. Must be called BEFORE acquiring the batchTx mutex to
// avoid deadlock: writers wait here without holding the mutex, so
// the defrag goroutine can still acquire the mutex to drain the journal.
func (j *defragJournal) waitForSpace() {
	j.mu.Lock()
	defer j.mu.Unlock()
	for j.maxOps > 0 && len(j.ops) >= j.maxOps && !j.closed {
		j.cond.Wait()
	}
}

func (j *defragJournal) appendPut(bucketName, key, value []byte, seq bool) {
	j.mu.Lock()
	defer j.mu.Unlock()
	if j.closed {
		panic("defragJournal: append to closed journal")
	}
	j.ops = append(j.ops, defragJournalOp{
		opType:     opPut,
		bucketName: bucketName,
		key:        key,
		value:      value,
		seq:        seq,
	})
}

func (j *defragJournal) appendDelete(bucketName, key []byte) {
	j.mu.Lock()
	defer j.mu.Unlock()
	if j.closed {
		panic("defragJournal: append to closed journal")
	}
	j.ops = append(j.ops, defragJournalOp{
		opType:     opDelete,
		bucketName: bucketName,
		key:        key,
	})
}

func (j *defragJournal) appendCreateBucket(bucketName []byte) {
	j.mu.Lock()
	defer j.mu.Unlock()
	if j.closed {
		panic("defragJournal: append to closed journal")
	}
	j.ops = append(j.ops, defragJournalOp{
		opType:     opCreateBucket,
		bucketName: bucketName,
	})
}

func (j *defragJournal) appendDeleteBucket(bucketName []byte) {
	j.mu.Lock()
	defer j.mu.Unlock()
	if j.closed {
		panic("defragJournal: append to closed journal")
	}
	j.ops = append(j.ops, defragJournalOp{
		opType:     opDeleteBucket,
		bucketName: bucketName,
	})
}

// closeAndDrain marks the journal as closed, wakes any writers
// blocked in waitForSpace, and returns all accumulated ops.
func (j *defragJournal) closeAndDrain() []defragJournalOp {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.closed = true
	ops := j.ops
	j.ops = nil
	j.cond.Broadcast()
	return ops
}
