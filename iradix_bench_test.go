package iradix

import (
	"testing"

	"github.com/google/uuid"
)

func benchBulkInsertTxn(b *testing.B, batchSize int, track bool) {
	r := New()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		txn := r.Txn()

		txn.TrackMutate(track)

		keys := make([][]byte, batchSize)
		values := make([]interface{}, batchSize)
		for j := 0; j < batchSize; j++ {
			keys[j] = []byte(uuid.New().String())
			values[j] = j
		}
		txn.BulkInsert(keys, values)

		r = txn.Commit()
	}
}

func Benchmark10BulkInsertTxnTrack(b *testing.B) {
	benchBulkInsertTxn(b, 10, true)
}
func Benchmark10BulkInsertTxnNoTrack(b *testing.B) {
	benchBulkInsertTxn(b, 10, false)
}

func Benchmark100BulkInsertTxnTrack(b *testing.B) {
	benchBulkInsertTxn(b, 100, true)
}
func Benchmark100BulkInsertTxnNoTrack(b *testing.B) {
	benchBulkInsertTxn(b, 100, false)
}

func Benchmark1000BulkInsertTxnTrack(b *testing.B) {
	benchBulkInsertTxn(b, 1000, true)
}
func Benchmark1000BulkInsertTxnNoTrack(b *testing.B) {
	benchBulkInsertTxn(b, 1000, false)
}

func Benchmark10000BulkInsertTxnTrack(b *testing.B) {
	benchBulkInsertTxn(b, 10000, true)
}
func Benchmark10000BulkInsertTxnNoTrack(b *testing.B) {
	benchBulkInsertTxn(b, 10000, false)
}

func Benchmark100000BulkInsertTxnTrack(b *testing.B) {
	benchBulkInsertTxn(b, 100000, true)
}
func Benchmark100000BulkInsertTxnNoTrack(b *testing.B) {
	benchBulkInsertTxn(b, 100000, false)
}
