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
	"sync"

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

	recordLock    sync.Mutex
	recordStorage recordsPerJet

	blobLock    sync.Mutex
	blobStorage blobsPerJet

	indexStorage indicesPerJet
}

// simple aliases for keys in compound map
type jetID = core.RecordID
type objectID = core.RecordID
type pulseNumber = core.PulseNumber

// aliases for records in storage
type recordValue = record.Record
type blobValue = []byte
type indexValue = index.ObjectLifeline

// structures for inner memory map
type recordMemory struct {
	mem map[objectID]recordValue
}

type blobMemory struct {
	rwLock sync.RWMutex
	mem    map[objectID]blobValue
}

type indexMemory struct {
	rwLock sync.RWMutex
	mem    map[objectID]indexValue
}

// structures for memory maps per pulses
type recordsPerPulse struct {
	mem map[pulseNumber]recordMemory
}

// type recordsPerPulse = map[pulseNumber]recordMemory

type blobsPerPulse struct {
	rwLock sync.RWMutex
	mem    map[pulseNumber]blobMemory
}

type indicesPerPulse struct {
	rwLock sync.RWMutex
	mem    map[pulseNumber]indexMemory
}

// structures for memory maps with pulses per jet
type recordsPerJet struct {
	mem map[jetID]recordsPerPulse
}

type blobsPerJet struct {
	rwLock sync.RWMutex
	mem    map[jetID]blobsPerPulse
}

type indicesPerJet struct {
	rwLock sync.RWMutex
	mem    map[jetID]indicesPerPulse
}

func NewObjectStorageMem() ObjectStorage {
	return &objectStorageMEM{
		recordStorage: newRecordsPerJet(),
		blobStorage:   newBlobsPerJet(),
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

	rec := os.getRecordValue(jetID, id.Pulse(), *id)

	if rec == nil {
		return nil, core.ErrNotFound
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
	panic("implement me")
}

func (os *objectStorageMEM) GetObjectIndex(
	ctx context.Context,
	jetID core.RecordID,
	id *core.RecordID,
	forupdate bool,
) (*index.ObjectLifeline, error) {
	panic("implement me")
}

func (os *objectStorageMEM) SetObjectIndex(
	ctx context.Context,
	jetID core.RecordID,
	id *core.RecordID,
	idx *index.ObjectLifeline,
) error {
	panic("implement me")
}

func (os *objectStorageMEM) RemoveObjectIndex(
	ctx context.Context,
	jetID core.RecordID,
	ref *core.RecordID,
) error {
	panic("implement me")
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

func (os *objectStorageMEM) getRecordValue(jetID jetID, pulseNumber pulseNumber, id objectID) recordValue {
	value := os.getRecords(jetID, pulseNumber).mem[id]

	return value
}

func (os *objectStorageMEM) setRecordValue(jetID jetID, id objectID, rec record.Record) (*objectID, error) {
	current := os.getRecordValue(jetID, id.Pulse(), id)
	if current != nil {
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
