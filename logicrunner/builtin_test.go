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

package logicrunner

import (
	"context"
	"crypto"
	"testing"

	"github.com/insolar/insolar/component"
	"github.com/insolar/insolar/instrumentation/inslogger"
	"github.com/insolar/insolar/ledger/pulsemanager"
	"github.com/insolar/insolar/ledger/recentstorage"
	"github.com/insolar/insolar/logicrunner/artifacts"
	"github.com/insolar/insolar/messagebus"
	"github.com/insolar/insolar/platformpolicy"
	"github.com/insolar/insolar/testutils/network"
	"github.com/insolar/insolar/testutils/nodekeeper"
	"github.com/stretchr/testify/require"

	"github.com/insolar/insolar/insolar/delegationtoken"
	"github.com/insolar/insolar/insolar/reply"
	"github.com/insolar/insolar/testutils"
	"github.com/stretchr/testify/assert"

	"github.com/insolar/insolar/configuration"
	"github.com/insolar/insolar/insolar"
	"github.com/insolar/insolar/insolar/message"
	"github.com/insolar/insolar/logicrunner/builtin/helloworld"

	"github.com/insolar/insolar/logicrunner/goplugin/goplugintestutils"
	"github.com/insolar/insolar/testutils/testmessagebus"
)

func byteRecorRef(b byte) insolar.Reference {
	var ref insolar.Reference
	ref[insolar.RecordRefSize-1] = b
	return ref
}

func TestBareHelloworld(t *testing.T) {
	t.Skip("Not ready for async")
	ctx := context.TODO()
	lr, err := NewLogicRunner(&configuration.LogicRunner{
		BuiltIn: &configuration.BuiltIn{},
	})

	mock := testutils.NewCryptographyServiceMock(t)
	mock.SignFunc = func(p []byte) (r *insolar.Signature, r1 error) {
		signature := insolar.SignatureFromBytes(nil)
		return &signature, nil
	}
	mock.GetPublicKeyFunc = func() (r crypto.PublicKey, r1 error) {
		return nil, nil
	}
	delegationTokenFactory := delegationtoken.NewDelegationTokenFactory()
	parcelFactory := messagebus.NewParcelFactory()
	nk := nodekeeper.GetTestNodekeeper(mock)

	mb := testmessagebus.NewTestMessageBus(t)

	// FIXME: TmpLedger is deprecated. Use mocks instead.
	l, db, cleaner := artifacts.TmpLedger(
		t, "", insolar.StaticRoleLightMaterial,
		insolar.Components{
			LogicRunner: lr,
			NodeNetwork: nk,
			MessageBus:  mb,
		},
		true,
	)
	defer cleaner()

	recent := recentstorage.NewProviderMock(t)

	gil := testutils.NewGlobalInsolarLockMock(t)
	gil.AcquireMock.Return()
	gil.ReleaseMock.Return()

	l.PulseManager.(*pulsemanager.PulseManager).GIL = gil

	currentPulse, err := mb.PulseStorage.Current(ctx)
	require.NoError(t, err)

	_ = l.GetPulseManager().Set(
		ctx,
		insolar.Pulse{PulseNumber: currentPulse.PulseNumber, Entropy: insolar.Entropy{}},
		true,
	)

	nw := network.GetTestNetwork()
	scheme := platformpolicy.NewPlatformCryptographyScheme()

	cm := &component.Manager{}
	cm.Register(scheme)
	cm.Register(l.GetPulseManager(), l.GetArtifactManager(), l.GetJetCoordinator())
	cm.Inject(db, nk, recent, l, lr, nw, mb, delegationTokenFactory, parcelFactory, mock)
	err = cm.Init(ctx)
	assert.NoError(t, err)
	err = cm.Start(ctx)
	assert.NoError(t, err)

	am := l.GetArtifactManager()

	MessageBusTrivialBehavior(mb, lr)

	hw := helloworld.NewHelloWorld()

	domain := byteRecorRef(2)
	request := byteRecorRef(3)
	_, _, protoRef, err := goplugintestutils.AMPublishCode(t, am, domain, request, insolar.MachineTypeBuiltin, []byte("helloworld"))
	assert.NoError(t, err)

	contract, err := am.RegisterRequest(ctx, *am.GenesisRef(), &message.Parcel{Msg: &message.CallConstructor{PrototypeRef: byteRecorRef(4)}})
	assert.NoError(t, err)

	// TODO: use proper conversion
	reqref := insolar.Reference{}
	reqref.SetRecord(*contract)

	_, err = am.ActivateObject(
		ctx, domain, reqref, *am.GenesisRef(), *protoRef, false,
		goplugintestutils.CBORMarshal(t, hw),
	)
	assert.NoError(t, err)
	assert.Equal(t, true, contract != nil, "contract created")

	msg := &message.CallMethod{
		ObjectRef: reqref,
		Method:    "Greet",
		Arguments: goplugintestutils.CBORMarshal(t, []interface{}{"Vany"}),
	}
	parcel, err := parcelFactory.Create(ctx, msg, testutils.RandomRef(), nil, *insolar.GenesisPulse)
	assert.NoError(t, err)
	// #1
	ctx = inslogger.ContextWithTrace(ctx, "TestBareHelloworld1")
	resp, err := lr.Execute(
		ctx,
		parcel,
	)
	assert.NoError(t, err, "contract call")

	r := goplugintestutils.CBORUnMarshal(t, resp.(*reply.CallMethod).Result)
	assert.Equal(t, []interface{}([]interface{}{"Hello Vany's world"}), r)

	msg = &message.CallMethod{
		ObjectRef: reqref,
		Method:    "Greet",
		Arguments: goplugintestutils.CBORMarshal(t, []interface{}{"Ruz"}),
	}
	parcel, err = parcelFactory.Create(ctx, msg, testutils.RandomRef(), nil, *insolar.GenesisPulse)
	assert.NoError(t, err)
	// #2
	ctx = inslogger.ContextWithTrace(ctx, "TestBareHelloworld2")
	resp, err = lr.Execute(
		ctx,
		parcel,
	)
	assert.NoError(t, err, "contract call")

	r = goplugintestutils.CBORUnMarshal(t, resp.(*reply.CallMethod).Result)
	assert.Equal(t, []interface{}([]interface{}{"Hello Ruz's world"}), r)
}
