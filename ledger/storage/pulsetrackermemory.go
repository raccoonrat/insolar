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

package storage

import (
	"context"
	"sync"

	"github.com/insolar/insolar/insolar"
	"github.com/insolar/insolar/instrumentation/inslogger"
	"go.opencensus.io/stats"
)

type pulseTrackerMemory struct {
	mutex       sync.RWMutex
	memory      map[insolar.PulseNumber]Pulse
	latestPulse insolar.PulseNumber
}

// NewPulseTracker returns new instance PulseTracker with in-memory realization
// DEPRECATED
func NewPulseTrackerMemory() PulseTracker {
	return &pulseTrackerMemory{memory: make(map[insolar.PulseNumber]Pulse)}
}

// GetPulse returns pulse for provided pulse number.
func (p *pulseTrackerMemory) GetPulse(ctx context.Context, num insolar.PulseNumber) (*Pulse, error) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	return p.getPulse(ctx, num)
}

// GetPreviousPulse returns pulse for provided pulse number.
func (p *pulseTrackerMemory) GetPreviousPulse(ctx context.Context, num insolar.PulseNumber) (*Pulse, error) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	return p.getNthPrevPulse(ctx, 1, num)
}

// GetNthPrevPulse returns Nth previous pulse from some pulse number
func (p *pulseTrackerMemory) GetNthPrevPulse(ctx context.Context, n uint, from insolar.PulseNumber) (*Pulse, error) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	return p.getNthPrevPulse(ctx, n, from)
}

// Deprecated: use insolar.PulseStorage.Current() instead (or private getLatestPulse if applicable).
func (p *pulseTrackerMemory) GetLatestPulse(ctx context.Context) (*Pulse, error) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	return p.getLatestPulse(ctx)
}

// AddPulse saves new pulse data and latestPulse index.
func (p *pulseTrackerMemory) AddPulse(ctx context.Context, pulse insolar.Pulse) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	pn := pulse.PulseNumber

	if pn <= p.latestPulse {
		return ErrBadPulse
	}

	var (
		previousPulseNumber  insolar.PulseNumber
		previousSerialNumber int
	)

	previousPulse, err := p.getLatestPulse(ctx)
	if err != nil && err != insolar.ErrNotFound {
		return err
	}

	// Set next on previousPulseNumber pulse if it exists.
	if err == nil {
		if previousPulse != nil {
			previousPulseNumber = previousPulse.Pulse.PulseNumber
			previousSerialNumber = previousPulse.SerialNumber
		}

		prevPulse, err := p.getPulse(ctx, previousPulseNumber)
		if err != nil {
			return err
		}
		prevPulse.Next = &pulse.PulseNumber
		p.memory[prevPulse.Pulse.PulseNumber] = *prevPulse
	}

	// Save new pulse.
	newPulse := Pulse{
		Prev:         &previousPulseNumber,
		SerialNumber: previousSerialNumber + 1,
		Pulse:        pulse,
	}

	p.memory[pn] = newPulse
	p.latestPulse = pn

	stats.Record(ctx,
		statPulseAdded.M(1),
	)

	return nil
}

// DeletePulse delete pulse data.
func (p *pulseTrackerMemory) DeletePulse(ctx context.Context, num insolar.PulseNumber) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	_, err := p.getPulse(ctx, num)

	if err == insolar.ErrNotFound {
		inslogger.FromContext(ctx).Error("can't delete non-existing pulse")
		return nil
	}

	if err != nil {
		return err
	}

	delete(p.memory, num)

	stats.Record(ctx,
		statPulseDeleted.M(1),
	)

	return nil
}

func (p *pulseTrackerMemory) getPulse(ctx context.Context, num insolar.PulseNumber) (*Pulse, error) {
	// TODO: @imarkin 14.02.18 - it's a hack for fill genesis pulse in memory realization
	if num == insolar.FirstPulseNumber {
		pulse := &Pulse{
			Pulse: insolar.Pulse{
				PulseNumber: insolar.FirstPulseNumber,
				Entropy:     insolar.GenesisPulse.Entropy,
			},
			SerialNumber: 1,
		}
		return pulse, nil
	}

	pulse, ok := p.memory[num]

	if !ok {
		return nil, insolar.ErrNotFound
	}

	return &pulse, nil
}

func (p *pulseTrackerMemory) getNthPrevPulse(ctx context.Context, n uint, num insolar.PulseNumber) (*Pulse, error) {
	pulse, err := p.getPulse(ctx, num)
	if err != nil {
		return nil, err
	}

	for n > 0 {
		if pulse.Prev == nil {
			return nil, ErrPrevPulse
		}

		pulse, err = p.getPulse(ctx, *pulse.Prev)

		if err != nil {
			return nil, insolar.ErrNotFound
		}
		n--
	}
	return pulse, nil
}

func (p *pulseTrackerMemory) getLatestPulse(ctx context.Context) (*Pulse, error) {
	if p.latestPulse == 0 {
		return nil, insolar.ErrNotFound
	}

	return p.getPulse(ctx, p.latestPulse)
}
