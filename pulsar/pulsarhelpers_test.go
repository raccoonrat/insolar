/*
 *    Copyright 2018 Insolar
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

package pulsar

import (
	"crypto"
	"testing"

	"github.com/insolar/insolar/core"
	"github.com/insolar/insolar/pulsar/entropygenerator"
	"github.com/insolar/insolar/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSingAndVerify(t *testing.T) {
	cs := testutils.NewCryptographyServiceMock(t)
	cs.SignFunc = func(p []byte) (r *core.Signature, r1 error) {
		signature := core.SignatureFromBytes([]byte("signature"))
		return &signature, nil
	}
	cs.VerifyFunc = func(p crypto.PublicKey, p1 core.Signature, p2 []byte) (r bool) {
		require.Equal(t, p, "publicKey")
		return true
	}

	kp := mockKeyProcessor(t)

	for i := 0; i < 20; i++ {
		testData := (&entropygenerator.StandardEntropyGenerator{}).GenerateEntropy()

		signature, err := signData(cs, testData)
		require.NoError(t, err)
		assert.ObjectsAreEqual([]byte("signature"), signature)

		// Act
		checkSignature, err := checkPayloadSignature(cs, kp, &Payload{PublicKey: "publicKey", Signature: signature, Body: testData})

		// Assert
		require.NoError(t, err)
		require.Equal(t, true, checkSignature)
	}

}
