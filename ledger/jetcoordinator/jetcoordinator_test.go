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

package jetcoordinator

import (
	"context"
	"testing"

	"github.com/insolar/insolar/component"
	"github.com/insolar/insolar/insolar"
	"github.com/insolar/insolar/insolar/jet"
	"github.com/insolar/insolar/instrumentation/inslogger"
	"github.com/insolar/insolar/ledger/storage"
	"github.com/insolar/insolar/ledger/storage/node"
	"github.com/insolar/insolar/ledger/storage/storagetest"
	"github.com/insolar/insolar/platformpolicy"
	"github.com/insolar/insolar/pulsar/entropygenerator"
	"github.com/insolar/insolar/testutils"
	"github.com/insolar/insolar/testutils/network"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type jetCoordinatorSuite struct {
	suite.Suite

	cm      *component.Manager
	ctx     context.Context
	cleaner func()

	pulseStorage *storage.PulseStorage
	pulseTracker storage.PulseTracker
	jetStorage   jet.Storage
	nodeStorage  *node.AccessorMock
	coordinator  *JetCoordinator
}

func NewJetCoordinatorSuite() *jetCoordinatorSuite {
	return &jetCoordinatorSuite{
		Suite: suite.Suite{},
	}
}

func TestCoordinator(t *testing.T) {
	suite.Run(t, NewJetCoordinatorSuite())
}

func (s *jetCoordinatorSuite) BeforeTest(suiteName, testName string) {
	s.cm = &component.Manager{}
	s.ctx = inslogger.TestContext(s.T())

	db, cleaner := storagetest.TmpDB(s.ctx, s.T())

	s.cleaner = cleaner
	s.pulseTracker = storage.NewPulseTracker()
	s.pulseStorage = storage.NewPulseStorage()
	storage := jet.NewStore()
	s.jetStorage = storage
	s.nodeStorage = node.NewAccessorMock(s.T())
	s.coordinator = NewJetCoordinator(5)
	s.coordinator.NodeNet = network.NewNodeNetworkMock(s.T())

	s.cm.Inject(
		testutils.NewPlatformCryptographyScheme(),
		db,
		s.pulseTracker,
		s.pulseStorage,
		storage,
		s.nodeStorage,
		s.coordinator,
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

func (s *jetCoordinatorSuite) AfterTest(suiteName, testName string) {
	err := s.cm.Stop(s.ctx)
	if err != nil {
		s.T().Error("ComponentManager stop failed", err)
	}
	s.cleaner()
}

func (s *jetCoordinatorSuite) TestJetCoordinator_QueryRole() {
	err := s.pulseTracker.AddPulse(s.ctx, insolar.Pulse{PulseNumber: 0, Entropy: insolar.Entropy{1, 2, 3}})
	require.NoError(s.T(), err)
	var nds []insolar.Node
	var nodeRefs []insolar.Reference
	for i := 0; i < 100; i++ {
		ref := *insolar.NewReference(insolar.DomainID, *insolar.NewID(0, []byte{byte(i)}))
		nds = append(nds, insolar.Node{ID: ref, Role: insolar.StaticRoleLightMaterial})
		nodeRefs = append(nodeRefs, ref)
	}
	require.NoError(s.T(), err)

	s.nodeStorage.InRoleMock.Return(nds, nil)

	objID := insolar.NewID(0, []byte{1, 42, 123})
	s.jetStorage.Update(s.ctx, 0, true, *insolar.NewJetID(50, []byte{1, 42, 123}))

	selected, err := s.coordinator.QueryRole(s.ctx, insolar.DynamicRoleLightValidator, *objID, 0)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 3, len(selected))

	// Indexes are hard-coded from previously calculated values.
	assert.Equal(s.T(), []insolar.Reference{nodeRefs[16], nodeRefs[21], nodeRefs[78]}, selected)
}

func TestJetCoordinator_Me(t *testing.T) {
	t.Parallel()
	// Arrange
	expectedID := testutils.RandomRef()
	nodeNet := network.NewNodeNetworkMock(t)
	node := network.NewNetworkNodeMock(t)
	nodeNet.GetOriginMock.Return(node)
	node.IDMock.Return(expectedID)
	jc := NewJetCoordinator(1)
	jc.NodeNet = nodeNet

	// Act
	resultID := jc.Me()

	// Assert
	require.Equal(t, expectedID, resultID)
}

func TestNewJetCoordinator(t *testing.T) {
	t.Parallel()
	// Act
	calc := NewJetCoordinator(12)

	// Assert
	require.NotNil(t, calc)
	require.Equal(t, 12, calc.lightChainLimit)
}

func TestJetCoordinator_IsBeyondLimit_ProblemsWithTracker(t *testing.T) {
	t.Parallel()
	// Arrange
	ctx := inslogger.TestContext(t)
	pulseTrackerMock := storage.NewPulseTrackerMock(t)
	pulseTrackerMock.GetPulseMock.Return(nil, errors.New("it's expected"))
	calc := NewJetCoordinator(12)
	calc.PulseTracker = pulseTrackerMock

	// Act
	res, err := calc.IsBeyondLimit(ctx, insolar.FirstPulseNumber, 0)

	// Assert
	require.NotNil(t, err)
	require.Equal(t, false, res)
}

func TestJetCoordinator_IsBeyondLimit_ProblemsWithTracker_SecondCall(t *testing.T) {
	t.Parallel()
	// Arrange
	ctx := inslogger.TestContext(t)
	pulseTrackerMock := storage.NewPulseTrackerMock(t)
	pulseTrackerMock.GetPulseFunc = func(p context.Context, p1 insolar.PulseNumber) (r *storage.Pulse, r1 error) {
		if p1 == insolar.FirstPulseNumber {
			return &storage.Pulse{}, nil
		}

		return nil, errors.New("it's expected")
	}
	calc := NewJetCoordinator(12)
	calc.PulseTracker = pulseTrackerMock

	// Act
	res, err := calc.IsBeyondLimit(ctx, insolar.FirstPulseNumber, 0)

	// Assert
	require.NotNil(t, err)
	require.Equal(t, false, res)
}

func TestJetCoordinator_IsBeyondLimit_OutsideOfLightChainLimit(t *testing.T) {
	t.Parallel()
	// Arrange
	ctx := inslogger.TestContext(t)
	pulseTrackerMock := storage.NewPulseTrackerMock(t)
	pulseTrackerMock.GetPulseFunc = func(p context.Context, p1 insolar.PulseNumber) (r *storage.Pulse, r1 error) {
		if p1 == insolar.FirstPulseNumber {
			return &storage.Pulse{SerialNumber: 50}, nil
		}

		return &storage.Pulse{SerialNumber: 24}, nil
	}
	calc := NewJetCoordinator(25)
	calc.PulseTracker = pulseTrackerMock

	// Act
	res, err := calc.IsBeyondLimit(ctx, insolar.FirstPulseNumber, 0)

	// Assert
	require.Nil(t, err)
	require.Equal(t, true, res)
}

func TestJetCoordinator_IsBeyondLimit_InsideOfLightChainLimit(t *testing.T) {
	t.Parallel()
	// Arrange
	ctx := inslogger.TestContext(t)
	pulseTrackerMock := storage.NewPulseTrackerMock(t)
	pulseTrackerMock.GetPulseFunc = func(p context.Context, p1 insolar.PulseNumber) (r *storage.Pulse, r1 error) {
		if p1 == insolar.FirstPulseNumber {
			return &storage.Pulse{SerialNumber: 50}, nil
		}

		return &storage.Pulse{SerialNumber: 34}, nil
	}
	calc := NewJetCoordinator(25)
	calc.PulseTracker = pulseTrackerMock

	// Act
	res, err := calc.IsBeyondLimit(ctx, insolar.FirstPulseNumber, 0)

	// Assert
	require.Nil(t, err)
	require.Equal(t, false, res)
}

func TestJetCoordinator_NodeForJet_CheckLimitFailed(t *testing.T) {
	t.Parallel()
	// Arrange
	ctx := inslogger.TestContext(t)
	pulseTrackerMock := storage.NewPulseTrackerMock(t)
	pulseTrackerMock.GetPulseMock.Return(nil, errors.New("it's expected"))
	calc := NewJetCoordinator(12)
	calc.PulseTracker = pulseTrackerMock

	// Act
	res, err := calc.NodeForJet(ctx, testutils.RandomJet(), insolar.FirstPulseNumber, 0)

	// Assert
	require.NotNil(t, err)
	require.Nil(t, res)
}

func TestJetCoordinator_NodeForJet_GoToHeavy(t *testing.T) {
	t.Parallel()
	// Arrange
	ctx := inslogger.TestContext(t)
	pulseTrackerMock := storage.NewPulseTrackerMock(t)
	pulseTrackerMock.GetPulseFunc = func(p context.Context, p1 insolar.PulseNumber) (r *storage.Pulse, r1 error) {
		if p1 == insolar.FirstPulseNumber {
			return &storage.Pulse{SerialNumber: 50}, nil
		}

		return &storage.Pulse{SerialNumber: 24}, nil
	}
	expectedID := insolar.NewReference(testutils.RandomID(), testutils.RandomID())
	activeNodesStorageMock := node.NewAccessorMock(t)
	activeNodesStorageMock.InRoleFunc = func(p insolar.PulseNumber, p1 insolar.StaticRole) (r []insolar.Node, r1 error) {
		require.Equal(t, insolar.FirstPulseNumber, int(p))
		require.Equal(t, insolar.StaticRoleHeavyMaterial, p1)

		return []insolar.Node{{ID: *expectedID}}, nil
	}

	pulseStorageMock := testutils.NewPulseStorageMock(t)
	pulseStorageMock.CurrentFunc = func(p context.Context) (r *insolar.Pulse, r1 error) {
		generator := entropygenerator.StandardEntropyGenerator{}
		return &insolar.Pulse{PulseNumber: insolar.FirstPulseNumber, Entropy: generator.GenerateEntropy()}, nil
	}

	calc := NewJetCoordinator(25)
	calc.PulseTracker = pulseTrackerMock
	calc.Nodes = activeNodesStorageMock
	calc.PulseStorage = pulseStorageMock
	calc.PlatformCryptographyScheme = platformpolicy.NewPlatformCryptographyScheme()

	// Act
	resNode, err := calc.NodeForJet(ctx, testutils.RandomJet(), insolar.FirstPulseNumber, 0)

	// Assert
	require.Nil(t, err)
	require.Equal(t, expectedID, resNode)
}

func TestJetCoordinator_NodeForJet_GoToLight(t *testing.T) {
	t.Parallel()
	// Arrange
	ctx := inslogger.TestContext(t)
	pulseTrackerMock := storage.NewPulseTrackerMock(t)
	pulseTrackerMock.GetPulseFunc = func(p context.Context, p1 insolar.PulseNumber) (r *storage.Pulse, r1 error) {
		if p1 == insolar.FirstPulseNumber {
			return &storage.Pulse{SerialNumber: 50}, nil
		}

		return &storage.Pulse{SerialNumber: 49}, nil
	}
	expectedID := insolar.NewReference(testutils.RandomID(), testutils.RandomID())
	activeNodesStorageMock := node.NewAccessorMock(t)
	activeNodesStorageMock.InRoleFunc = func(p insolar.PulseNumber, p1 insolar.StaticRole) (r []insolar.Node, r1 error) {
		require.Equal(t, 0, int(p))
		require.Equal(t, insolar.StaticRoleLightMaterial, p1)

		return []insolar.Node{{ID: *expectedID}}, nil
	}

	pulseStorageMock := testutils.NewPulseStorageMock(t)
	pulseStorageMock.CurrentFunc = func(p context.Context) (r *insolar.Pulse, r1 error) {
		generator := entropygenerator.StandardEntropyGenerator{}
		return &insolar.Pulse{PulseNumber: insolar.FirstPulseNumber, Entropy: generator.GenerateEntropy()}, nil
	}

	calc := NewJetCoordinator(25)
	calc.PulseTracker = pulseTrackerMock
	calc.Nodes = activeNodesStorageMock
	calc.PulseStorage = pulseStorageMock
	calc.PlatformCryptographyScheme = platformpolicy.NewPlatformCryptographyScheme()

	// Act
	resNode, err := calc.NodeForJet(ctx, testutils.RandomJet(), insolar.FirstPulseNumber, 0)

	// Assert
	require.Nil(t, err)
	require.Equal(t, expectedID, resNode)
}
