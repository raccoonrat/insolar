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

	"github.com/insolar/insolar/core"
	"github.com/insolar/insolar/ledger/storage/index"
	"github.com/insolar/insolar/ledger/storage/record"
)

type objectStorageMEM struct {
	// TODO implement this
}

func NewObjectStorageMem() ObjectStorage {
	panic("implement me")
	// return new(objectStorageMEM)
}

func (os *objectStorageMEM) GetBlob(ctx context.Context, jetID core.RecordID, id *core.RecordID) ([]byte, error) {
	panic("implement me")
}

func (os *objectStorageMEM) SetBlob(ctx context.Context, jetID core.RecordID, pulseNumber core.PulseNumber, blob []byte) (*core.RecordID, error) {
	panic("implement me")
}

func (os *objectStorageMEM) GetRecord(ctx context.Context, jetID core.RecordID, id *core.RecordID) (record.Record, error) {
	panic("implement me")
}

func (os *objectStorageMEM) SetRecord(ctx context.Context, jetID core.RecordID, pulseNumber core.PulseNumber, rec record.Record) (*core.RecordID, error) {
	panic("implement me")
}

func (os *objectStorageMEM) SetMessage(ctx context.Context, jetID core.RecordID, pulseNumber core.PulseNumber, genericMessage core.Message) error {
	panic("implement me")
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
