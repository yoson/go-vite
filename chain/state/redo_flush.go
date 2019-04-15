package chain_state

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/vitelabs/go-vite/chain/utils"
	"github.com/vitelabs/go-vite/common/types"
)

func (redo *Redo) Id() types.Hash {
	return redo.id
}

func (redo *Redo) Prepare() {
	redo.flushingBatchMap = make(map[uint64]*FlushingBatch, len(redo.snapshotLogMap))

	for snapshotHeight, snapshotLog := range redo.snapshotLogMap {
		if snapshotHeight == redo.currentSnapshotHeight &&
			snapshotLog.FlushOpt == optWrite {
			continue
		}

		flushingBatch := &FlushingBatch{
			Operation: snapshotLog.FlushOpt,
		}
		switch snapshotLog.FlushOpt {
		case optWrite:
			batch := new(leveldb.Batch)
			for addr, redoLogList := range snapshotLog.RedoLogMap {
				var valueBuffer bytes.Buffer
				enc := gob.NewEncoder(&valueBuffer)

				for _, redoLog := range redoLogList {
					err := enc.Encode(redoLog)
					if err != nil {
						panic(fmt.Sprintf("enc.Encode: %+v. Error: %s", redoLog, err.Error()))
					}

				}
				batch.Put(addr.Bytes(), valueBuffer.Bytes())
			}
			flushingBatch.Batch = batch
		case optRollback:
		case optCover:
		}

		redo.flushingBatchMap[snapshotHeight] = flushingBatch
	}

}

func (redo *Redo) CancelPrepare() {
	redo.flushingBatchMap = nil
}

func (redo *Redo) RedoLog() ([]byte, error) {
	redoLogSize := 0
	for _, flushingBatch := range redo.flushingBatchMap {

		redoLogSize += 9
		if flushingBatch.Batch != nil {
			redoLogSize += 4 + len(flushingBatch.Batch.Dump())
		}

	}

	redoLog := make([]byte, 0, redoLogSize)

	for height, flushingBatch := range redo.flushingBatchMap {
		redoLog = append(redoLog, chain_utils.Uint64ToBytes(height)...)
		redoLog = append(redoLog, flushingBatch.Operation)

		switch flushingBatch.Operation {
		case optWrite:
			batchLen := len(flushingBatch.Batch.Dump())
			batchLenBytes := make([]byte, 4)
			binary.BigEndian.PutUint32(batchLenBytes, uint32(batchLen))

			redoLog = append(redoLog, batchLenBytes...)
			redoLog = append(redoLog, flushingBatch.Batch.Dump()...)
		case optRollback:
		case optCover:

		}
	}

	return redoLog, nil
}

// assume commit immediately after delete
func (redo *Redo) Commit() error {
	defer func() {
		// clear flushing batch
		redo.flushingBatchMap = nil

		// keep current
		current := redo.snapshotLogMap[redo.currentSnapshotHeight]

		redo.snapshotLogMap = make(map[uint64]*SnapshotLog)

		current.FlushOpt = optWrite

		redo.snapshotLogMap[redo.currentSnapshotHeight] = current

	}()

	tx, err := redo.store.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for snapshotHeight, flushingBatch := range redo.flushingBatchMap {
		switch flushingBatch.Operation {
		case optWrite:
			if err := redo.flush(tx, snapshotHeight, flushingBatch.Batch); err != nil {
				return err
			}
		case optRollback:
			fallthrough
		case optCover:
			tx.DeleteBucket(chain_utils.Uint64ToBytes(snapshotHeight))
		}

	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (redo *Redo) PatchRedoLog(redoLog []byte) error {
	currentPointer := 0
	endPointer := len(redoLog) - 1
	status := 0
	size := 8

	var snapshotHeight uint64
	var operation byte
	var batchSize uint32
	var batch *leveldb.Batch

	tx, err := redo.store.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for currentPointer < endPointer {
		buff := redoLog[currentPointer : currentPointer+size]
		currentPointer += size
		switch status {
		case 0:
			if snapshotHeight > 0 {
				switch operation {
				case optWrite:
					if err := redo.flush(tx, snapshotHeight, batch); err != nil {
						return err
					}
				case optRollback:
					fallthrough
				case optCover:
					tx.DeleteBucket(chain_utils.Uint64ToBytes(snapshotHeight))
				}
			}

			snapshotHeight = binary.BigEndian.Uint64(buff)
			size = 1
		case 1:
			operation = buff[0]
			size = 4
		case 2:
			batchSize = binary.BigEndian.Uint32(buff)
			size = int(batchSize)
		case 3:
			batch = new(leveldb.Batch)
			if err := batch.Load(buff); err != nil {
				return err
			}

			size = 8
		}
		status = (status + 1) % 4

	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil

}
func (redo *Redo) flush(tx *bolt.Tx, snapshotHeight uint64, batch *leveldb.Batch) error {
	bu, err := tx.CreateBucketIfNotExists(chain_utils.Uint64ToBytes(snapshotHeight))
	if err != nil {
		return err
	}

	// add
	batch.Replay(NewBatchFlush(bu))

	// delete
	if snapshotHeight > redo.retainHeight {
		// ignore error
		tx.DeleteBucket(chain_utils.Uint64ToBytes(snapshotHeight - redo.retainHeight))
	}
	return nil
}

type BatchFlush struct {
	bu *bolt.Bucket
}

func NewBatchFlush(bu *bolt.Bucket) *BatchFlush {
	return &BatchFlush{
		bu: bu,
	}
}

func (flush *BatchFlush) Put(key []byte, value []byte) {
	if err := flush.bu.Put(key, value); err != nil {
		panic(err)
	}
}

func (flush *BatchFlush) Delete(key []byte) {
	if err := flush.bu.Delete(key); err != nil {
		panic(err)
	}
}