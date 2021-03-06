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

package pulsemanager

import (
	"github.com/insolar/insolar/insolar"
)

// It's a internal interface of pulse manager, you shouldn't use outside of pm
//go:generate minimock -i github.com/insolar/insolar/ledger/pulsemanager.pulseStoragePm -o ./ -s _mock.go
type pulseStoragePm interface {
	insolar.PulseStorage

	Set(pulse *insolar.Pulse)

	Lock()
	Unlock()
}
