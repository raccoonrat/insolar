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

package blob

import (
	"testing"

	"github.com/insolar/insolar/insolar/gen"
	"github.com/stretchr/testify/assert"
)

func TestClone(t *testing.T) {
	t.Parallel()

	jetID := gen.JetID()
	rawBlob := slice()
	blob := Blob{
		JetID: jetID,
		Value: rawBlob,
	}

	clonedBlob := Clone(blob)

	assert.Equal(t, blob, clonedBlob)
	assert.False(t, &blob == &clonedBlob)
}
