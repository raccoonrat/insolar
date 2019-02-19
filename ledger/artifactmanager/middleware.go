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

package artifactmanager

import (
	"context"

	"github.com/pkg/errors"

	"github.com/insolar/insolar/configuration"
	"github.com/insolar/insolar/core"
	"github.com/insolar/insolar/core/message"
	"github.com/insolar/insolar/core/reply"
	"github.com/insolar/insolar/instrumentation/inslogger"
	"github.com/insolar/insolar/ledger/storage"
	"github.com/insolar/insolar/ledger/storage/jet"
)

type middleware struct {
	objectStorage  storage.ObjectStorage
	jetStorage     storage.JetStorage
	jetCoordinator core.JetCoordinator
	messageBus     core.MessageBus
	pulseStorage   core.PulseStorage
	hotDataWaiter  HotDataWaiter
	conf           *configuration.Ledger
	handler        *MessageHandler
}

func newMiddleware(
	h *MessageHandler,
) *middleware {
	return &middleware{
		objectStorage:  h.ObjectStorage,
		jetStorage:     h.JetStorage,
		jetCoordinator: h.JetCoordinator,
		messageBus:     h.Bus,
		pulseStorage:   h.PulseStorage,
		hotDataWaiter:  h.HotDataWaiter,
		handler:        h,
		conf:           h.conf,
	}
}

func (m *middleware) addFieldsToLogger(handler core.MessageHandler) core.MessageHandler {
	return func(ctx context.Context, parcel core.Parcel) (core.Reply, error) {
		cntext, _ := inslogger.WithField(ctx, "targetid", parcel.DefaultTarget().String())

		return handler(cntext, parcel)
	}
}

type jetKey struct{}

func contextWithJet(ctx context.Context, jetID core.RecordID) context.Context {
	return context.WithValue(ctx, jetKey{}, jetID)
}

func jetFromContext(ctx context.Context) core.RecordID {
	val := ctx.Value(jetKey{})
	j, ok := val.(core.RecordID)
	if !ok {
		panic("failed to extract jet from context")
	}

	return j
}

func (m *middleware) zeroJetForHeavy(handler core.MessageHandler) core.MessageHandler {
	return func(ctx context.Context, parcel core.Parcel) (core.Reply, error) {
		return handler(contextWithJet(ctx, *jet.NewID(0, nil)), parcel)
	}
}

func addJetIDToLogger(ctx context.Context, jetID core.RecordID) context.Context {
	ctx, _ = inslogger.WithField(ctx, "jetid", jetID.DebugString())

	return ctx
}

func (m *middleware) checkJet(handler core.MessageHandler) core.MessageHandler {
	return func(ctx context.Context, parcel core.Parcel) (core.Reply, error) {
		msg := parcel.Message()
		if msg.DefaultTarget() == nil {
			return nil, errors.New("unexpected message")
		}

		// FIXME: @andreyromancev. 17.01.19. Temporary allow any genesis request. Remove it.
		if parcel.Pulse() == core.FirstPulseNumber {
			return handler(contextWithJet(ctx, *jet.NewID(0, nil)), parcel)
		}

		// Check token jet.
		token := parcel.DelegationToken()
		if token != nil {
			// Calculate jet for target pulse.
			target := *msg.DefaultTarget().Record()
			pulse := target.Pulse()
			switch tm := msg.(type) {
			case *message.GetObject:
				pulse = tm.State.Pulse()
			case *message.GetChildren:
				if tm.FromChild == nil {
					return nil, errors.New("fetching children without child pointer is forbidden")
				}
				pulse = tm.FromChild.Pulse()
			case *message.GetRequest:
				pulse = tm.Request.Pulse()
			}
			jetID, actual := m.jetStorage.FindJet(ctx, pulse, target)
			if !actual {
				inslogger.FromContext(ctx).WithFields(map[string]interface{}{
					"msg":   msg.Type().String(),
					"jet":   jetID.DebugString(),
					"pulse": pulse,
				}).Error("jet is not actual")
			}

			return handler(contextWithJet(ctx, *jetID), parcel)
		}

		// Calculate jet for current pulse.
		var jetID core.RecordID
		if msg.DefaultTarget().Record().Pulse() == core.PulseNumberJet {
			jetID = *msg.DefaultTarget().Record()
		} else {
			j, err := m.handler.jetTreeUpdater.fetchJet(ctx, *msg.DefaultTarget().Record(), parcel.Pulse())
			if err != nil {
				return nil, errors.Wrap(err, "failed to fetch jet tree")
			}

			jetID = *j
		}

		// Check if jet is ours.
		node, err := m.jetCoordinator.LightExecutorForJet(ctx, jetID, parcel.Pulse())
		if err != nil {
			return nil, errors.Wrap(err, "failed to calculate executor for jet")
		}

		if *node != m.jetCoordinator.Me() {
			return &reply.JetMiss{JetID: jetID}, nil
		}

		ctx = addJetIDToLogger(ctx, jetID)

		return handler(contextWithJet(ctx, jetID), parcel)
	}
}

func (m *middleware) checkHeavySync(handler core.MessageHandler) core.MessageHandler {
	return func(ctx context.Context, parcel core.Parcel) (core.Reply, error) {
		// TODO: @andreyromancev. 10.01.2019. Uncomment to enable backpressure for writing requests.
		// Currently disabled due to big initial difference in pulse numbers, which prevents requests from being accepted.
		// jetID := jetFromContext(ctx)
		// replicated, err := m.db.GetReplicatedPulse(ctx, jetID)
		// if err != nil {
		// 	return nil, err
		// }
		// if parcel.Pulse()-replicated >= m.conf.LightChainLimit {
		// 	return nil, errors.New("failed to write data (waiting for heavy replication)")
		// }

		return handler(ctx, parcel)
	}
}

func (m *middleware) waitForHotData(handler core.MessageHandler) core.MessageHandler {
	return func(ctx context.Context, parcel core.Parcel) (core.Reply, error) {
		// TODO: 15.01.2019 @egorikas
		// Hack is needed for genesis
		if parcel.Pulse() == core.FirstPulseNumber {
			return handler(ctx, parcel)
		}

		// If the call is a call in redirect-chain
		// skip waiting for the hot records
		if parcel.DelegationToken() != nil {
			return handler(ctx, parcel)
		}

		jetID := jetFromContext(ctx)
		err := m.hotDataWaiter.Wait(ctx, jetID)
		if err != nil {
			return &reply.Error{ErrType: reply.ErrHotDataTimeout}, nil
		}
		return handler(ctx, parcel)
	}
}

func (m *middleware) releaseHotDataWaiters(handler core.MessageHandler) core.MessageHandler {
	return func(ctx context.Context, parcel core.Parcel) (core.Reply, error) {
		rep, err := handler(ctx, parcel)

		hotDataMessage := parcel.Message().(*message.HotData)
		jetID := hotDataMessage.Jet.Record()
		unlockErr := m.hotDataWaiter.Unlock(ctx, *jetID)
		if unlockErr != nil {
			inslogger.FromContext(ctx).Error(err)
		}

		return rep, err
	}
}
