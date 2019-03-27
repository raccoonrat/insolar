//
// Copyright 2019 Insolar Technologies GmbH
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package storage_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/insolar/insolar/component"
	"github.com/insolar/insolar/insolar"
	"github.com/insolar/insolar/instrumentation/inslogger"
	"github.com/insolar/insolar/ledger/recentstorage"
	"github.com/insolar/insolar/ledger/storage"
	"github.com/insolar/insolar/ledger/storage/db"
	"github.com/insolar/insolar/ledger/storage/drop"
	"github.com/insolar/insolar/ledger/storage/storagetest"
	"github.com/insolar/insolar/platformpolicy"
	"github.com/insolar/insolar/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type cleanerSuite struct {
	suite.Suite

	cm      *component.Manager
	ctx     context.Context
	cleaner func()

	objectStorage  storage.ObjectStorage
	dropModifier   drop.Modifier
	dropAccessor   drop.Accessor
	storageCleaner storage.Cleaner
}

func NewCleanerSuite() *cleanerSuite {
	return &cleanerSuite{
		Suite: suite.Suite{},
	}
}

// Init and run suite
func TestCleaner(t *testing.T) {
	suite.Run(t, NewCleanerSuite())
}

func (s *cleanerSuite) BeforeTest(suiteName, testName string) {
	s.cm = &component.Manager{}
	s.ctx = inslogger.TestContext(s.T())

	tmpDB, cleaner := storagetest.TmpDB(s.ctx, s.T())
	s.cleaner = cleaner

	s.objectStorage = storage.NewObjectStorage()
	s.storageCleaner = storage.NewCleaner()

	storageDB := db.NewDBWithBadger(tmpDB.GetBadgerDB())
	dropStorage := drop.NewStorageDB(storageDB)
	s.dropAccessor = dropStorage
	s.dropModifier = dropStorage

	s.cm.Inject(
		platformpolicy.NewPlatformCryptographyScheme(),
		tmpDB,
		db.NewMemoryMockDB(),
		s.objectStorage,
		s.storageCleaner,
		s.dropAccessor,
		s.dropModifier,
	)

	err := s.cm.Init(s.ctx)
	if err != nil {
		s.T().Error("ComponentManager init failed", err)
	}
	err = s.cm.Start(s.ctx)
	if err != nil {
		s.T().Error("ComponentManager start failed", err)
	}
}

func (s *cleanerSuite) AfterTest(suiteName, testName string) {
	err := s.cm.Stop(s.ctx)
	if err != nil {
		s.T().Error("ComponentManager stop failed", err)
	}
	s.cleaner()
}

func (s *cleanerSuite) Test_RemoveRecords() {
	t := s.T()
	ctx := inslogger.TestContext(t)

	jetID00 := testutils.JetFromString("00")
	jetID01 := testutils.JetFromString("01")
	jetID11 := testutils.JetFromString("11")
	jets := []insolar.ID{jetID00, jetID01, jetID11}

	// should remove all records in rmJetID on pulses 1, 2, but all in pulse 3 for rmJetID should left
	// and other jets records should not be removed too
	var checks []cleanChecker
	until := 2
	rmUntilPN := insolar.PulseNumber(insolar.FirstPulseNumber + until + 1)
	rmJetID := jetID01

	for _, jetID := range jets {
		for i := 1; i <= 3; i++ {
			pn := insolar.PulseNumber(insolar.FirstPulseNumber + i)

			shouldLeft := true
			if jetID == rmJetID {
				shouldLeft = i > until
			}

			blobID, err := storagetest.AddRandBlob(ctx, s.objectStorage, jetID, pn)
			require.NoError(t, err)
			blobCC := cleanCase{
				rectype:    "blob",
				id:         blobID,
				jetID:      jetID,
				pulseNum:   pn,
				shouldLeft: shouldLeft,
			}
			checks = append(checks, blobCase{
				cleanCase:     blobCC,
				objectStorage: s.objectStorage,
			})

			recID, err := storagetest.AddRandRecord(ctx, s.objectStorage, jetID, pn)
			require.NoError(t, err)
			recCC := cleanCase{
				rectype:    "record",
				id:         recID,
				jetID:      jetID,
				pulseNum:   pn,
				shouldLeft: shouldLeft,
			}
			checks = append(checks, recordCase{
				cleanCase:     recCC,
				objectStorage: s.objectStorage,
			})
		}
	}

	s.storageCleaner.CleanJetRecordsUntilPulse(ctx, rmJetID, rmUntilPN)

	for _, check := range checks {
		check.Check(ctx, t)
	}
}

func (s *cleanerSuite) Test_RemoveJetIndexes() {
	t := s.T()
	ctx := inslogger.TestContext(t)

	jetID00 := testutils.JetFromString("00")
	jetID01 := testutils.JetFromString("01")
	jetID11 := testutils.JetFromString("11")
	jets := []insolar.ID{jetID00, jetID01, jetID11}

	// should remove records in Pulse 1, 2, but left 3
	var checks []cleanChecker
	until := 2
	rmJetID := jetID01
	var removeIndexes []insolar.ID

	for _, jetID := range jets {
		for i := 1; i <= 3; i++ {
			pn := insolar.PulseNumber(insolar.FirstPulseNumber + i)
			idxID, err := storagetest.AddRandIndex(ctx, s.objectStorage, jetID, pn)
			require.NoError(t, err)

			shouldLeft := true
			if jetID == rmJetID {
				shouldLeft = i > until
				if !shouldLeft {
					removeIndexes = append(removeIndexes, *idxID)
				}
			}

			cc := cleanCase{
				id:         idxID,
				jetID:      jetID,
				pulseNum:   pn,
				shouldLeft: shouldLeft,
			}
			checks = append(checks, indexCase{
				cleanCase:     cc,
				objectStorage: s.objectStorage,
			})
		}
	}

	recent := recentstorage.NewRecentIndexStorageMock(s.T())
	recent.FilterNotExistWithLockFunc = func(ctx context.Context, candidates []insolar.ID, fn func(fordelete []insolar.ID)) {
		fn(candidates)
	}

	s.storageCleaner.CleanJetIndexes(ctx, rmJetID, recent, removeIndexes)

	for _, check := range checks {
		check.Check(ctx, t)
	}
}

// check helpers

type cleanChecker interface {
	Check(ctx context.Context, t *testing.T)
	String() string
}

type cleanCase struct {
	rectype    string
	id         *insolar.ID
	jetID      insolar.ID
	pulseNum   insolar.PulseNumber
	shouldLeft bool
}

func (cc cleanCase) String() string {
	return fmt.Sprintf("%v jetID=%v, pulseNum=%v, shouldLeft=%v",
		cc.rectype, cc.jetID.DebugString(), cc.pulseNum, cc.shouldLeft)
}

func (cc cleanCase) check(t *testing.T, err error) {
	if cc.shouldLeft {
		if !assert.NoError(t, err) {
			fmt.Printf("%v => err: %T\n", cc, err)
		}
		return
	}
	if !assert.Exactly(t, err, insolar.ErrNotFound) {
		fmt.Printf("%v => err: %T\n", cc, err)
	}
}

type indexCase struct {
	cleanCase
	objectStorage storage.ObjectStorage
}

func (c indexCase) Check(ctx context.Context, t *testing.T) {
	_, err := c.objectStorage.GetObjectIndex(ctx, c.jetID, c.id)
	c.check(t, err)
}

type blobCase struct {
	cleanCase
	objectStorage storage.ObjectStorage
}

func (c blobCase) Check(ctx context.Context, t *testing.T) {
	_, err := c.objectStorage.GetBlob(ctx, c.jetID, c.id)
	c.check(t, err)
}

type recordCase struct {
	cleanCase
	objectStorage storage.ObjectStorage
}

func (c recordCase) Check(ctx context.Context, t *testing.T) {
	_, err := c.objectStorage.GetRecord(ctx, c.jetID, c.id)
	c.check(t, err)
}
