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

package storagetest

import (
	"context"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/insolar/insolar/component"
	"github.com/insolar/insolar/configuration"
	"github.com/insolar/insolar/insolar/jet"
	"github.com/insolar/insolar/ledger/storage"
	"github.com/insolar/insolar/ledger/storage/db"
	"github.com/insolar/insolar/ledger/storage/drop"
	"github.com/insolar/insolar/ledger/storage/genesis"
	"github.com/insolar/insolar/testutils"
)

type tmpDBOptions struct {
	dir         string
	nobootstrap bool
}

// Option provides functional option for TmpDB.
type Option func(*tmpDBOptions)

// Dir defines temporary directory for database.
func Dir(dir string) Option {
	return func(opts *tmpDBOptions) {
		opts.dir = dir
	}
}

// DisableBootstrap skip bootstrap records creation.
func DisableBootstrap() Option {
	return func(opts *tmpDBOptions) {
		opts.nobootstrap = true
	}
}

// TmpDB returns BadgerDB's storage implementation and cleanup function.
//
// Creates BadgerDB in temporary directory and uses t for errors reporting.
func TmpDB(ctx context.Context, t testing.TB, options ...Option) (storage.DBContext, func()) {
	opts := &tmpDBOptions{}
	for _, o := range options {
		o(opts)
	}
	tmpdir, err := ioutil.TempDir(opts.dir, "bdb-test-")
	assert.NoError(t, err)

	tmpDB, err := storage.NewDB(configuration.Ledger{
		Storage: configuration.Storage{
			DataDirectory: tmpdir,
		},
	}, nil)
	require.NoError(t, err)

	cm := &component.Manager{}

	storageDB := db.NewDBWithBadger(tmpDB.GetBadgerDB())
	ds := drop.NewStorageDB(storageDB)

	cm.Inject(
		testutils.NewPlatformCryptographyScheme(),
		tmpDB,
		jet.NewStore(),
		db.NewMemoryMockDB(),
		storage.NewObjectStorage(),
		ds,
		storage.NewPulseTracker(),
	)

	if !opts.nobootstrap {
		gi := genesis.NewGenesisInitializer()
		cm.Inject(gi)
	}

	err = cm.Init(ctx)
	if err != nil {
		t.Error("ComponentManager init failed", err)
	}
	err = cm.Start(ctx)
	if err != nil {
		t.Error("ComponentManager start failed", err)
	}

	return tmpDB, func() {
		rmErr := os.RemoveAll(tmpdir)
		if rmErr != nil {
			t.Fatal("temporary tmpDB dir cleanup failed", rmErr)
		}
	}
}
