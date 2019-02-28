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
	"testing"

	"github.com/insolar/insolar/core"
	"github.com/insolar/insolar/instrumentation/inslogger"
	"github.com/insolar/insolar/ledger/storage/record"
	"github.com/insolar/insolar/testutils"
	"github.com/stretchr/testify/assert"
)

func TestObjectStorageMEM_SetRecord(t *testing.T) {
	t.Parallel()

	// Arrange
	ctx := inslogger.TestContext(t)
	jetID := testutils.RandomJet()
	objectStorage := &objectStorageMEM{
		recordStorage: newRecordsPerJet(),
	}
	scheme := testutils.NewPlatformCryptographyScheme()
	objectStorage.PlatformCryptographyScheme = scheme

	rec := &record.RequestRecord{}

	// Act and Assertions
	gotRef, err := objectStorage.SetRecord(ctx, jetID, core.GenesisPulse.PulseNumber, rec)
	assert.Nil(t, err)

	gotRec, err := objectStorage.GetRecord(ctx, jetID, gotRef)
	assert.Nil(t, err)
	assert.Equal(t, rec, gotRec)

	_, err = objectStorage.SetRecord(ctx, jetID, core.GenesisPulse.PulseNumber, rec)
	assert.Equal(t, ErrOverride, err)
}

func TestObjectStorageMEM_SetBlob(t *testing.T) {
	t.Parallel()

	// Arrange
	ctx := inslogger.TestContext(t)
	jetID := testutils.RandomJet()
	objectStorage := &objectStorageMEM{
		blobStorage: newBlobsPerJet(),
	}
	scheme := testutils.NewPlatformCryptographyScheme()
	objectStorage.PlatformCryptographyScheme = scheme

	blob := []byte("100500")

	// Act and Assertions
	gotRef, err := objectStorage.SetBlob(ctx, jetID, core.GenesisPulse.PulseNumber, blob)
	assert.Nil(t, err)

	gotBlob, err := objectStorage.GetBlob(ctx, jetID, gotRef)
	assert.Nil(t, err)
	assert.Equal(t, blob, gotBlob)

	_, err = objectStorage.SetBlob(ctx, jetID, core.GenesisPulse.PulseNumber, blob)
	assert.NoError(t, err)
}
