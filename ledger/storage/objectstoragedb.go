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
	"github.com/insolar/insolar/core/message"
	"github.com/insolar/insolar/ledger/storage/index"
	"github.com/insolar/insolar/ledger/storage/jet"
	"github.com/insolar/insolar/ledger/storage/record"
)

type objectStorageDB struct {
	DB                         DBContext                       `inject:""`
	PlatformCryptographyScheme core.PlatformCryptographyScheme `inject:""`
}

func NewObjectStorageDB() ObjectStorage {
	return new(objectStorageDB)
}

// GetBlob returns binary value stored by record ID.
// TODO: switch from reference to passing blob id for consistency - @nordicdyno 6.Dec.2018
func (os *objectStorageDB) GetBlob(ctx context.Context, jetID core.RecordID, id *core.RecordID) ([]byte, error) {
	var (
		blob []byte
		err  error
	)

	err = os.DB.View(ctx, func(tx *TransactionManager) error {
		blob, err = tx.GetBlob(ctx, jetID, id)
		return err
	})
	if err != nil {
		return nil, err
	}
	return blob, nil
}

// SetBlob saves binary value for provided pulse.
func (os *objectStorageDB) SetBlob(ctx context.Context, jetID core.RecordID, pulseNumber core.PulseNumber, blob []byte) (*core.RecordID, error) {
	var (
		id  *core.RecordID
		err error
	)
	err = os.DB.Update(ctx, func(tx *TransactionManager) error {
		id, err = tx.SetBlob(ctx, jetID, pulseNumber, blob)
		return err
	})
	if err != nil {
		return nil, err
	}
	return id, nil
}

// GetRecord wraps matching transaction manager method.
func (os *objectStorageDB) GetRecord(ctx context.Context, jetID core.RecordID, id *core.RecordID) (record.Record, error) {
	var (
		fetchedRecord record.Record
		err           error
	)

	err = os.DB.View(ctx, func(tx *TransactionManager) error {
		fetchedRecord, err = tx.GetRecord(ctx, jetID, id)
		return err
	})
	if err != nil {
		return nil, err
	}
	return fetchedRecord, nil
}

// SetRecord wraps matching transaction manager method.
func (os *objectStorageDB) SetRecord(ctx context.Context, jetID core.RecordID, pulseNumber core.PulseNumber, rec record.Record) (*core.RecordID, error) {
	var (
		id  *core.RecordID
		err error
	)
	err = os.DB.Update(ctx, func(tx *TransactionManager) error {
		id, err = tx.SetRecord(ctx, jetID, pulseNumber, rec)
		return err
	})
	if err != nil {
		return nil, err
	}
	return id, nil
}

// SetMessage persists message to the database
func (os *objectStorageDB) SetMessage(ctx context.Context, jetID core.RecordID, pulseNumber core.PulseNumber, genericMessage core.Message) error {
	_, prefix := jet.Jet(jetID)
	messageBytes := message.ToBytes(genericMessage)
	hw := os.PlatformCryptographyScheme.ReferenceHasher()
	_, err := hw.Write(messageBytes)
	if err != nil {
		return err
	}
	hw.Sum(nil)

	return os.DB.set(
		ctx,
		prefixkey(scopeIDMessage, prefix, pulseNumber.Bytes(), hw.Sum(nil)),
		messageBytes,
	)
}

// IterateIndexIDs iterates over index IDs on provided Jet ID.
func (os *objectStorageDB) IterateIndexIDs(
	ctx context.Context,
	jetID core.RecordID,
	handler func(id core.RecordID) error,
) error {
	_, jetPrefix := jet.Jet(jetID)
	prefix := prefixkey(scopeIDLifeline, jetPrefix)

	return os.DB.iterate(ctx, prefix, func(k, v []byte) error {
		pn := pulseNumFromKey(0, k)
		id := core.NewRecordID(pn, k[core.PulseNumberSize:])
		err := handler(*id)
		if err != nil {
			return err
		}
		return nil
	})
}

// GetObjectIndex wraps matching transaction manager method.
func (os *objectStorageDB) GetObjectIndex(
	ctx context.Context,
	jetID core.RecordID,
	id *core.RecordID,
	forupdate bool,
) (*index.ObjectLifeline, error) {
	tx, err := os.DB.BeginTransaction(false)
	if err != nil {
		return nil, err
	}
	defer tx.Discard()

	idx, err := tx.GetObjectIndex(ctx, jetID, id, forupdate)
	if err != nil {
		return nil, err
	}
	return idx, nil
}

// SetObjectIndex wraps matching transaction manager method.
func (os *objectStorageDB) SetObjectIndex(
	ctx context.Context,
	jetID core.RecordID,
	id *core.RecordID,
	idx *index.ObjectLifeline,
) error {
	return os.DB.Update(ctx, func(tx *TransactionManager) error {
		return tx.SetObjectIndex(ctx, jetID, id, idx)
	})
}

// RemoveObjectIndex removes an index of an object
func (os *objectStorageDB) RemoveObjectIndex(
	ctx context.Context,
	jetID core.RecordID,
	ref *core.RecordID,
) error {
	return os.DB.Update(ctx, func(tx *TransactionManager) error {
		return tx.RemoveObjectIndex(ctx, jetID, ref)
	})
}
