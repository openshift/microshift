package revbump

import (
	"go.etcd.io/etcd/server/v3/mvcc"
	"go.etcd.io/etcd/server/v3/mvcc/backend"
	"go.etcd.io/etcd/server/v3/mvcc/buckets"
	"go.uber.org/zap"
)

func UnsafeModifyLastRevision(lg *zap.Logger, bumpAmount uint64, be backend.Backend) error {
	defer be.ForceCommit()

	tx := be.BatchTx()
	tx.LockOutsideApply()
	defer tx.Unlock()

	latest, err := unsafeGetLatestRevision(tx)
	if err != nil {
		return err
	}

	latest = unsafeBumpRevision(lg, tx, latest, int64(bumpAmount))
	unsafeMarkRevisionCompacted(lg, tx, latest)
	return nil
}

func unsafeBumpRevision(lg *zap.Logger, tx backend.BatchTx, latest revision, amount int64) revision {
	lg.Info(
		"bumping latest revision",
		zap.Int64("latest-revision", latest.main),
		zap.Int64("bump-amount", amount),
		zap.Int64("new-latest-revision", latest.main+amount),
	)

	latest.main += amount
	latest.sub = 0
	k := make([]byte, revBytesLen)
	revToBytes(k, latest)
	tx.UnsafePut(buckets.Key, k, []byte{})

	return latest
}

func unsafeMarkRevisionCompacted(lg *zap.Logger, tx backend.BatchTx, latest revision) {
	lg.Info(
		"marking revision compacted",
		zap.Int64("revision", latest.main),
	)

	mvcc.UnsafeSetScheduledCompact(tx, latest.main)
}

func unsafeGetLatestRevision(tx backend.BatchTx) (revision, error) {
	var latest revision
	err := tx.UnsafeForEach(buckets.Key, func(k, _ []byte) (err error) {
		rev := bytesToRev(k)

		if rev.GreaterThan(latest) {
			latest = rev
		}

		return nil
	})
	return latest, err
}
