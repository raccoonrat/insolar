/*
 *    Copyright 2019 Insolar Technologies
 *
 *    Licensed under the Apache License, Version 2.0 (the "License");
 *    you may not use this file except in compliance with the License.
 *    You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 *    Unless required by applicable law or agreed to in writing, software
 *    distributed under the License is distributed on an "AS IS" BASIS,
 *    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *    See the License for the specific language governing permissions and
 *    limitations under the License.
 */

package storage

import (
	"context"
	"fmt"
	"sync"

	"github.com/dgraph-io/badger"
	"github.com/insolar/insolar/core"
	"github.com/insolar/insolar/ledger/storage/index"
	"github.com/insolar/insolar/ledger/storage/record"
)

/*
recordStorage:
	recordsPerJet = map[jetID]recordsPerPulse = map[jetID]map[pulseNumber]recordMemory = map[jeiID]map[pulseNumber]map[objectID]recordValue

blobStorage:
	blobsPerJet = map[jetID]blobsPerPulse = map[jetID]map[pulseNumber]blobMemory = map[jeiID]map[pulseNumber]map[objectID]blobValue

indexStorage:
	indicesPerJet = map[jetID]indicesPerPulse = map[jetID]map[pulseNumber]indexMemory = map[jeiID]map[pulseNumber]map[objectID]indexValue
*/
type objectStorageMEM struct {
	PlatformCryptographyScheme core.PlatformCryptographyScheme `inject:""`

	locks    []*core.RecordID
	idlocker *IDLocker

	recordLock    sync.Mutex
	recordStorage recordsPerJet

	blobLock    sync.Mutex
	blobStorage blobsPerJet

	indexLock    sync.Mutex
	indexStorage indicesPerJet
}

func (os *objectStorageMEM) BeginTransaction(update bool) (*TransactionManager, error) {
	panic("implement me")
}

func (os *objectStorageMEM) View(ctx context.Context, fn func(*TransactionManager) error) error {
	panic("implement me")
}

func (os *objectStorageMEM) Update(ctx context.Context, fn func(*TransactionManager) error) error {
	// panic("implement me")
	return fn(os)
}

func (os *objectStorageMEM) IterateRecordsOnPulse(
	ctx context.Context,
	jetID core.RecordID,
	pulse core.PulseNumber,
	handler func(id core.RecordID, rec record.Record) error,
) error {
	panic("implement me")
}

func (os *objectStorageMEM) StoreKeyValues(ctx context.Context, kvs []core.KV) error {
	panic("implement me")
}

func (os *objectStorageMEM) GetBadgerDB() *badger.DB {
	panic("implement me")
}

func (os *objectStorageMEM) Close() error {
	panic("implement me")
}

func (os *objectStorageMEM) set(ctx context.Context, key, value []byte) error {
	panic("implement me")
}

func (os *objectStorageMEM) get(ctx context.Context, key []byte) ([]byte, error) {
	panic("implement me")
}

func (os *objectStorageMEM) waitingFlight() {
	panic("implement me")
}

func (os *objectStorageMEM) iterate(ctx context.Context,
	prefix []byte,
	handler func(k, v []byte) error,
) error {
	panic("implement me")
}

// simple aliases for keys in compound map
type jetID = core.RecordID
type objectID = core.RecordID
type pulseNumber = core.PulseNumber

// aliases for records in storage
type recordValue = record.Record
type blobValue = []byte
type indexValue = *index.ObjectLifeline

// structures for inner memory map
type recordMemory struct {
	mem map[objectID]recordValue
}

type blobMemory struct {
	mem map[objectID]blobValue
}

type indexMemory struct {
	mem map[objectID]indexValue
}

// structures for memory maps per pulses
type recordsPerPulse struct {
	mem map[pulseNumber]recordMemory
}

// type recordsPerPulse = map[pulseNumber]recordMemory

type blobsPerPulse struct {
	mem map[pulseNumber]blobMemory
}

type indicesPerPulse struct {
	mem map[pulseNumber]indexMemory
}

// structures for memory maps with pulses per jet
type recordsPerJet struct {
	mem map[jetID]recordsPerPulse
}

type blobsPerJet struct {
	mem map[jetID]blobsPerPulse
}

type indicesPerJet struct {
	mem map[jetID]indicesPerPulse
}

func NewObjectStorageMem() ObjectStorage {
	return &objectStorageMEM{
		recordStorage: newRecordsPerJet(),
		blobStorage:   newBlobsPerJet(),
		indexStorage:  newIndicesPerJet(),
		idlocker:      NewIDLocker(),
	}
}

func (os *objectStorageMEM) GetBlob(ctx context.Context, jetID core.RecordID, id *core.RecordID) ([]byte, error) {
	os.blobLock.Lock()
	defer os.blobLock.Unlock()

	blob := os.getBlobValue(jetID, id.Pulse(), *id)

	if blob == nil {
		return nil, core.ErrNotFound
	}

	return blob, nil
}

func (os *objectStorageMEM) SetBlob(ctx context.Context, jetID core.RecordID, pulseNumber core.PulseNumber, blob []byte) (*core.RecordID, error) {
	os.blobLock.Lock()
	defer os.blobLock.Unlock()

	id := record.CalculateIDForBlob(os.PlatformCryptographyScheme, pulseNumber, blob)

	return os.setBlobValue(jetID, *id, blob)
}

func (os *objectStorageMEM) GetRecord(ctx context.Context, jetID core.RecordID, id *core.RecordID) (record.Record, error) {
	os.recordLock.Lock()
	defer os.recordLock.Unlock()

	rec, err := os.getRecordValue(jetID, id.Pulse(), *id)

	if err != nil {
		return nil, err
	}

	return rec, nil
}

func (os *objectStorageMEM) SetRecord(ctx context.Context, jetID core.RecordID, pulseNumber core.PulseNumber, rec record.Record) (*core.RecordID, error) {
	os.recordLock.Lock()
	defer os.recordLock.Unlock()

	id := record.NewRecordIDFromRecord(os.PlatformCryptographyScheme, pulseNumber, rec)

	return os.setRecordValue(jetID, *id, rec)
}

func (os *objectStorageMEM) IterateIndexIDs(
	ctx context.Context,
	jetID core.RecordID,
	handler func(id core.RecordID) error,
) error {
	os.indexLock.Lock()
	defer os.indexLock.Unlock()

	return os.iterate(jetID, handler)
}

func (os *objectStorageMEM) iterate(jetID jetID, handler func(id core.RecordID) error) error {
	for _, indicesPerPulse := range os.getPulseIndices(jetID).mem {
		for id := range indicesPerPulse.mem {
			err := handler(id)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (os *objectStorageMEM) GetObjectIndex(
	ctx context.Context,
	jetID core.RecordID,
	id *core.RecordID,
	forupdate bool,
) (*index.ObjectLifeline, error) {
	if forupdate {
		os.lockOnID(id)
	}
	defer os.releaseLocks()

	os.indexLock.Lock()
	defer os.indexLock.Unlock()

	fmt.Println("=== GET ===: ")
	fmt.Println("JET_ID: ", jetID)
	fmt.Println("PULSE: ", id.Pulse())
	fmt.Println("ID: ", id)

	result, err := os.getIndexValue(jetID, id.Pulse(), *id)

	if err != nil {
		fmt.Println("OOOOOOOOOPS")
		return nil, err
	}

	return result, nil
}

func (os *objectStorageMEM) SetObjectIndex(
	ctx context.Context,
	jetID core.RecordID,
	id *core.RecordID,
	idx *index.ObjectLifeline,
) error {
	os.indexLock.Lock()
	defer os.indexLock.Unlock()

	fmt.Println("+++ SET +++: ")
	fmt.Println("JET_ID: ", jetID)
	fmt.Println("PULSE: ", id.Pulse())
	fmt.Println("ID: ", id)

	return os.setIndexValue(jetID, *id, idx)
}

func (os *objectStorageMEM) RemoveObjectIndex(
	ctx context.Context,
	jetID core.RecordID,
	ref *core.RecordID,
) error {
	os.indexLock.Lock()
	defer os.indexLock.Unlock()

	return os.removeObjectIndex(jetID, *ref)
}

// Block for records
func (os *objectStorageMEM) getJetRecords() *recordsPerJet {
	return &os.recordStorage
}

func (os *objectStorageMEM) getPulseRecords(jetID jetID) *recordsPerPulse {
	storage, ok := os.getJetRecords().mem[jetID]

	if !ok {
		storage = newRecordsPerPulse()
		os.getJetRecords().mem[jetID] = storage
	}

	return &storage
}

func (os *objectStorageMEM) getRecords(jetID jetID, pulseNumber pulseNumber) *recordMemory {
	storage, ok := os.getPulseRecords(jetID).mem[pulseNumber]

	if !ok {
		storage = newRecordMemory()
		os.getPulseRecords(jetID).mem[pulseNumber] = storage
	}

	return &storage
}

func (os *objectStorageMEM) getRecordValue(jetID jetID, pulseNumber pulseNumber, id objectID) (recordValue, error) {
	value, ok := os.getRecords(jetID, pulseNumber).mem[id]

	if !ok {
		return nil, core.ErrNotFound
	}

	return value, nil
}

func (os *objectStorageMEM) setRecordValue(jetID jetID, id objectID, rec record.Record) (*objectID, error) {
	_, err := os.getRecordValue(jetID, id.Pulse(), id)

	if err == nil {
		return nil, ErrOverride
	}
	// set value in recordMemory map
	os.getRecords(jetID, id.Pulse()).mem[id] = rec

	return &id, nil
}

func newRecordsPerJet() recordsPerJet {
	return recordsPerJet{mem: map[jetID]recordsPerPulse{}}
}

func newRecordsPerPulse() recordsPerPulse {
	return recordsPerPulse{mem: map[pulseNumber]recordMemory{}}
}

func newRecordMemory() recordMemory {
	return recordMemory{mem: map[objectID]recordValue{}}
}

// Block for blobs
func (os *objectStorageMEM) getJetBlobs() *blobsPerJet {
	return &os.blobStorage
}

func (os *objectStorageMEM) getPulseBlobs(jetID jetID) *blobsPerPulse {
	storage, ok := os.getJetBlobs().mem[jetID]

	if !ok {
		storage = newBlobsPerPulse()
		os.getJetBlobs().mem[jetID] = storage
	}

	return &storage
}

func (os *objectStorageMEM) getBlobs(jetID jetID, pulseNumber pulseNumber) *blobMemory {
	storage, ok := os.getPulseBlobs(jetID).mem[pulseNumber]

	if !ok {
		storage = newBlobMemory()
		os.getPulseBlobs(jetID).mem[pulseNumber] = storage
	}

	return &storage
}

func (os *objectStorageMEM) getBlobValue(jetID jetID, pulseNumber pulseNumber, id objectID) blobValue {
	value := os.getBlobs(jetID, pulseNumber).mem[id]

	return value
}

func (os *objectStorageMEM) setBlobValue(jetID jetID, id objectID, blob []byte) (*objectID, error) {
	// TODO: @imarkin. 28.02.19. Blob override is ok.

	// set value in blobMemory map
	os.getBlobs(jetID, id.Pulse()).mem[id] = blob

	return &id, nil
}

func newBlobsPerJet() blobsPerJet {
	return blobsPerJet{mem: map[jetID]blobsPerPulse{}}
}

func newBlobsPerPulse() blobsPerPulse {
	return blobsPerPulse{mem: map[pulseNumber]blobMemory{}}
}

func newBlobMemory() blobMemory {
	return blobMemory{mem: map[objectID]blobValue{}}
}

// Block for indices
func (os *objectStorageMEM) getJetIndices() *indicesPerJet {
	return &os.indexStorage
}

func (os *objectStorageMEM) getPulseIndices(jetID jetID) *indicesPerPulse {
	storage, ok := os.getJetIndices().mem[jetID]

	if !ok {
		storage = newIndicesPerPulse()
		os.getJetIndices().mem[jetID] = storage
	}

	return &storage
}

func (os *objectStorageMEM) getIndices(jetID jetID, pulseNumber pulseNumber) *indexMemory {
	storage, ok := os.getPulseIndices(jetID).mem[pulseNumber]

	if !ok {
		storage = newIndexMemory()
		os.getPulseIndices(jetID).mem[pulseNumber] = storage
	}

	return &storage
}

func (os *objectStorageMEM) getIndexValue(jetID jetID, pulseNumber pulseNumber, id objectID) (indexValue, error) {
	value, ok := os.getIndices(jetID, pulseNumber).mem[id]

	if !ok {
		return nil, core.ErrNotFound
	}

	return value, nil
}

func (os *objectStorageMEM) setIndexValue(jetID jetID, id objectID, idx indexValue) error {
	if idx.Delegates == nil {
		idx.Delegates = map[core.RecordRef]core.RecordRef{}
	}

	os.getIndices(jetID, id.Pulse()).mem[id] = idx

	return nil
}

func (os *objectStorageMEM) removeObjectIndex(jetID jetID, id objectID) error {
	delete(os.getIndices(jetID, id.Pulse()).mem, id)

	return nil
}

func newIndicesPerJet() indicesPerJet {
	return indicesPerJet{mem: map[jetID]indicesPerPulse{}}
}

func newIndicesPerPulse() indicesPerPulse {
	return indicesPerPulse{mem: map[pulseNumber]indexMemory{}}
}

func newIndexMemory() indexMemory {
	return indexMemory{mem: map[objectID]indexValue{}}
}

func (os *objectStorageMEM) lockOnID(id *core.RecordID) {
	os.idlocker.Lock(id)
	os.locks = append(os.locks, id)
}

func (os *objectStorageMEM) releaseLocks() {
	for _, id := range os.locks {
		os.idlocker.Unlock(id)
	}
}
