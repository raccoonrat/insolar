/*
 *    Copyright 2019 Insolar
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

package jet

import (
	"bytes"
	"context"
	"sync"

	"github.com/insolar/insolar/ledger/storage"
	"github.com/ugorji/go/codec"

	"github.com/insolar/insolar/core"
)

// JetStorage provides methods for working with jets
//go:generate minimock -i github.com/insolar/insolar/ledger/storage/jet.JetStorage -o ./ -s _mock.go
type JetStorage interface {
	UpdateJetTree(ctx context.Context, pulse core.PulseNumber, setActual bool, ids ...core.RecordID)
	FindJet(ctx context.Context, pulse core.PulseNumber, id core.RecordID) (*core.RecordID, bool)
	SplitJetTree(ctx context.Context, pulse core.PulseNumber, jetID core.RecordID) (*core.RecordID, *core.RecordID, error)
	CloneJetTree(ctx context.Context, from, to core.PulseNumber) *Tree
	DeleteJetTree(ctx context.Context, pulse core.PulseNumber)

	AddJets(ctx context.Context, jetIDs ...core.RecordID) error
	GetJets(ctx context.Context) (IDSet, error)
}

type jetStorage struct {
	DB storage.DBContext `inject:""`

	trees     map[core.PulseNumber]*Tree
	treesLock sync.RWMutex

	addJetLock sync.RWMutex
}

func NewJetStorage() JetStorage {
	return &jetStorage{
		trees: map[core.PulseNumber]*Tree{},
	}
}

// FindJet finds jet for specified pulse and object.
func (js *jetStorage) FindJet(ctx context.Context, pulse core.PulseNumber, id core.RecordID) (*core.RecordID, bool) {
	js.treesLock.RLock()

	if t, ok := js.trees[pulse]; ok {
		defer js.treesLock.RUnlock()
		return t.Find(id)
	}
	js.treesLock.RUnlock()

	js.treesLock.Lock()
	defer js.treesLock.Unlock()
	return js.getJetTree(ctx, pulse).Find(id)
}

// UpdateJetTree updates jet tree for specified pulse.
func (js *jetStorage) UpdateJetTree(ctx context.Context, pulse core.PulseNumber, setActual bool, ids ...core.RecordID) {
	js.treesLock.Lock()
	defer js.treesLock.Unlock()

	tree := js.getJetTree(ctx, pulse)
	for _, id := range ids {
		tree.Update(id, setActual)
	}
}

// SplitJetTree performs jet split and returns resulting jet ids.
func (js *jetStorage) SplitJetTree(
	ctx context.Context, pulse core.PulseNumber, jetID core.RecordID,
) (*core.RecordID, *core.RecordID, error) {
	js.treesLock.Lock()
	defer js.treesLock.Unlock()

	tree := js.getJetTree(ctx, pulse)

	left, right, err := tree.Split(jetID)
	if err != nil {
		return nil, nil, err
	}

	return left, right, nil
}

// CloneJetTree copies tree from one pulse to another. Use it to copy past tree into new pulse.
func (js *jetStorage) CloneJetTree(
	ctx context.Context, from, to core.PulseNumber,
) *Tree {
	js.treesLock.Lock()
	defer js.treesLock.Unlock()

	tree := js.getJetTree(ctx, from)

	res := tree.Clone(false)
	js.trees[to] = res
	return res
}

func (js *jetStorage) DeleteJetTree(
	ctx context.Context, pulse core.PulseNumber,
) {
	js.treesLock.Lock()
	defer js.treesLock.Unlock()

	delete(js.trees, pulse)
}

func (js *jetStorage) getJetTree(ctx context.Context, pulse core.PulseNumber) *Tree {
	if t, ok := js.trees[pulse]; ok {
		return t
	}

	tree := NewTree(pulse == core.GenesisPulse.PulseNumber)
	js.trees[pulse] = tree
	return tree
}

// AddJets stores a list of jets of the current node.
func (js *jetStorage) AddJets(ctx context.Context, jetIDs ...core.RecordID) error {
	js.addJetLock.Lock()
	defer js.addJetLock.Unlock()

	k := storage.JetListPrefixKey()

	var jets IDSet
	buff, err := js.DB.Get(ctx, k)
	if err == nil {
		dec := codec.NewDecoder(bytes.NewReader(buff), &codec.CborHandle{})
		err = dec.Decode(&jets)
		if err != nil {
			return err
		}
	} else if err == core.ErrNotFound {
		jets = IDSet{}
	} else {
		return err
	}

	for _, id := range jetIDs {
		jets[id] = struct{}{}
	}
	return js.DB.Set(ctx, k, jets.Bytes())
}

// GetJets returns jets of the current node
func (js *jetStorage) GetJets(ctx context.Context) (IDSet, error) {
	js.addJetLock.RLock()
	defer js.addJetLock.RUnlock()

	k := storage.JetListPrefixKey()
	buff, err := js.DB.Get(ctx, k)
	if err != nil {
		return nil, err
	}

	dec := codec.NewDecoder(bytes.NewReader(buff), &codec.CborHandle{})
	var jets IDSet
	err = dec.Decode(&jets)
	if err != nil {
		return nil, err
	}

	return jets, nil
}